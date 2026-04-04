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
import {Favorites, FavoritesList} from "../../../wailsjs/go/api/BiliBili";
import {model} from "../../../wailsjs/go/models";
import ErrorToast from "../../components/ErrorToast";
import StatusToast from "../../components/StatusToast";
import {BrowserOpenURL, ClipboardSetText} from "../../../wailsjs/runtime/runtime";

export const Route = createFileRoute('/bilibili/favorite')({
    component: Favorite,
})


const EMPTY_FAVORITE_LIST: readonly model.FavoriteItem[] = [];
const FAVORITE_PAGE = 1;
const FAVORITE_PAGE_SIZE = 40;

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

function FavoriteDetailPanel(props: {
    item: model.FavoriteItem;
    detail: model.FavoriteData;
    selectedMediaIds: () => readonly number[];
    onToggleSelect: (id: number) => void;
    onToggleSelectAll: () => void;
    onDownloadOne: (media: model.FavoriteMedias) => void;
    onDownloadSelected: () => void;
    onDownloadAll: () => void;
    hasMore: () => boolean;
    loadingMore: () => boolean;
    onLoadMore: () => void;
}): JSXElement {
    const selectedSet = () => new Set(props.selectedMediaIds());
    const allSelected = () => {
        const medias = props.detail.medias ?? [];
        if (medias.length === 0) return false;
        const selected = selectedSet();
        return medias.every((media) => selected.has(media.id));
    };

    return (
        <div class="flex h-full min-h-0 flex-col">
            <div class="flex shrink-0 items-center gap-2 border-b border-base-300 px-5 py-3.5">
                <div class="min-w-0 flex-1">
                    <h2 class="truncate text-sm font-bold text-base-content">{props.item.title}</h2>
                    <p class="text-xs text-orange-500">
                        {props.detail.info?.media_count ?? props.item.media_count} 个视频
                    </p>
                </div>
                <button class="btn btn-ghost btn-sm" onClick={props.onToggleSelectAll}>
                    {allSelected() ? '取消全选' : '全选'}
                </button>
                <button class="btn btn-outline btn-primary btn-sm" onClick={props.onDownloadSelected}
                        disabled={props.selectedMediaIds().length === 0}>
                    下载已选 ({props.selectedMediaIds().length})
                </button>
                <button class="btn btn-primary btn-sm" onClick={props.onDownloadAll}>下载全部</button>
            </div>

            <div class="min-h-0 flex-1 overflow-auto p-3">
                <div class="overflow-hidden rounded-xl border border-base-300 bg-base-100">
                    <For each={props.detail.medias ?? []}>
                        {(media) => (
                            <div
                                class="flex items-center gap-3 border-b border-base-300/80 px-3 py-2.5 transition-colors hover:bg-base-200/40 last:border-b-0">
                                <div class="h-14 w-24 shrink-0 overflow-hidden rounded bg-base-200">
                                    <Show
                                        when={media.cover}
                                        fallback={<div
                                            class="flex h-full w-full items-center justify-center text-xs text-base-content/40">无封面</div>}
                                    >
                                        <img src={media.cover} alt={media.title}
                                             class="h-full w-full object-cover" loading="lazy"/>
                                    </Show>
                                </div>
                                <div class="min-w-0 flex-1">
                                    <p class="truncate text-sm font-semibold text-base-content">{media.title}</p>
                                    <p class="mt-1 truncate text-xs text-base-content/60">UP: {media.upper?.name || '未知'}</p>
                                    <p class="mt-1 text-xs text-base-content/50">
                                        时长 {formatDuration(media.duration)} ·
                                        播放 {formatCount(media.cnt_info?.play)} ·
                                        弹幕 {formatCount(media.cnt_info?.danmaku)}
                                    </p>
                                </div>
                                <div class="flex shrink-0 items-center gap-2">
                                    {/*选择单个视频是否下载*/}
                                    <input
                                        type="checkbox"
                                        class="checkbox checkbox-md checkbox-success"
                                        checked={selectedSet().has(media.id)}
                                        onChange={() => props.onToggleSelect(media.id)}
                                    />
                                    <button class="btn btn-outline btn-xs" onClick={() => props.onDownloadOne(media)}>
                                        下载
                                    </button>
                                </div>
                            </div>
                        )}
                    </For>
                </div>

                <Show when={props.hasMore()}>
                    <div class="mt-3 flex justify-center">
                        <button class="btn btn-outline btn-sm" onClick={props.onLoadMore}
                                disabled={props.loadingMore()}>
                            {props.loadingMore() ? '加载中...' : '加载剩余视频'}
                        </button>
                    </div>
                </Show>
            </div>
        </div>
    );
}

function FavoriteDetailLoading(): JSXElement {
    return (
        <div class="flex h-full items-center justify-center">
            <span class="loading loading-spinner loading-md text-primary"></span>
        </div>
    );
}

function FavoriteDetailError(props: { message: string; onRetry: () => void }): JSXElement {
    return (
        <div class="flex h-full items-center justify-center p-6">
            <div class="rounded-xl border border-error/30 bg-error/10 px-5 py-4 text-center">
                <p class="text-sm font-semibold text-error">加载收藏夹详情失败</p>
                <p class="mt-1 max-w-md text-xs text-error/80">{props.message}</p>
                <button class="btn btn-error btn-sm mt-3" onClick={props.onRetry}>重试</button>
            </div>
        </div>
    );
}

function Favorite(): JSXElement {
    const [loading, setLoading] = createSignal(true);
    const [favorites, setFavorites] = createSignal<model.FavoritesData | null>(null);
    const [errorText, setErrorText] = createSignal('');
    const [selectedItem, setSelectedItem] = createSignal<model.FavoriteItem | null>(null);
    const [detailLoading, setDetailLoading] = createSignal(false);
    const [loadingMore, setLoadingMore] = createSignal(false);
    const [detailError, setDetailError] = createSignal('');
    const [favoriteDetail, setFavoriteDetail] = createSignal<model.FavoriteData | null>(null);
    const [detailPage, setDetailPage] = createSignal(FAVORITE_PAGE);
    const [detailHasMore, setDetailHasMore] = createSignal(false);
    const [selectedMediaIds, setSelectedMediaIds] = createSignal<number[]>([]);
    const [statusText, setStatusText] = createSignal('');
    const [statusTone, setStatusTone] = createSignal<"info" | "success" | "warning">("info");
    let detailRequestSeq = 0;

    const listAccessor = () => favorites()?.list ?? EMPTY_FAVORITE_LIST;
    const selectedId = () => selectedItem()?.id ?? null;

    const pushStatus = (message: string, tone: "info" | "success" | "warning" = "info") => {
        setStatusTone(tone);
        setStatusText(message);
    };

    const resolveMediaURL = (media: model.FavoriteMedias): string => {
        if (media.link) return media.link;
        if (media.bvid) return `https://www.bilibili.com/video/${media.bvid}`;
        return '';
    };

    const downloadMediaList = async (medias: model.FavoriteMedias[], label: string) => {
        const urls = medias.map(resolveMediaURL).filter((url) => !!url);
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

    const toggleSelectMedia = (id: number) => {
        setSelectedMediaIds((prev) => {
            const set = new Set(prev);
            if (set.has(id)) set.delete(id);
            else set.add(id);
            return Array.from(set);
        });
    };

    const toggleSelectAllMedia = () => {
        const medias = favoriteDetail()?.medias ?? [];
        const ids = medias.map((media) => media.id);
        const selected = new Set(selectedMediaIds());
        const allSelected = ids.length > 0 && ids.every((id) => selected.has(id));
        if (allSelected) {
            setSelectedMediaIds([]);
            return;
        }
        setSelectedMediaIds(ids);
    };

    const downloadSelectedMedia = async () => {
        const selected = new Set(selectedMediaIds());
        const medias = (favoriteDetail()?.medias ?? []).filter((media) => selected.has(media.id));
        await downloadMediaList(medias, `已选择 ${medias.length} 个视频`);
    };

    const downloadAllMedia = async () => {
        await downloadMediaList(favoriteDetail()?.medias ?? [], '全部视频');
    };

    const downloadOneMedia = async (media: model.FavoriteMedias) => {
        await downloadMediaList([media], `视频 ${media.title}`);
    };

    const loadFavoriteDetail = async (item: model.FavoriteItem, append = false) => {
        const targetPage = append ? detailPage() + 1 : FAVORITE_PAGE;
        setSelectedItem(item);
        if (append) {
            setLoadingMore(true);
        } else {
            setDetailLoading(true);
            setDetailError('');
            setFavoriteDetail(null);
            setDetailPage(FAVORITE_PAGE);
            setDetailHasMore(false);
            setSelectedMediaIds([]);
        }

        const requestSeq = ++detailRequestSeq;
        try {
            const data = await Favorites(item.id, targetPage, FAVORITE_PAGE_SIZE);
            if (requestSeq !== detailRequestSeq) return;

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
            setDetailPage(targetPage);
            setDetailHasMore(!!data.has_more);
        } catch (error) {
            if (requestSeq !== detailRequestSeq) return;
            const message = error instanceof Error ? error.message : String(error);
            if (append) {
                pushStatus(`加载更多失败: ${message}`, 'warning');
            } else {
                setFavoriteDetail(null);
                setDetailError(message);
            }
        } finally {
            if (requestSeq === detailRequestSeq) {
                if (append) {
                    setLoadingMore(false);
                } else {
                    setDetailLoading(false);
                }
            }
        }
    };

    const handleSelect = (item: model.FavoriteItem) => {
        if (selectedItem()?.id === item.id && favoriteDetail() && !detailLoading()) return;
        void loadFavoriteDetail(item);
    };

    const handleLoadMore = () => {
        const item = selectedItem();
        if (!item || !detailHasMore() || loadingMore()) return;
        void loadFavoriteDetail(item, true);
    };

    const loadFavorites = async () => {
        setLoading(true);
        setErrorText('');
        try {
            const data = await FavoritesList();
            setFavorites(data);
            const current = selectedItem()?.id;
            const nextItem = data.list?.find((item) => item.id === current) ?? data.list?.[0];
            if (nextItem) {
                void loadFavoriteDetail(nextItem);
            } else {
                setSelectedItem(null);
                setFavoriteDetail(null);
                setDetailError('');
                setDetailPage(FAVORITE_PAGE);
                setDetailHasMore(false);
                setSelectedMediaIds([]);
            }
        } catch (error) {
            setErrorText(error instanceof Error ? error.message : String(error));
        } finally {
            setLoading(false);
        }
    }

    onMount(() => {
        void loadFavorites();
    });

    return (
        <section class="flex h-full min-h-0 gap-3 overflow-hidden bg-base-200/40 p-3">
            {/* 左侧收藏夹列表 */}
            <aside class="flex w-56 shrink-0 flex-col overflow-hidden rounded-xl border border-base-300 bg-base-100">
                <div class="flex shrink-0 items-center gap-1.5 border-b border-base-200 px-3 py-2.5">
                    <h1 class="min-w-0 flex-1 text-sm font-bold text-base-content">
                        收藏夹
                        <span class="ml-1 text-xs font-normal text-base-content/60 tabular-nums">
                            {favorites()?.count ?? 0}
                        </span>
                    </h1>
                    <button
                        class="flex h-6 w-6 shrink-0 items-center justify-center rounded transition-colors hover:bg-base-200 disabled:cursor-not-allowed"
                        onClick={() => void loadFavorites()}
                        disabled={loading()}
                        title="刷新"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg"
                             class={`h-3 w-3 text-base-content/50 ${loading() ? 'animate-spin' : ''}`} fill="none"
                             viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                            <path stroke-linecap="round" stroke-linejoin="round"
                                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
                        </svg>
                    </button>
                </div>

                <Switch>
                    <Match when={loading()}>
                        <div class="flex flex-1 items-center justify-center py-12">
                            <span class="loading loading-spinner loading-sm text-primary"></span>
                        </div>
                    </Match>
                    <Match when={!loading() && (favorites()?.list?.length || 0) === 0}>
                        <EmptyState title="暂无收藏夹" compact/>
                    </Match>
                    <Match when={!loading() && (favorites()?.list?.length || 0) > 0}>
                        <FavoriteSidebar
                            list={listAccessor}
                            selectedId={selectedId}
                            onSelect={handleSelect}
                        />
                    </Match>
                </Switch>
            </aside>

            {/* 右侧视频内容区 */}
            <main class="flex min-w-0 flex-1 flex-col overflow-hidden rounded-xl border border-base-300 bg-base-100">
                <Show
                    when={selectedItem()}
                    fallback={
                        <EmptyState title="选择一个收藏夹查看内容" description="右侧将展示该收藏夹的视频列表"/>
                    }
                >
                    {(item) => (
                        <Switch>
                            <Match when={detailLoading()}>
                                <FavoriteDetailLoading/>
                            </Match>
                            <Match when={!detailLoading() && !!detailError()}>
                                <FavoriteDetailError message={detailError()}
                                                     onRetry={() => void loadFavoriteDetail(item())}/>
                            </Match>
                            <Match
                                when={!detailLoading() && !!favoriteDetail() && (favoriteDetail()?.medias?.length || 0) === 0}>
                                <EmptyState title="该收藏夹暂无视频" description="可以切换到其他收藏夹查看"/>
                            </Match>
                            <Match when={!!favoriteDetail()}>
                                <FavoriteDetailPanel
                                    item={item()}
                                    detail={favoriteDetail()!}
                                    selectedMediaIds={selectedMediaIds}
                                    onToggleSelect={toggleSelectMedia}
                                    onToggleSelectAll={toggleSelectAllMedia}
                                    onDownloadOne={(media) => void downloadOneMedia(media)}
                                    onDownloadSelected={() => void downloadSelectedMedia()}
                                    onDownloadAll={() => void downloadAllMedia()}
                                    hasMore={detailHasMore}
                                    loadingMore={loadingMore}
                                    onLoadMore={handleLoadMore}
                                />
                            </Match>
                            <Match when={true}>
                                <FavoriteDetailLoading/>
                            </Match>
                        </Switch>
                    )}
                </Show>
            </main>

            <ErrorToast message={errorText()}/>
            <StatusToast message={statusText()} tone={statusTone()}/>
        </section>
    )
}

