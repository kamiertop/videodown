import {createFileRoute, Link} from '@tanstack/solid-router'
import {createMemo, createSignal, type JSXElement, Match, onCleanup, onMount, Show, Switch} from "solid-js";
import {model} from "../../../../wailsjs/go/models.ts";
import {Info, SeasonsArchivesList, SeasonsSeriesList, SeriesList, VideoList} from "../../../../wailsjs/go/api/BiliBili";
import UpCommonLayout from "../../../components/UpCommonLayout.tsx";
import ErrorToast from "../../../components/ErrorToast.tsx";
import StatusToast from "../../../components/StatusToast.tsx";
import EmptyState from "../../../components/EmptyState.tsx";
import DetailLoading from "../../../components/DetailLoading.tsx";
import DetailError from "../../../components/DetailError.tsx";
import DetailToolbar from "../../../components/DetailToolbar.tsx";
import VideoCardGrid, {type MediaCardItem} from "../../../components/VideoCardGrid.tsx";
import SidebarList, {type SidebarListItem} from "../../../components/SidebarList.tsx";
import {BrowserOpenURL, ClipboardSetText} from "../../../../wailsjs/runtime";
import {parseBilibiliLengthToSeconds} from "../../../lib/format.ts";

export const Route = createFileRoute('/bilibili/up/$mid')({
    component: UpDetail,
})

type UpTab = 'videos' | 'lists';

type VideoListResp = {
    list?: { vlist?: Array<{
        aid: number;
        bvid: string;
        title: string;
        pic: string;
        length: string;
        created: number;
        play?: number;
        meta?: { stat?: { danmaku?: number } };
    }> };
    page?: { pn: number; ps: number; count: number };
};

type SeasonArchivesResp = {
    meta?: { title?: string; total?: number; season_id?: number };
    page?: { page_num?: number; page_size?: number; total?: number };
    archives?: Array<{
        aid: number;
        bvid: string;
        title: string;
        pic: string;
        duration: number;
        pubdate: number;
        stat?: { view?: number; danmaku?: number };
    }>;
};

type SeriesArchivesResp = {
    page?: { num?: number; size?: number; total?: number };
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

function resolveVideoURL(bvid?: string, link?: string): string {
    if (link?.trim()) return link.trim();
    if (bvid?.trim()) return `https://www.bilibili.com/video/${bvid.trim()}`;
    return '';
}

function UpDetail(): JSXElement {
    const params = Route.useParams();
    const [loading, setLoading] = createSignal(true);
    const [errorText, setErrorText] = createSignal('');
    const [statusText, setStatusText] = createSignal('');
    const [statusTone, setStatusTone] = createSignal<"info" | "success" | "warning">("info");
    const [info, setInfo] = createSignal<model.UserInfoData | null>(null);

    let errorToastTimer: number | undefined;
    const showErrorToast = (message: string) => {
        setErrorText(message);
        if (errorToastTimer !== undefined) {
            window.clearTimeout(errorToastTimer);
        }
        errorToastTimer = window.setTimeout(() => {
            setErrorText('');
            errorToastTimer = undefined;
        }, 3000)
    }

    const pushStatus = (message: string, tone: "info" | "success" | "warning" = "info") => {
        setStatusTone(tone);
        setStatusText(message);
    };

    // ── Tab ──

    const [activeTab, setActiveTab] = createSignal<UpTab>('videos');

    // ── 全部投稿视频 ──

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

    const mapVlistToCards = (vlist: VideoListResp["list"] extends { vlist?: infer V } ? V : any): MediaCardItem[] => {
        const upperName = info()?.name || '未知';
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

    const loadVideoList = async (mid: string, append = false) => {
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
            const total = Number(data.page?.count) || 0;
            setVideoTotal(total);
            if (append) setVideoCards(prev => [...prev, ...cards]);
            else setVideoCards(cards);
            setVideoPage(targetPage);
        } catch (error) {
            if (seq !== videoReqSeq) return;
            const msg = error instanceof Error ? error.message : String(error);
            if (append) pushStatus(`加载更多失败: ${msg}`, 'warning');
            else setVideoError(msg);
        } finally {
            if (seq === videoReqSeq) {
                if (append) setVideoLoadingMore(false);
                else setVideoLoading(false);
            }
        }
    };

    // ── 合集|系列 ──

    const SS_PAGE_SIZE = 20;
    const [ssLoading, setSsLoading] = createSignal(false);
    const [ssError, setSsError] = createSignal('');
    const [seasons, setSeasons] = createSignal<model.SeasonsItem[]>([]);
    const [series, setSeries] = createSignal<model.SeriesItem[]>([]);
    const [ssLoadedOnce, setSsLoadedOnce] = createSignal(false);

    type ListKind = 'season' | 'series';
    interface ListsSidebarItem extends SidebarListItem {
        kind: ListKind;
        raw: any;
        cover?: string;
    }

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

    const loadSeasonsSeriesAll = async (mid: string) => {
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

            if (first) void loadListDetail(mid, first, false);
        } catch (error) {
            setSsError(error instanceof Error ? error.message : String(error));
        } finally {
            setSsLoading(false);
        }
    };

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

    const loadListDetail = async (mid: string, item: ListsSidebarItem, append: boolean) => {
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
            const upperName = info()?.name || '未知';
            if (item.kind === 'season') {
                const data = await SeasonsArchivesList(mid, 20, targetPage, item.id) as SeasonArchivesResp;
                if (seq !== listReqSeq) return;
                const cards = mapArchivesToCards(data.archives, upperName);
                const total = Number(data.meta?.total ?? data.page?.total ?? item.count ?? 0) || 0;
                setListTotal(total);
                if (append) setListCards(prev => [...prev, ...cards]); else setListCards(cards);
                setListPage(targetPage);
            } else {
                const data = await SeriesList(mid, 20, targetPage, item.id) as SeriesArchivesResp;
                if (seq !== listReqSeq) return;
                const cards = mapArchivesToCards(data.archives as any[], upperName);
                const total = Number(data.page?.total ?? item.count ?? 0) || 0;
                setListTotal(total);
                if (append) setListCards(prev => [...prev, ...cards]); else setListCards(cards);
                setListPage(targetPage);
            }
        } catch (error) {
            if (seq !== listReqSeq) return;
            const msg = error instanceof Error ? error.message : String(error);
            if (append) pushStatus(`加载更多失败: ${msg}`, 'warning');
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
        void loadListDetail(params().mid, item, false);
    };

    const handleLoadMoreList = () => {
        const item = selectedListItem();
        if (!item || !hasMoreListVideos() || listLoadingMore()) return;
        void loadListDetail(params().mid, item, true);
    };

    // ── 选择/下载 ──

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

    const downloadMediaList = async (medias: MediaCardItem[], label: string) => {
        const urls = medias.map(m => resolveVideoURL(m.bvid, m.link)).filter(Boolean);
        if (urls.length === 0) {
            pushStatus('没有可下载的视频链接', 'warning');
            return;
        }
        if (urls.length === 1) {
            BrowserOpenURL(urls[0]);
            pushStatus(`${label}，已在浏览器打开`, 'success');
            return;
        }
        await ClipboardSetText(urls.join('\n'));
        pushStatus(`${label}，链接已复制到剪贴板`, 'success');
    };

    const downloadSelectedMedia = async () => {
        const s = selectedSet();
        const medias = currentCards().filter(m => s.has(m.id));
        await downloadMediaList(medias, `已选择 ${medias.length} 个视频`);
    };

    const downloadAllMedia = () => downloadMediaList(currentCards(), '全部视频');
    const downloadOneMedia = (media: MediaCardItem) => downloadMediaList([media], `视频 ${media.title}`);

    const loadInfo = async (midOrSpaceUrl: string) => {
        setLoading(true);
        setErrorText('');
        try {
            const data = await Info(midOrSpaceUrl);
            setInfo(data);
        } catch (error) {
            showErrorToast(error instanceof Error ? error.message : String(error));
        } finally {
            setLoading(false);
        }
    };

    onMount(() => {
        const mid = params().mid;
        void loadInfo(mid);
        void loadVideoList(mid, false);
    })
    onCleanup(() => {
        if (errorToastTimer !== undefined) {
            window.clearTimeout(errorToastTimer);
        }
    })

    return (
        <UpCommonLayout
            headerLeft={
                <>
                    <Link
                        to="/bilibili/up"
                        class="btn btn-ghost btn-sm gap-1"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24"
                             stroke="currentColor" stroke-width="2">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7"/>
                        </svg>
                        返回
                    </Link>
                    <div class="h-5 w-px bg-base-300"></div>
                    <h2 class="text-sm font-bold text-base-content">UP主详情</h2>
                    <span class="rounded-full bg-base-200 px-2 py-0.5 text-xs tabular-nums text-base-content/60">
                        mid: {params().mid}
                    </span>
                </>
            }
            headerRight={
                <Switch>
                    <Match when={loading()}>
                        <div class="flex items-center gap-2">
                            <span class="loading loading-spinner loading-xs text-primary"></span>
                            <span class="text-xs text-base-content/50">获取UP主信息...</span>
                        </div>
                    </Match>
                    <Match when={!loading() && info()}>
                        <div class="flex min-w-0 items-center gap-2">
                            <div class="h-8 w-8 shrink-0 overflow-hidden rounded-full bg-base-200 ring-2 ring-base-200">
                                <img
                                    src={info()!.face}
                                    alt={info()!.name}
                                    referrerPolicy="no-referrer"
                                    class="h-full w-full object-cover"
                                />
                            </div>
                            <div class="min-w-0">
                                <div class="flex min-w-0 items-center gap-2">
                                    <span class="max-w-[16rem] truncate text-sm font-black text-base-content">
                                        {info()!.name}
                                    </span>
                                    <span class="badge badge-outline badge-sm">Lv.{info()!.level}</span>
                                    <span
                                        class={`badge badge-sm ${info()!.is_followed ? 'badge-primary' : 'badge-ghost'}`}
                                    >
                                        {info()!.is_followed ? '已关注' : '未关注'}
                                    </span>
                                </div>
                            </div>
                        </div>
                    </Match>
                </Switch>
            }
        >
            <section class="flex h-full min-h-0 gap-3 overflow-hidden bg-base-200/40 p-3">
                <main class="flex min-w-0 flex-1 flex-col overflow-hidden rounded-xl border border-base-300 bg-base-100">
                    {/* Tab 切换 */}
                    <div class="flex shrink-0 border-b border-base-300">
                        <button
                            class={`flex-1 py-3 text-center text-sm font-bold transition-colors ${
                                activeTab() === 'videos'
                                    ? 'border-b-2 border-success text-success'
                                    : 'text-base-content/60 hover:text-base-content'
                            }`}
                            onClick={() => setActiveTab('videos')}
                        >
                            全部视频
                            <span class="ml-1 text-xs font-normal tabular-nums">{videoTotal() || videoCards().length}</span>
                        </button>
                        <button
                            class={`flex-1 py-3 text-center text-sm font-bold transition-colors ${
                                activeTab() === 'lists'
                                    ? 'border-b-2 border-success text-success'
                                    : 'text-base-content/60 hover:text-base-content'
                            }`}
                            onClick={() => {
                                setActiveTab('lists');
                                if (!ssLoadedOnce()) void loadSeasonsSeriesAll(params().mid);
                            }}
                        >
                            合集 | 系列
                            <span class="ml-1 text-xs font-normal tabular-nums">{listSidebarItems().length}</span>
                        </button>
                    </div>

                    <Switch>
                        {/* 全部视频 */}
                        <Match when={activeTab() === 'videos'}>
                            <Switch>
                                <Match when={videoLoading()}>
                                    <DetailLoading/>
                                </Match>
                                <Match when={!!videoError()}>
                                    <DetailError message={videoError()} onRetry={() => void loadVideoList(params().mid, false)}/>
                                </Match>
                                <Match when={!videoLoading() && videoCards().length === 0}>
                                    <EmptyState title="暂无视频" description="该 UP 主暂无投稿视频或接口返回为空"/>
                                </Match>
                                <Match when={true}>
                                    <DetailToolbar
                                        title="全部投稿视频"
                                        mediaCount={videoTotal() || videoCards().length}
                                        selectedCount={selectedMediaIds().length}
                                        allSelected={allSelected()}
                                        onToggleSelectAll={toggleSelectAllMedia}
                                        onClearSelection={() => setSelectedMediaIds([])}
                                        onDownloadSelected={() => void downloadSelectedMedia()}
                                        onDownloadAll={() => void downloadAllMedia()}
                                    />
                                    <div class="min-h-0 flex-1 overflow-auto p-4">
                                        <VideoCardGrid
                                            medias={videoCards()}
                                            selectedSet={selectedSet}
                                            onToggleSelect={toggleSelectMedia}
                                            onDownloadOne={m => void downloadOneMedia(m)}
                                        />
                                        <Show when={hasMoreVideos()}>
                                            <div class="mt-4 flex justify-center pb-2">
                                                <button
                                                    class="btn btn-outline btn-sm"
                                                    onClick={() => void loadVideoList(params().mid, true)}
                                                    disabled={videoLoadingMore()}
                                                >
                                                    {videoLoadingMore() ? '加载中...' : '加载更多视频'}
                                                </button>
                                            </div>
                                        </Show>
                                    </div>
                                </Match>
                            </Switch>
                        </Match>

                        {/* 合集 | 系列 */}
                        <Match when={activeTab() === 'lists'}>
                            <div class="flex min-h-0 flex-1 overflow-hidden">
                                {/* 左侧 sidebar */}
                                <aside class="flex w-64 shrink-0 flex-col overflow-hidden border-r border-base-300 bg-base-100">
                                    <div class="flex shrink-0 items-center justify-between border-b border-base-200 px-3 py-2">
                                        <span class="text-xs font-bold text-base-content/70">列表</span>
                                        <button
                                            class="flex h-7 w-7 shrink-0 items-center justify-center rounded transition-colors hover:bg-base-200 disabled:cursor-not-allowed"
                                            onClick={() => void loadSeasonsSeriesAll(params().mid)}
                                            disabled={ssLoading()}
                                            title="刷新"
                                        >
                                            <svg xmlns="http://www.w3.org/2000/svg"
                                                 class={`h-3.5 w-3.5 text-base-content/50 ${ssLoading() ? 'animate-spin' : ''}`}
                                                 fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                                                <path stroke-linecap="round" stroke-linejoin="round"
                                                      d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
                                            </svg>
                                        </button>
                                    </div>
                                    <Switch>
                                        <Match when={ssLoading()}>
                                            <div class="flex flex-1 items-center justify-center py-12">
                                                <span class="loading loading-spinner loading-sm text-primary"></span>
                                            </div>
                                        </Match>
                                        <Match when={!!ssError()}>
                                            <div class="p-3">
                                                <DetailError message={ssError()} onRetry={() => void loadSeasonsSeriesAll(params().mid)}/>
                                            </div>
                                        </Match>
                                        <Match when={listSidebarItems().length === 0}>
                                            <EmptyState title="暂无合集/系列" compact/>
                                        </Match>
                                        <Match when={true}>
                                            <SidebarList
                                                list={listSidebarItems}
                                                selectedId={() => selectedListItem()?.id ?? null}
                                                onSelect={handleSelectListItem}
                                                icon={() => (
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3"
                                                         viewBox="0 0 20 20" fill="currentColor">
                                                        <path d="M4 3a2 2 0 012-2h8a2 2 0 012 2v12a2 2 0 01-2 2H6a2 2 0 01-2-2V3z"/>
                                                    </svg>
                                                )}
                                            />
                                        </Match>
                                    </Switch>
                                </aside>

                                {/* 右侧视频内容区 */}
                                <div class="flex min-w-0 flex-1 flex-col overflow-hidden">
                                    <Show
                                        when={!!selectedListItem()}
                                        fallback={<EmptyState title="选择一个合集/系列查看内容" description="右侧将展示视频列表"/>}
                                    >
                                        <Switch>
                                            <Match when={listDetailLoading()}>
                                                <DetailLoading/>
                                            </Match>
                                            <Match when={!!listDetailError()}>
                                                <DetailError message={listDetailError()} onRetry={() => {
                                                    const it = selectedListItem();
                                                    if (it) void loadListDetail(params().mid, it, false);
                                                }}/>
                                            </Match>
                                            <Match when={!listDetailLoading() && listCards().length === 0}>
                                                <EmptyState title="暂无视频" description="该合集/系列暂无可用视频或接口返回为空"/>
                                            </Match>
                                            <Match when={true}>
                                                <DetailToolbar
                                                    title={`${selectedListItem()!.subtitle}: ${selectedListItem()!.title}`}
                                                    mediaCount={listTotal() || selectedListItem()!.count || listCards().length}
                                                    selectedCount={selectedMediaIds().length}
                                                    allSelected={allSelected()}
                                                    onToggleSelectAll={toggleSelectAllMedia}
                                                    onClearSelection={() => setSelectedMediaIds([])}
                                                    onDownloadSelected={() => void downloadSelectedMedia()}
                                                    onDownloadAll={() => void downloadAllMedia()}
                                                />
                                                <div class="min-h-0 flex-1 overflow-auto p-4">
                                                    <VideoCardGrid
                                                        medias={listCards()}
                                                        selectedSet={selectedSet}
                                                        onToggleSelect={toggleSelectMedia}
                                                        onDownloadOne={m => void downloadOneMedia(m)}
                                                    />
                                                    <Show when={hasMoreListVideos()}>
                                                        <div class="mt-4 flex justify-center pb-2">
                                                            <button class="btn btn-outline btn-sm"
                                                                    onClick={handleLoadMoreList}
                                                                    disabled={listLoadingMore()}>
                                                                {listLoadingMore() ? '加载中...' : '加载更多视频'}
                                                            </button>
                                                        </div>
                                                    </Show>
                                                </div>
                                            </Match>
                                        </Switch>
                                    </Show>
                                </div>
                            </div>
                        </Match>
                    </Switch>
                </main>
            </section>

            <ErrorToast message={errorText()}/>
            <StatusToast message={statusText()} tone={statusTone()}/>
        </UpCommonLayout>
    );
}
