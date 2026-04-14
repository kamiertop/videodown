import {createMemo, createSignal} from "solid-js";
import type {NavigateOptions} from "@tanstack/solid-router";
import {model} from "../../../../wailsjs/go/models";
import {Info, SeasonsArchivesList, SeasonsSeriesList, SeriesList, VideoList} from "../../../../wailsjs/go/api/BiliBili";
import type {SidebarListItem} from "../../../components/SidebarList";
import {parseBilibiliLengthToSeconds} from "../../../lib/format";
import type {UpTab} from "../../../components/bilibili/up/UpDetailBody";
import type {MediaCardItem} from "../../../lib/model.ts";

type VideoListResp = {
    list?: {
        vlist?: Array<{
            aid: number;
            bvid: string;
            title: string;
            pic: string;
            length: string;
            created: number;
            play?: number;
            meta?: { stat?: { danmaku?: number } };
        }>
    };
    page: { pn: number; ps: number; count: number };
};

type SeriesArchivesResp = {
    page: { num: number; size: number; total: number };
    archives?: Array<{
        aid: number;
        bvid: string;
        title: string;
        pic: string;
        duration: number;
        pubdate: number;
        stat?: any;
    }>;
};

function normalizeBiliCover(url?: string): string {
    const u = (url ?? '').trim();
    if (!u) return '';
    if (u.startsWith('//')) return `https:${u}`;
    return u;
}

type ListKind = 'season' | 'series';

export interface ListsSidebarItem extends SidebarListItem {
    kind: ListKind;
    raw: any;
    cover?: string;
}

/**
 * UP 详情页逻辑层（状态 + 请求 + 映射 + 选择/下载 + toast）。
 * 路由组件只负责渲染与把 signals/actions 传给子组件。
 *
 * 数据流概览：
 * - `init()`：并发触发 `loadInfo(mid)` + `loadVideoList(false)`
 * - `Info(mid)` 的返回会回填已加载卡片的 `upperName`，避免“列表先返回导致昵称显示未知”的竞态
 * - 视频列表/合集列表的分页请求都用 `reqSeq`（自增序号）丢弃旧响应，避免快速切换时旧请求覆盖新状态
 *
 * 状态划分：
 * - Header：`loading/info`
 * - Tab：`activeTab`
 * - Videos：`videoLoading/videoError/videoCards/videoTotal/...`
 * - Lists：`ssLoading/ssError/listSidebarItems/selectedListItem/listCards/...`
 * - Selection：`selectedMediaIds/selectedSet/allSelected`
 * - Toast：`errorText/statusText/statusTone`
 */
export function createUpDetailLogic(
    getMid: () => string,
    navigate: (opts: NavigateOptions) => void,
    showToast: (message: string, type?: "error" | "success" | "info" | "warning") => void,
) {
    const [loading, setLoading] = createSignal<boolean>(true);
    const [info, setInfo] = createSignal<model.UserInfoData | null>(null);
    // 用序号丢弃旧响应：避免路由参数变更/快速切换时旧请求回写新页面状态。
    let infoReqSeq = 0;

    // ── Tab（只存“当前选中视图”） ──
    const [activeTab, setActiveTab] = createSignal<UpTab>('videos');

    // ── 全部投稿视频（列表 + 分页） ──
    const VIDEO_PAGE_SIZE = 40;
    const [videoLoading, setVideoLoading] = createSignal(false);
    const [videoError, setVideoError] = createSignal('');
    const [videoCards, setVideoCards] = createSignal<MediaCardItem[]>([]);
    const [videoPage, setVideoPage] = createSignal(1);
    const [videoTotal, setVideoTotal] = createSignal(0);
    const [videoLoadingMore, setVideoLoadingMore] = createSignal(false);
    let videoReqSeq = 0;

    const hasMoreVideos = createMemo(() => {
        const total = videoTotal();
        const loaded = videoCards().length;
        return total > 0 && loaded < total;
    });

    const currentUpperName = () => info()?.name || '未知';

    const mapVlistToCards = (vlist: VideoListResp["list"] extends { vlist?: infer V } ? V : any): MediaCardItem[] => {
        // 注意：视频列表接口可能先于 Info(mid) 返回；不要在“请求开始时”就把 up 名称拍扁到卡片里。
        // 这里使用当前的名称快照；并在 Info() 成功后统一回填卡片，保证名称最终一致。
        const upperName = currentUpperName();
        return (vlist ?? []).map((v: any) => ({
            id: Number(v.aid) || 0,
            title: v.title ?? '',
            cover: normalizeBiliCover(v.pic),
            duration: parseBilibiliLengthToSeconds(v.length ?? ''),
            bvid: v.bvid ?? '',
            upperName,
            play: v.play,
            danmaku: v.meta?.stat?.danmaku,
            pubtime: v.created,
        }));
    };

    const [selectedMediaIds, setSelectedMediaIds] = createSignal<number[]>([]);
    const selectedSet = createMemo(() => new Set(selectedMediaIds()));
    const currentCards = () => activeTab() === 'videos' ? videoCards() : listCards();

    const allSelected = () => {
        const cards = currentCards();
        if (cards.length === 0) return false;
        const s = selectedSet();
        return cards.every(c => s.has(c.id));
    };

    const toggleSelectMedia = (id: number) => {
        setSelectedMediaIds(prev => {
            const set = new Set(prev);
            if (set.has(id)) set.delete(id); else set.add(id);
            return Array.from(set);
        });
    };

    const toggleSelectAllMedia = () => {
        const cards = currentCards();
        setSelectedMediaIds(allSelected() ? [] : cards.map(c => c.id));
    };
    const clearSelection = () => setSelectedMediaIds([]);

    const downloadMediaList = (medias: MediaCardItem[], label: string) => {
        const bvids = [...new Set(medias.map(m => m.bvid?.trim()).filter(Boolean))];
        if (bvids.length === 0) {
            showToast('没有可下载的视频（缺少 BV 号）', 'warning');
            return;
        }
        navigate({
            to: '/bilibili/download',
            search: {bvids: bvids.join(',')},
        });
        showToast(`${label}，已打开下载页`, 'success');
    };

    const downloadSelectedMedia = async () => {
        const s = selectedSet();
        const medias = currentCards().filter(m => s.has(m.id));
        await downloadMediaList(medias, `已选择 ${medias.length} 个视频`);
    };

    const downloadAllMedia = () => downloadMediaList(currentCards(), '全部视频');
    const downloadOneMedia = (media: MediaCardItem) => downloadMediaList([media], `视频 ${media.title}`);

    const loadVideoList = async (append = false) => {
        const mid = getMid();
        const targetPage = append ? videoPage() + 1 : 1;
        if (append) setVideoLoadingMore(true);
        else {
            setVideoLoading(true);
            setVideoError('');
            setVideoCards([]);
            setVideoPage(1);
            setVideoTotal(0);
            setSelectedMediaIds([]);
        }

        const seq = ++videoReqSeq;
        try {
            const data = await VideoList(Number(mid), VIDEO_PAGE_SIZE, targetPage) as VideoListResp;
            if (seq !== videoReqSeq) return;
            const cards = mapVlistToCards(data.list?.vlist);
            const total = Number(data.page.count) || 0;
            setVideoTotal(total);
            if (append) setVideoCards(prev => [...prev, ...cards]);
            else setVideoCards(cards);
            setVideoPage(targetPage);
        } catch (error) {
            if (seq !== videoReqSeq) return;
            const msg = error instanceof Error ? error.message : String(error);
            if (append) showToast(`加载更多失败: ${msg}`, 'warning');
            else setVideoError(msg);
        } finally {
            if (seq === videoReqSeq) {
                if (append) setVideoLoadingMore(false);
                else setVideoLoading(false);
            }
        }
    };

    // ── 合集|系列（左侧 sidebar + 右侧详情列表） ──
    const SS_PAGE_SIZE = 20;
    const [ssLoading, setSsLoading] = createSignal(false);
    const [ssError, setSsError] = createSignal('');
    const [seasons, setSeasons] = createSignal<model.SeasonsItem[]>([]);
    const [series, setSeries] = createSignal<model.SeriesItem[]>([]);
    const [ssLoadedOnce, setSsLoadedOnce] = createSignal(false);

    const [selectedListItem, setSelectedListItem] = createSignal<ListsSidebarItem | null>(null);
    const [listDetailLoading, setListDetailLoading] = createSignal(false);
    const [listDetailError, setListDetailError] = createSignal('');
    const [listCards, setListCards] = createSignal<MediaCardItem[]>([]);
    const [listPage, setListPage] = createSignal(1);
    const [listTotal, setListTotal] = createSignal(0);
    const [listLoadingMore, setListLoadingMore] = createSignal(false);
    let listReqSeq = 0;

    const listSidebarItems = createMemo((): ListsSidebarItem[] => {
        const sItems = seasons().map((s): ListsSidebarItem => ({
            id: s.meta?.season_id ?? 0,
            title: s.meta?.title ?? s.meta?.name ?? '未命名合集',
            count: s.meta?.total ?? 0,
            subtitle: '合集',
            kind: 'season',
            raw: s,
            cover: s.meta?.cover,
        }));
        const rItems = series().map((s): ListsSidebarItem => ({
            id: s.meta?.series_id ?? 0,
            title: s.meta?.name ?? '未命名系列',
            count: s.meta?.total ?? 0,
            subtitle: '系列',
            kind: 'series',
            raw: s,
            cover: s.meta?.cover,
        }));
        return [...sItems, ...rItems].filter(i => i.id > 0);
    });

    const hasMoreListVideos = createMemo(() => {
        const total = listTotal();
        const loaded = listCards().length;
        return total > 0 && loaded < total;
    });

    const mapArchivesToCards = (archives: any[] | undefined, upperName: string): MediaCardItem[] => {
        return (archives ?? []).map((a: any) => ({
            id: Number(a.aid) || 0,
            title: a.title ?? '',
            cover: normalizeBiliCover(a.pic),
            duration: Number(a.duration) || 0,
            bvid: a.bvid ?? '',
            upperName,
            play: a.stat?.view,
            danmaku: a.stat?.danmaku,
            pubtime: a.pubdate ?? a.pubDate,
        }));
    };

    const loadListDetail = async (item: ListsSidebarItem, append: boolean) => {
        const mid = getMid();
        const targetPage = append ? listPage() + 1 : 1;
        setSelectedListItem(item);
        if (append) setListLoadingMore(true);
        else {
            setListDetailLoading(true);
            setListDetailError('');
            setListCards([]);
            setListPage(1);
            setListTotal(item.count ?? 0);
            setSelectedMediaIds([]);
        }

        const seq = ++listReqSeq;
        try {
            const upperName = currentUpperName();
            if (item.kind === 'season') {
                const data = await SeasonsArchivesList(mid, 20, targetPage, item.id) as unknown as model.SeasonsArchivesData;
                if (seq !== listReqSeq) return;
                const cards = mapArchivesToCards(data.archives, upperName);
                const total = Number(data.meta?.total ?? data.page?.Total ?? item.count ?? 0) || 0;
                setListTotal(total);
                if (append) setListCards(prev => [...prev, ...cards]); else setListCards(cards);
                setListPage(targetPage);
            } else {
                const data = await SeriesList(mid, 20, targetPage, item.id) as SeriesArchivesResp;
                if (seq !== listReqSeq) return;
                const cards = mapArchivesToCards(data.archives as any[], upperName);
                const total = Number((data as any).page?.total ?? (data as any).page?.Total ?? item.count ?? 0) || 0;
                setListTotal(total);
                if (append) setListCards(prev => [...prev, ...cards]); else setListCards(cards);
                setListPage(targetPage);
            }
        } catch (error) {
            if (seq !== listReqSeq) return;
            const msg = error instanceof Error ? error.message : String(error);
            if (append) showToast(`加载更多失败: ${msg}`, 'warning');
            else setListDetailError(msg);
        } finally {
            if (seq === listReqSeq) {
                if (append) setListLoadingMore(false);
                else setListDetailLoading(false);
            }
        }
    };

    const handleSelectListItem = (item: ListsSidebarItem) => {
        if (selectedListItem()?.kind === item.kind && selectedListItem()?.id === item.id && listCards().length > 0 && !listDetailLoading()) return;
        void loadListDetail(item, false);
    };

    const handleLoadMoreList = () => {
        const item = selectedListItem();
        if (!item || !hasMoreListVideos() || listLoadingMore()) return;
        void loadListDetail(item, true);
    };
    const retryListDetail = () => {
        const it = selectedListItem();
        if (!it) return;
        void loadListDetail(it, false);
    };

    const loadSeasonsSeriesAll = async () => {
        const mid = getMid();
        setSsLoading(true);
        setSsError('');
        setSeasons([]);
        setSeries([]);
        setSelectedListItem(null);
        setListCards([]);
        setListPage(1);
        setListTotal(0);
        setSelectedMediaIds([]);

        try {
            const allSeasons: model.SeasonsItem[] = [];
            const allSeries: model.SeriesItem[] = [];

            let pageNum = 1;
            let total = 0;
            let fetched = 0;
            const maxPages = 50;

            while (pageNum <= maxPages) {
                const data = await SeasonsSeriesList(mid, SS_PAGE_SIZE, pageNum);
                const list = data.items_lists;
                const seasonsList = list?.seasons_list ?? [];
                const seriesList = list?.series_list ?? [];
                allSeasons.push(...seasonsList);
                allSeries.push(...seriesList);

                total = list?.page?.total ?? total;
                fetched = allSeasons.length + allSeries.length;
                if (total > 0 && fetched >= total) break;

                // 如果接口没给 total，就以“本页为空”作为终止条件
                if (seasonsList.length === 0 && seriesList.length === 0) break;
                pageNum += 1;
            }

            setSeasons(allSeasons);
            setSeries(allSeries);
            setSsLoadedOnce(true);

            const first = [...allSeasons.map(s => ({
                id: s.meta?.season_id ?? 0,
                title: s.meta?.title ?? s.meta?.name ?? '未命名合集',
                count: s.meta?.total ?? 0,
                subtitle: '合集',
                kind: 'season' as const,
                raw: s,
                cover: s.meta?.cover,
            })), ...allSeries.map(s => ({
                id: s.meta?.series_id ?? 0,
                title: s.meta?.name ?? '未命名系列',
                count: s.meta?.total ?? 0,
                subtitle: '系列',
                kind: 'series' as const,
                raw: s,
                cover: s.meta?.cover,
            }))].find(i => i.id > 0) ?? null;

            if (first) void loadListDetail(first, false);
        } catch (error) {
            setSsError(error instanceof Error ? error.message : String(error));
        } finally {
            setSsLoading(false);
        }
    };

    const loadInfo = async (midOrSpaceUrl: string) => {
        setLoading(true);
        const seq = ++infoReqSeq;
        try {
            const data = await Info(midOrSpaceUrl);
            if (seq !== infoReqSeq) return;
            setInfo(data);
            // Info 返回后，回填已加载的卡片，修复“偶现 up 名称为未知”的竞态问题。
            const name = data?.name || '未知';
            setVideoCards(prev => prev.map(c => (c.upperName === name ? c : {...c, upperName: name})));
            setListCards(prev => prev.map(c => (c.upperName === name ? c : {...c, upperName: name})));
        } catch (error) {
            if (seq !== infoReqSeq) return;
            showToast(error instanceof Error ? error.message : String(error), 'error');
        } finally {
            if (seq === infoReqSeq) setLoading(false);
        }
    };

    const init = () => {
        const mid = getMid();
        void loadInfo(mid);
        void loadVideoList(false);
    };

    return {
        // Header
        loading,
        info,

        // Tab
        activeTab,
        setActiveTab,

        // 视频 Tab
        videoLoading,
        videoError,
        videoCards,
        videoTotal,
        videoLoadingMore,
        hasMoreVideos,
        loadVideoList,

        // 合集/系列 Tab
        ssLoadedOnce,
        ssLoading,
        ssError,
        listSidebarItems,
        selectedListItem,
        listDetailLoading,
        listDetailError,
        listCards,
        listTotal,
        listLoadingMore,
        hasMoreListVideos,
        loadSeasonsSeriesAll,
        handleSelectListItem,
        handleLoadMoreList,

        // 选择/下载
        selectedMediaIds,
        selectedSet,
        allSelected,
        toggleSelectMedia,
        toggleSelectAllMedia,
        clearSelection,
        downloadSelectedMedia,
        downloadAllMedia,
        downloadOneMedia,

        // 生命周期
        init,

        // 给 UI 层的便捷操作
        retryListDetail,
    };
}

