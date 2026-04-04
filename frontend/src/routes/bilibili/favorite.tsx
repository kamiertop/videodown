import {createFileRoute} from '@tanstack/solid-router'
import {
    createSignal,
    For,
    Match,
    onMount,
    Show,
    Switch,
    type JSXElement
} from "solid-js";
import {Collection, CollectionItem, Favorites, FavoritesList} from "../../../wailsjs/go/api/BiliBili";
import {model} from "../../../wailsjs/go/models";
import ErrorToast from "../../components/ErrorToast";
import StatusToast from "../../components/StatusToast";
import {BrowserOpenURL, ClipboardSetText} from "../../../wailsjs/runtime/runtime";

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

function formatDuration(totalSeconds: number): string {
    const safe = Number.isFinite(totalSeconds) ? Math.max(0, Math.floor(totalSeconds)) : 0;
    const hours = Math.floor(safe / 3600);
    const minutes = Math.floor((safe % 3600) / 60);
    const seconds = safe % 60;
    if (hours > 0) {
        return `${hours}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
    }
    return `${minutes}:${String(seconds).padStart(2, '0')}`;
}

function formatCount(value?: number): string {
    if (!value || value <= 0) return '0';
    if (value >= 10000) return `${(value / 10000).toFixed(1)}万`;
    return String(value);
}

function EmptyState(props: { title: string; description?: string; compact?: boolean }): JSXElement {
    return (
        <div class={`flex h-full items-center justify-center text-base-content/40 ${props.compact ? 'py-10' : ''}`}>
            <div class="text-center">
                <svg xmlns="http://www.w3.org/2000/svg" class="mx-auto h-14 w-14" fill="none" viewBox="0 0 24 24"
                     stroke="currentColor" stroke-width="1">
                    <path stroke-linecap="round" stroke-linejoin="round"
                          d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
                </svg>
                <p class="mt-3 text-sm font-semibold text-base-content/60">{props.title}</p>
                <Show when={props.description}>
                    <p class="mt-1 text-xs text-base-content/50">{props.description}</p>
                </Show>
            </div>
        </div>
    );
}

// ─── 收藏夹 sidebar 列表 ────────────────────────────────────────────

function FavoriteSidebar(props: {
    list: () => readonly model.FavoriteItem[];
    selectedId: () => number | null;
    onSelect: (item: model.FavoriteItem) => void;
}): JSXElement {
    return (
        <div class="min-h-0 flex-1 overflow-y-auto overscroll-y-contain px-2 py-1">
            <For each={props.list()}>
                {(item) => {
                    const isSelected = () => props.selectedId() === item.id;
                    return (
                        <button
                            class={`group my-0.5 flex h-[42px] w-full items-center gap-2.5 rounded-lg px-2.5 text-left transition-all duration-100 ${
                                isSelected()
                                    ? 'bg-success/15 ring-1 ring-success/30'
                                    : 'hover:bg-accent/10'
                            }`}
                            onClick={() => props.onSelect(item)}
                        >
                            <div
                                class={`flex h-6 w-6 shrink-0 items-center justify-center rounded-md border transition-colors ${
                                    isSelected()
                                        ? 'border-success/40 bg-success/20 text-success'
                                        : 'border-base-300 bg-base-200/50 text-accent group-hover:border-accent/50 group-hover:bg-accent/10'
                                }`}>
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20"
                                     fill="currentColor">
                                    <path
                                        d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"/>
                                </svg>
                            </div>
                            <span class={`min-w-0 flex-1 truncate text-sm font-semibold ${
                                isSelected() ? 'text-success' : 'text-base-content'
                            }`}>{item.title}</span>
                            <span class={`shrink-0 rounded-full px-1.5 py-0.5 text-xs tabular-nums font-bold ${
                                isSelected()
                                    ? 'bg-success/15 text-success'
                                    : 'bg-base-200 text-base-content/70'
                            }`}>{item.media_count}</span>
                        </button>
                    );
                }}
            </For>
        </div>
    );
}

// ─── 合集 sidebar 列表 ──────────────────────────────────────────────

function CollectionSidebar(props: {
    list: () => readonly model.CollectionDataList[];
    selectedId: () => number | null;
    onSelect: (item: model.CollectionDataList) => void;
}): JSXElement {
    return (
        <div class="min-h-0 flex-1 overflow-y-auto overscroll-y-contain px-2 py-1">
            <For each={props.list()}>
                {(item) => {
                    const isSelected = () => props.selectedId() === item.id;
                    return (
                        <button
                            class={`group my-0.5 flex h-[42px] w-full items-center gap-2.5 rounded-lg px-2.5 text-left transition-all duration-100 ${
                                isSelected()
                                    ? 'bg-success/15 ring-1 ring-success/30'
                                    : 'hover:bg-accent/10'
                            }`}
                            onClick={() => props.onSelect(item)}
                        >
                            <div
                                class={`flex h-6 w-6 shrink-0 items-center justify-center rounded-md border transition-colors ${
                                    isSelected()
                                        ? 'border-success/40 bg-success/20 text-success'
                                        : 'border-base-300 bg-base-200/50 text-accent group-hover:border-accent/50 group-hover:bg-accent/10'
                                }`}>
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20"
                                     fill="currentColor">
                                    <path
                                        d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zM2 11a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"/>
                                </svg>
                            </div>
                            <div class="min-w-0 flex-1">
                                <span class={`block truncate text-sm font-semibold ${
                                    isSelected() ? 'text-success' : 'text-base-content'
                                }`}>{item.title}</span>
                                <span class="block truncate text-xs text-base-content/50">{item.upper?.name}</span>
                            </div>
                            <span class={`shrink-0 rounded-full px-1.5 py-0.5 text-xs tabular-nums font-bold ${
                                isSelected()
                                    ? 'bg-success/15 text-success'
                                    : 'bg-base-200 text-base-content/70'
                            }`}>{item.media_count}</span>
                        </button>
                    );
                }}
            </For>
        </div>
    );
}

// ─── 视频卡片（共用） ───────────────────────────────────────────────

interface MediaCardItem {
    id: number;
    title: string;
    cover: string;
    duration: number;
    bvid: string;
    link?: string;
    upperName: string;
    play?: number;
    danmaku?: number;
    pubtime?: number;
}

function formatDate(timestamp?: number): string {
    if (!timestamp) return '';
    const d = new Date(timestamp * 1000);
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    return `${y}-${m}-${day}`;
}

function VideoCardGrid(props: {
    medias: MediaCardItem[];
    selectedSet: () => Set<number>;
    onToggleSelect: (id: number) => void;
    onDownloadOne: (media: MediaCardItem) => void;
}): JSXElement {
    return (
        <div class="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-4">
            <For each={props.medias}>
                {(media) => (
                    <div
                        class={`group relative flex flex-col rounded-lg transition-all duration-150 hover:shadow-md ${
                            props.selectedSet().has(media.id)
                                ? 'border-2 border-success bg-success/5 shadow-sm shadow-success/20'
                                : 'border border-base-300 bg-base-100 hover:border-base-content/20'
                        }`}
                    >
                        <div
                            class="relative aspect-video w-full cursor-pointer overflow-hidden rounded-t-lg bg-base-200"
                            onClick={() => props.onToggleSelect(media.id)}>
                            <Show
                                when={media.cover}
                                fallback={
                                    <div
                                        class="flex h-full w-full items-center justify-center text-xs text-base-content/40">
                                        无封面
                                    </div>
                                }
                            >
                                <img src={media.cover} alt={media.title}
                                     class="h-full w-full object-cover transition-transform duration-200 group-hover:scale-105"
                                     loading="lazy"/>
                            </Show>

                            <div
                                class="pointer-events-none absolute inset-x-0 bottom-0 flex items-end justify-between bg-linear-to-t from-black/70 to-transparent px-2 pb-1.5 pt-5">
                                <div class="flex items-center gap-2.5 text-white/90">
                                    <span class="flex items-center gap-0.5 text-xs">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3"
                                             viewBox="0 0 20 20" fill="currentColor">
                                            <path d="M10 12a2 2 0 100-4 2 2 0 000 4z"/>
                                            <path fill-rule="evenodd"
                                                  d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z"
                                                  clip-rule="evenodd"/>
                                        </svg>
                                        {formatCount(media.play)}
                                    </span>
                                    <span class="flex items-center gap-0.5 text-xs">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3"
                                             viewBox="0 0 20 20" fill="currentColor">
                                            <path
                                                d="M2 5a2 2 0 012-2h7a2 2 0 012 2v4a2 2 0 01-2 2H9l-3 3v-3H4a2 2 0 01-2-2V5z"/>
                                            <path
                                                d="M15 7v2a4 4 0 01-4 4H9.828l-1.766 1.767c.28.149.599.233.938.233h2l3 3v-3h2a2 2 0 002-2V9a2 2 0 00-2-2h-1z"/>
                                        </svg>
                                        {formatCount(media.danmaku)}
                                    </span>
                                </div>
                                <span
                                    class="rounded bg-black/60 px-1 py-0.5 text-xs tabular-nums text-white/90">
                                    {formatDuration(media.duration)}
                                </span>
                            </div>
                        </div>

                        <div class="flex min-h-0 flex-1 cursor-pointer flex-col px-2.5 pb-2.5 pt-2"
                             onClick={() => props.onToggleSelect(media.id)}>
                            {(() => {
                                const [clamped, setClamped] = createSignal(false);
                                let titleRef!: HTMLParagraphElement;
                                const checkClamped = () => setClamped(titleRef.scrollHeight > titleRef.clientHeight);
                                return (
                                    <div class="group/title relative"
                                         onMouseEnter={checkClamped}>
                                        <p ref={titleRef}
                                           class="line-clamp-2 min-h-10 text-sm font-semibold leading-5 text-base-content">
                                            {media.title}
                                        </p>
                                        <Show when={clamped()}>
                                            <div
                                                class="pointer-events-none absolute inset-x-0 top-full z-50 mt-1 hidden rounded-md border border-base-300 bg-base-100 px-3 py-1.5 text-xs text-base-content shadow-md group-hover/title:block">
                                                {media.title}
                                            </div>
                                        </Show>
                                    </div>
                                );
                            })()}
                            <div class="mt-1.5 flex items-center justify-between">
                                <p class="flex min-w-0 flex-1 items-center gap-1 truncate text-xs text-base-content/55">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 shrink-0"
                                         viewBox="0 0 20 20" fill="currentColor">
                                        <path fill-rule="evenodd"
                                              d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"
                                              clip-rule="evenodd"/>
                                    </svg>
                                    <span class="truncate">{media.upperName}</span>
                                </p>
                                <div class="flex shrink-0 items-center gap-1.5">
                                    <Show when={media.pubtime}>
                                        <span
                                            class="text-xs tabular-nums text-base-content/45">{formatDate(media.pubtime)}</span>
                                    </Show>
                                    <button
                                        class="btn btn-ghost btn-xs opacity-0 transition-opacity group-hover:opacity-100"
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            props.onDownloadOne(media);
                                        }}
                                    >
                                        下载
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                )}
            </For>
        </div>
    );
}

// ─── 详情面板顶栏（共用） ───────────────────────────────────────────

function DetailToolbar(props: {
    title: string;
    mediaCount: number;
    selectedCount: number;
    allSelected: boolean;
    onToggleSelectAll: () => void;
    onClearSelection: () => void;
    onDownloadSelected: () => void;
    onDownloadAll: () => void;
}): JSXElement {
    return (
        <div class="flex shrink-0 items-center gap-2 border-b border-base-300 px-5 py-3.5">
            <div class="min-w-0 flex-1">
                <h2 class="truncate text-sm font-bold text-base-content">{props.title}</h2>
                <p class="text-xs text-orange-500">{props.mediaCount} 个视频</p>
            </div>
            <button class="btn btn-ghost btn-sm" onClick={props.onToggleSelectAll}>
                {props.allSelected ? '取消全选' : '全选'}
            </button>
            <Show when={props.selectedCount > 0}>
                <button class="btn btn-ghost btn-sm text-error" onClick={props.onClearSelection}>
                    取消已选
                </button>
            </Show>
            <button class="btn btn-outline btn-primary btn-sm" onClick={props.onDownloadSelected}
                    disabled={props.selectedCount === 0}>
                下载已选 ({props.selectedCount})
            </button>
            <button class="btn btn-primary btn-sm" onClick={props.onDownloadAll}>下载全部</button>
        </div>
    );
}

function DetailLoading(): JSXElement {
    return (
        <div class="flex h-full items-center justify-center">
            <span class="loading loading-spinner loading-md text-primary"></span>
        </div>
    );
}

function DetailError(props: { message: string; onRetry: () => void }): JSXElement {
    return (
        <div class="flex h-full items-center justify-center p-6">
            <div class="rounded-xl border border-error/30 bg-error/10 px-5 py-4 text-center">
                <p class="text-sm font-semibold text-error">加载详情失败</p>
                <p class="mt-1 max-w-md text-xs text-error/80">{props.message}</p>
                <button class="btn btn-error btn-sm mt-3" onClick={props.onRetry}>重试</button>
            </div>
        </div>
    );
}

// ─── 主组件 ─────────────────────────────────────────────────────────

function Favorite(): JSXElement {
    const [activeTab, setActiveTab] = createSignal<SidebarTab>('favorite');

    // ── 收藏夹状态 ──
    const [favLoading, setFavLoading] = createSignal(true);
    const [favorites, setFavorites] = createSignal<model.FavoritesData | null>(null);
    const [selectedFavItem, setSelectedFavItem] = createSignal<model.FavoriteItem | null>(null);
    const [favDetailLoading, setFavDetailLoading] = createSignal(false);
    const [favDetailError, setFavDetailError] = createSignal('');
    const [favoriteDetail, setFavoriteDetail] = createSignal<model.FavoriteData | null>(null);
    const [favDetailPage, setFavDetailPage] = createSignal(FAVORITE_PAGE);
    const [favDetailHasMore, setFavDetailHasMore] = createSignal(false);
    const [favLoadingMore, setFavLoadingMore] = createSignal(false);

    // ── 合集状态 ──
    const [colLoading, setColLoading] = createSignal(false);
    const [collections, setCollections] = createSignal<model.CollectionData | null>(null);
    const [colLoadedOnce, setColLoadedOnce] = createSignal(false);
    const [selectedColItem, setSelectedColItem] = createSignal<model.CollectionDataList | null>(null);
    const [colDetailLoading, setColDetailLoading] = createSignal(false);
    const [colDetailError, setColDetailError] = createSignal('');
    const [collectionDetail, setCollectionDetail] = createSignal<model.CollectionItemData | null>(null);

    // ── 共用状态 ──
    const [selectedMediaIds, setSelectedMediaIds] = createSignal<number[]>([]);
    const [errorText, setErrorText] = createSignal('');
    const [statusText, setStatusText] = createSignal('');
    const [statusTone, setStatusTone] = createSignal<"info" | "success" | "warning">("info");
    let favDetailRequestSeq = 0;
    let colDetailRequestSeq = 0;

    const favListAccessor = () => favorites()?.list ?? EMPTY_FAVORITE_LIST;
    const colListAccessor = () => collections()?.list ?? EMPTY_COLLECTION_LIST;

    const pushStatus = (message: string, tone: "info" | "success" | "warning" = "info") => {
        setStatusTone(tone);
        setStatusText(message);
    };

    // ── 下载逻辑 ──

    const resolveURL = (media: MediaCardItem): string => {
        if (media.link) return media.link;
        if (media.bvid) return `https://www.bilibili.com/video/${media.bvid}`;
        return '';
    };

    const downloadMediaList = async (medias: MediaCardItem[], label: string) => {
        const urls = medias.map(resolveURL).filter(Boolean);
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

    const currentMediaCards = (): MediaCardItem[] => {
        if (activeTab() === 'favorite') {
            return (favoriteDetail()?.medias ?? []).map(m => ({
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
        return (collectionDetail()?.medias ?? []).map(m => ({
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
    };

    const toggleSelectMedia = (id: number) => {
        setSelectedMediaIds((prev) => {
            const set = new Set(prev);
            if (set.has(id)) set.delete(id);
            else set.add(id);
            return Array.from(set);
        });
    };

    const selectedSet = () => new Set(selectedMediaIds());

    const allSelected = () => {
        const cards = currentMediaCards();
        if (cards.length === 0) return false;
        const s = selectedSet();
        return cards.every(c => s.has(c.id));
    };

    const toggleSelectAllMedia = () => {
        const cards = currentMediaCards();
        if (allSelected()) {
            setSelectedMediaIds([]);
        } else {
            setSelectedMediaIds(cards.map(c => c.id));
        }
    };

    const downloadSelectedMedia = async () => {
        const s = selectedSet();
        const medias = currentMediaCards().filter(m => s.has(m.id));
        await downloadMediaList(medias, `已选择 ${medias.length} 个视频`);
    };

    const downloadAllMedia = async () => {
        await downloadMediaList(currentMediaCards(), '全部视频');
    };

    const downloadOneMedia = async (media: MediaCardItem) => {
        await downloadMediaList([media], `视频 ${media.title}`);
    };

    // ── 收藏夹加载逻辑 ──

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

        const requestSeq = ++favDetailRequestSeq;
        try {
            const data = await Favorites(item.id, targetPage, FAVORITE_PAGE_SIZE);
            if (requestSeq !== favDetailRequestSeq) return;

            if (append) {
                setFavoriteDetail((prev) => {
                    const prevMedias = prev?.medias ?? [];
                    return model.FavoriteData.createFrom({
                        ...data,
                        medias: [...prevMedias, ...(data.medias ?? [])],
                    });
                });
            } else {
                setFavoriteDetail(data);
            }
            setFavDetailPage(targetPage);
            setFavDetailHasMore(!!data.has_more);
        } catch (error) {
            if (requestSeq !== favDetailRequestSeq) return;
            const message = error instanceof Error ? error.message : String(error);
            if (append) {
                pushStatus(`加载更多失败: ${message}`, 'warning');
            } else {
                setFavoriteDetail(null);
                setFavDetailError(message);
            }
        } finally {
            if (requestSeq === favDetailRequestSeq) {
                if (append) setFavLoadingMore(false);
                else setFavDetailLoading(false);
            }
        }
    };

    const handleFavSelect = (item: model.FavoriteItem) => {
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
            const nextItem = data.list?.find((item) => item.id === current) ?? data.list?.[0];
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

    // ── 合集加载逻辑 ──

    const loadCollectionDetail = async (item: model.CollectionDataList) => {
        setSelectedColItem(item);
        setColDetailLoading(true);
        setColDetailError('');
        setCollectionDetail(null);
        setSelectedMediaIds([]);

        const requestSeq = ++colDetailRequestSeq;
        try {
            const data = await CollectionItem(String(item.id), 1, 20);
            if (requestSeq !== colDetailRequestSeq) return;
            setCollectionDetail(data);
        } catch (error) {
            if (requestSeq !== colDetailRequestSeq) return;
            setColDetailError(error instanceof Error ? error.message : String(error));
        } finally {
            if (requestSeq === colDetailRequestSeq) {
                setColDetailLoading(false);
            }
        }
    };

    const handleColSelect = (item: model.CollectionDataList) => {
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
        if (tab === 'collection' && !colLoadedOnce()) {
            void loadCollections();
        }
    };

    // ── 初始化 ──

    onMount(() => {
        void loadFavorites();
    });

    // ── 当前详情状态（根据 tab 切换） ──

    const isDetailLoading = () => activeTab() === 'favorite' ? favDetailLoading() : colDetailLoading();
    const detailError = () => activeTab() === 'favorite' ? favDetailError() : colDetailError();
    const hasDetail = () => activeTab() === 'favorite' ? !!favoriteDetail() : !!collectionDetail();
    const detailMediaCount = () => {
        if (activeTab() === 'favorite') {
            return favoriteDetail()?.info?.media_count ?? selectedFavItem()?.media_count ?? 0;
        }
        return collectionDetail()?.info?.media_count ?? selectedColItem()?.media_count ?? 0;
    };
    const detailTitle = () => {
        if (activeTab() === 'favorite') return selectedFavItem()?.title ?? '';
        return selectedColItem()?.title ?? '';
    };
    const hasAnySelection = () => {
        if (activeTab() === 'favorite') return !!selectedFavItem();
        return !!selectedColItem();
    };
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
                {/* Tab 切换 */}
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
                        <span class="ml-1 text-xs font-normal tabular-nums">
                            {favorites()?.count ?? 0}
                        </span>
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
                        <span class="ml-1 text-xs font-normal tabular-nums">
                            {collections()?.count ?? 0}
                        </span>
                    </button>
                </div>

                {/* 刷新按钮 */}
                <div class="flex shrink-0 items-center justify-end border-b border-base-200 px-3 py-1.5">
                    <button
                        class="flex h-6 w-6 shrink-0 items-center justify-center rounded transition-colors hover:bg-base-200 disabled:cursor-not-allowed"
                        onClick={() => activeTab() === 'favorite' ? void loadFavorites() : void loadCollections()}
                        disabled={sidebarLoading()}
                        title="刷新"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg"
                             class={`h-3 w-3 text-base-content/50 ${sidebarLoading() ? 'animate-spin' : ''}`}
                             fill="none"
                             viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                            <path stroke-linecap="round" stroke-linejoin="round"
                                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
                        </svg>
                    </button>
                </div>

                {/* 列表内容 */}
                <Switch>
                    <Match when={sidebarLoading()}>
                        <div class="flex flex-1 items-center justify-center py-12">
                            <span class="loading loading-spinner loading-sm text-primary"></span>
                        </div>
                    </Match>
                    <Match when={activeTab() === 'favorite'}>
                        <Show
                            when={(favListAccessor().length || 0) > 0}
                            fallback={<EmptyState title="暂无收藏夹" compact/>}
                        >
                            <FavoriteSidebar
                                list={favListAccessor}
                                selectedId={() => selectedFavItem()?.id ?? null}
                                onSelect={handleFavSelect}
                            />
                        </Show>
                    </Match>
                    <Match when={activeTab() === 'collection'}>
                        <Show
                            when={(colListAccessor().length || 0) > 0}
                            fallback={
                                <EmptyState title="暂无合集" compact/>
                            }
                        >
                            <CollectionSidebar
                                list={colListAccessor}
                                selectedId={() => selectedColItem()?.id ?? null}
                                onSelect={handleColSelect}
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
                                    onDownloadOne={(m) => void downloadOneMedia(m)}
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
    )
}
