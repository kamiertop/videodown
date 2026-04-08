import {createFileRoute, useNavigate} from '@tanstack/solid-router'
import {createMemo, createSignal, Match, onMount, Show, Switch, type JSXElement} from "solid-js";
import {Collection, CollectionItem, Favorites, FavoritesList} from "../../../wailsjs/go/api/BiliBili";
import {model} from "../../../wailsjs/go/models";
import ErrorToast from "../../components/ErrorToast";
import StatusToast from "../../components/StatusToast";
import EmptyState from "../../components/EmptyState";
import DetailLoading from "../../components/DetailLoading";
import DetailError from "../../components/DetailError";
import DetailToolbar from "../../components/DetailToolbar";
import VideoCardGrid, {type MediaCardItem} from "../../components/VideoCardGrid";
import SidebarList, {type SidebarListItem} from "../../components/SidebarList";

export const Route = createFileRoute('/bilibili/favorite')({
    component: Favorite,
})

type SidebarTab = 'favorite' | 'collection';

const EMPTY_FAVORITE_LIST: readonly model.FavoriteItem[] = [];
const EMPTY_COLLECTION_LIST: readonly model.CollectionDataList[] = [];
const FAVORITE_PAGE = 1;
const FAVORITE_PAGE_SIZE = 40;
const COLLECTION_PAGE = 1;
const COLLECTION_PAGE_SIZE = 20;

const StarIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20" fill="currentColor">
        <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"/>
    </svg>
);

const StackIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20" fill="currentColor">
        <path d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zM2 11a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"/>
    </svg>
);

interface FavSidebarItem extends SidebarListItem {
    raw: model.FavoriteItem;
}

interface ColSidebarItem extends SidebarListItem {
    raw: model.CollectionDataList;
}

function toFavSidebarItems(list: readonly model.FavoriteItem[]): FavSidebarItem[] {
    return list.map(item => ({id: item.id, title: item.title, count: item.media_count, raw: item}));
}

function toColSidebarItems(list: readonly model.CollectionDataList[]): ColSidebarItem[] {
    return list.map(item => ({
        id: item.id,
        title: item.title,
        count: item.media_count,
        subtitle: item.upper?.name,
        raw: item,
    }));
}

function toMediaCards(medias: model.FavoriteMedias[]): MediaCardItem[] {
    return medias.map(m => ({
        id: m.id,
        title: m.title,
        cover: m.cover,
        duration: m.duration,
        bvid: m.bvid,
        link: m.link,
        upperName: m.upper?.name || '未知',
        play: m.cnt_info?.play,
        danmaku: m.cnt_info?.danmaku,
        pubtime: m.pubtime,
    }));
}

function toCollectionMediaCards(medias: model.CollectionItemMedias[]): MediaCardItem[] {
    return medias.map(m => ({
        id: Number(m.id),
        title: m.title,
        cover: m.cover,
        duration: m.duration,
        bvid: m.bvid,
        upperName: m.upper?.name || '未知',
        play: m.cnt_info?.play,
        danmaku: m.cnt_info?.danmaku,
        pubtime: m.pubtime,
    }));
}

function Favorite(): JSXElement {
    const navigate = useNavigate();
    const [activeTab, setActiveTab] = createSignal<SidebarTab>('favorite');

    const [favLoading, setFavLoading] = createSignal(true);
    const [favorites, setFavorites] = createSignal<model.FavoritesData | null>(null);
    const [selectedFavItem, setSelectedFavItem] = createSignal<model.FavoriteItem | null>(null);
    const [favDetailLoading, setFavDetailLoading] = createSignal(false);
    const [favDetailError, setFavDetailError] = createSignal('');
    const [favoriteDetail, setFavoriteDetail] = createSignal<model.FavoriteData | null>(null);
    const [favDetailPage, setFavDetailPage] = createSignal(FAVORITE_PAGE);
    const [favDetailHasMore, setFavDetailHasMore] = createSignal(false);
    const [favLoadingMore, setFavLoadingMore] = createSignal(false);

    const [colLoading, setColLoading] = createSignal(false);
    const [collections, setCollections] = createSignal<model.CollectionData | null>(null);
    const [colLoadedOnce, setColLoadedOnce] = createSignal(false);
    const [selectedColItem, setSelectedColItem] = createSignal<model.CollectionDataList | null>(null);
    const [colDetailLoading, setColDetailLoading] = createSignal(false);
    const [colDetailError, setColDetailError] = createSignal('');
    const [collectionDetail, setCollectionDetail] = createSignal<model.CollectionItemData | null>(null);

    const [selectedMediaIds, setSelectedMediaIds] = createSignal<number[]>([]);
    const [errorText, setErrorText] = createSignal('');
    const [statusText, setStatusText] = createSignal('');
    const [statusTone, setStatusTone] = createSignal<"info" | "success" | "warning">("info");
    let favDetailRequestSeq = 0;
    let colDetailRequestSeq = 0;

    const favSidebarItems = () => toFavSidebarItems(favorites()?.list ?? EMPTY_FAVORITE_LIST);
    const colSidebarItems = () => toColSidebarItems(collections()?.list ?? EMPTY_COLLECTION_LIST);

    const pushStatus = (message: string, tone: "info" | "success" | "warning" = "info") => {
        setStatusTone(tone);
        setStatusText(message);
    };

    // ── 下载 ──

    const downloadMediaList = (medias: MediaCardItem[], label: string) => {
        const bvids = [...new Set(medias.map(m => m.bvid?.trim()).filter(Boolean))];
        if (bvids.length === 0) {
            pushStatus('没有可下载的视频（缺少 BV 号）', 'warning');
            return;
        }
        navigate({
            to: '/bilibili/download',
            search: {bvids: bvids.join(',')},
        });
        pushStatus(`${label}，已打开下载页`, 'success');
    };

    const currentMediaCards = (): MediaCardItem[] => {
        if (activeTab() === 'favorite') {
            return toMediaCards(favoriteDetail()?.medias ?? []);
        }
        return toCollectionMediaCards(collectionDetail()?.medias ?? []);
    };

    const selectedSet = createMemo(() => new Set(selectedMediaIds()));

    const allSelected = () => {
        const cards = currentMediaCards();
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
        const cards = currentMediaCards();
        setSelectedMediaIds(allSelected() ? [] : cards.map(c => c.id));
    };

    const downloadSelectedMedia = async () => {
        const s = selectedSet();
        const medias = currentMediaCards().filter(m => s.has(m.id));
        await downloadMediaList(medias, `已选择 ${medias.length} 个视频`);
    };

    const downloadAllMedia = () => downloadMediaList(currentMediaCards(), '全部视频');
    const downloadOneMedia = (media: MediaCardItem) => downloadMediaList([media], `视频 ${media.title}`);

    // ── 收藏夹加载 ──

    const loadFavoriteDetail = async (item: model.FavoriteItem, append = false) => {
        const targetPage = append ? favDetailPage() + 1 : FAVORITE_PAGE;
        setSelectedFavItem(item);
        if (append) {
            setFavLoadingMore(true);
        } else {
            setFavDetailLoading(true);
            setFavDetailError('');
            setFavoriteDetail(null);
            setFavDetailPage(FAVORITE_PAGE);
            setFavDetailHasMore(false);
            setSelectedMediaIds([]);
        }

        const seq = ++favDetailRequestSeq;
        try {
            const data = await Favorites(item.id, targetPage, FAVORITE_PAGE_SIZE);
            if (seq !== favDetailRequestSeq) return;

            if (append) {
                setFavoriteDetail(prev => {
                    const prevMedias = prev?.medias ?? [];
                    return model.FavoriteData.createFrom({...data, medias: [...prevMedias, ...(data.medias ?? [])]});
                });
            } else {
                setFavoriteDetail(data);
            }
            setFavDetailPage(targetPage);
            setFavDetailHasMore(!!data.has_more);
        } catch (error) {
            if (seq !== favDetailRequestSeq) return;
            const msg = error instanceof Error ? error.message : String(error);
            if (append) pushStatus(`加载更多失败: ${msg}`, 'warning');
            else { setFavoriteDetail(null); setFavDetailError(msg); }
        } finally {
            if (seq === favDetailRequestSeq) {
                if (append) setFavLoadingMore(false); else setFavDetailLoading(false);
            }
        }
    };

    const handleFavSelect = (si: FavSidebarItem) => {
        const item = si.raw;
        if (selectedFavItem()?.id === item.id && favoriteDetail() && !favDetailLoading()) return;
        void loadFavoriteDetail(item);
    };

    const handleFavLoadMore = () => {
        const item = selectedFavItem();
        if (!item || !favDetailHasMore() || favLoadingMore()) return;
        void loadFavoriteDetail(item, true);
    };

    const loadFavorites = async () => {
        setFavLoading(true);
        setErrorText('');
        try {
            const data = await FavoritesList();
            setFavorites(data);
            const current = selectedFavItem()?.id;
            const nextItem = data.list?.find(i => i.id === current) ?? data.list?.[0];
            if (nextItem) {
                void loadFavoriteDetail(nextItem);
            } else {
                setSelectedFavItem(null);
                setFavoriteDetail(null);
                setFavDetailError('');
                setFavDetailPage(FAVORITE_PAGE);
                setFavDetailHasMore(false);
                setSelectedMediaIds([]);
            }
        } catch (error) {
            setErrorText(error instanceof Error ? error.message : String(error));
        } finally {
            setFavLoading(false);
        }
    };

    // ── 合集加载 ──

    const loadCollectionDetail = async (item: model.CollectionDataList) => {
        setSelectedColItem(item);
        setColDetailLoading(true);
        setColDetailError('');
        setCollectionDetail(null);
        setSelectedMediaIds([]);

        const seq = ++colDetailRequestSeq;
        try {
            const data = await CollectionItem(String(item.id), 1, 20);
            if (seq !== colDetailRequestSeq) return;
            setCollectionDetail(data);
        } catch (error) {
            if (seq !== colDetailRequestSeq) return;
            setColDetailError(error instanceof Error ? error.message : String(error));
        } finally {
            if (seq === colDetailRequestSeq) setColDetailLoading(false);
        }
    };

    const handleColSelect = (si: ColSidebarItem) => {
        const item = si.raw;
        if (selectedColItem()?.id === item.id && collectionDetail() && !colDetailLoading()) return;
        void loadCollectionDetail(item);
    };

    const loadCollections = async () => {
        setColLoading(true);
        try {
            const data = await Collection(COLLECTION_PAGE, COLLECTION_PAGE_SIZE);
            setCollections(data);
            setColLoadedOnce(true);
            if (!selectedColItem() && data.list?.length > 0) {
                void loadCollectionDetail(data.list[0]);
            }
        } catch (error) {
            setErrorText(error instanceof Error ? error.message : String(error));
        } finally {
            setColLoading(false);
        }
    };

    // ── Tab 切换 ──

    const switchTab = (tab: SidebarTab) => {
        if (activeTab() === tab) return;
        setActiveTab(tab);
        setSelectedMediaIds([]);
        if (tab === 'collection' && !colLoadedOnce()) void loadCollections();
    };

    onMount(() => void loadFavorites());

    // ── 派生状态 ──

    const isDetailLoading = () => activeTab() === 'favorite' ? favDetailLoading() : colDetailLoading();
    const detailError = () => activeTab() === 'favorite' ? favDetailError() : colDetailError();
    const hasDetail = () => activeTab() === 'favorite' ? !!favoriteDetail() : !!collectionDetail();
    const detailMediaCount = () => {
        if (activeTab() === 'favorite')
            return favoriteDetail()?.info?.media_count ?? selectedFavItem()?.media_count ?? 0;
        return collectionDetail()?.info?.media_count ?? selectedColItem()?.media_count ?? 0;
    };
    const detailTitle = () => activeTab() === 'favorite'
        ? (selectedFavItem()?.title ?? '')
        : (selectedColItem()?.title ?? '');
    const hasAnySelection = () => activeTab() === 'favorite' ? !!selectedFavItem() : !!selectedColItem();
    const retryDetail = () => {
        if (activeTab() === 'favorite') {
            const item = selectedFavItem();
            if (item) void loadFavoriteDetail(item);
        } else {
            const item = selectedColItem();
            if (item) void loadCollectionDetail(item);
        }
    };
    const sidebarLoading = () => activeTab() === 'favorite' ? favLoading() : colLoading();

    return (
        <section class="flex h-full min-h-0 gap-3 overflow-hidden bg-base-200/40 p-3">
            {/* 左侧 sidebar */}
            <aside class="flex w-56 shrink-0 flex-col overflow-hidden rounded-xl border border-base-300 bg-base-100">
                <div class="flex shrink-0 border-b border-base-200">
                    <button
                        class={`flex-1 py-2.5 text-center text-sm font-bold transition-colors ${
                            activeTab() === 'favorite'
                                ? 'border-b-2 border-success text-success'
                                : 'text-base-content/60 hover:text-base-content'
                        }`}
                        onClick={() => switchTab('favorite')}
                    >
                        收藏夹
                        <span class="ml-1 text-xs font-normal tabular-nums">{favorites()?.count ?? 0}</span>
                    </button>
                    <button
                        class={`flex-1 py-2.5 text-center text-sm font-bold transition-colors ${
                            activeTab() === 'collection'
                                ? 'border-b-2 border-success text-success'
                                : 'text-base-content/60 hover:text-base-content'
                        }`}
                        onClick={() => switchTab('collection')}
                    >
                        合集
                        <span class="ml-1 text-xs font-normal tabular-nums">{collections()?.count ?? 0}</span>
                    </button>
                </div>

                <div class="flex shrink-0 items-center justify-end border-b border-base-200 px-3 py-1.5">
                    <button
                        class="flex h-6 w-6 shrink-0 items-center justify-center rounded transition-colors hover:bg-base-200 disabled:cursor-not-allowed"
                        onClick={() => activeTab() === 'favorite' ? void loadFavorites() : void loadCollections()}
                        disabled={sidebarLoading()}
                        title="刷新"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg"
                             class={`h-3 w-3 text-base-content/50 ${sidebarLoading() ? 'animate-spin' : ''}`}
                             fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                            <path stroke-linecap="round" stroke-linejoin="round"
                                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
                        </svg>
                    </button>
                </div>

                <Switch>
                    <Match when={sidebarLoading()}>
                        <div class="flex flex-1 items-center justify-center py-12">
                            <span class="loading loading-spinner loading-sm text-primary"></span>
                        </div>
                    </Match>
                    <Match when={activeTab() === 'favorite'}>
                        <Show
                            when={favSidebarItems().length > 0}
                            fallback={<EmptyState title="暂无收藏夹" compact/>}
                        >
                            <SidebarList
                                list={favSidebarItems}
                                selectedId={() => selectedFavItem()?.id ?? null}
                                onSelect={handleFavSelect}
                                icon={StarIcon}
                            />
                        </Show>
                    </Match>
                    <Match when={activeTab() === 'collection'}>
                        <Show
                            when={colSidebarItems().length > 0}
                            fallback={<EmptyState title="暂无合集" compact/>}
                        >
                            <SidebarList
                                list={colSidebarItems}
                                selectedId={() => selectedColItem()?.id ?? null}
                                onSelect={handleColSelect}
                                icon={StackIcon}
                            />
                        </Show>
                    </Match>
                </Switch>
            </aside>

            {/* 右侧视频内容区 */}
            <main class="flex min-w-0 flex-1 flex-col overflow-hidden rounded-xl border border-base-300 bg-base-100">
                <Show
                    when={hasAnySelection()}
                    fallback={
                        <EmptyState
                            title={activeTab() === 'favorite' ? '选择一个收藏夹查看内容' : '选择一个合集查看内容'}
                            description="右侧将展示视频列表"
                        />
                    }
                >
                    <Switch>
                        <Match when={isDetailLoading()}>
                            <DetailLoading/>
                        </Match>
                        <Match when={!!detailError()}>
                            <DetailError message={detailError()} onRetry={retryDetail}/>
                        </Match>
                        <Match when={hasDetail() && currentMediaCards().length === 0}>
                            <EmptyState title="暂无视频" description="可以切换到其他项查看"/>
                        </Match>
                        <Match when={hasDetail()}>
                            <DetailToolbar
                                title={detailTitle()}
                                mediaCount={detailMediaCount()}
                                selectedCount={selectedMediaIds().length}
                                allSelected={allSelected()}
                                onToggleSelectAll={toggleSelectAllMedia}
                                onClearSelection={() => setSelectedMediaIds([])}
                                onDownloadSelected={() => void downloadSelectedMedia()}
                                onDownloadAll={() => void downloadAllMedia()}
                            />
                            <div class="min-h-0 flex-1 overflow-auto p-4">
                                <VideoCardGrid
                                    medias={currentMediaCards()}
                                    selectedSet={selectedSet}
                                    onToggleSelect={toggleSelectMedia}
                                    onDownloadOne={m => void downloadOneMedia(m)}
                                />
                                <Show when={activeTab() === 'favorite' && favDetailHasMore()}>
                                    <div class="mt-4 flex justify-center pb-2">
                                        <button class="btn btn-outline btn-sm" onClick={handleFavLoadMore}
                                                disabled={favLoadingMore()}>
                                            {favLoadingMore() ? '加载中...' : '加载剩余视频'}
                                        </button>
                                    </div>
                                </Show>
                            </div>
                        </Match>
                        <Match when={true}>
                            <DetailLoading/>
                        </Match>
                    </Switch>
                </Show>
            </main>

            <ErrorToast message={errorText()}/>
            <StatusToast message={statusText()} tone={statusTone()}/>
        </section>
    );
}
