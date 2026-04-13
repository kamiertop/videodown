import {createEffect, createMemo, createSignal, Match, Show, Switch, type JSXElement} from "solid-js";
import EmptyState from "../../EmptyState";
import DetailLoading from "../../DetailLoading";
import DetailError from "../../DetailError";
import DetailToolbar from "../../DetailToolbar";
import VideoCardGrid, {type MediaCardItem} from "../../VideoCardGrid";
import SidebarList, {type SidebarListItem} from "../../SidebarList";
import IconRefresh from "../../icons/IconRefresh";

export default function FavoriteCollectionView<T extends SidebarListItem>(props: {
    emptySidebarTitle: string;
    emptyContentTitle: string;
    sidebarItems: () => readonly T[];
    selectedSidebarId: () => number | null;
    onSelectSidebar: (item: T) => void;
    sidebarIcon: () => JSXElement;
    /** 侧栏左上角标题，如「收藏夹」「合集」 */
    sidebarLabel: () => string;
    /** 对应列表数量（与顶栏 tab 数字同源） */
    sidebarCount: () => number;
    sidebarLoading: () => boolean;
    onRefresh: () => void;
    hasSelection: () => boolean;
    detailLoading: () => boolean;
    detailError: () => string;
    onRetryDetail: () => void;
    detailTitle: () => string;
    detailMediaCount: () => number;
    mediaCards: () => MediaCardItem[];
    detailVersion: () => number;
    hasMore?: () => boolean;
    loadingMore?: () => boolean;
    onLoadMore?: () => void;
    onDownloadMediaList: (medias: MediaCardItem[], label: string) => Promise<void>;
}): JSXElement {
    const [selectedMediaIds, setSelectedMediaIds] = createSignal<number[]>([]);
    const selectedSet = createMemo(() => new Set(selectedMediaIds()));

    createEffect(() => {
        props.selectedSidebarId();
        props.detailVersion();
        setSelectedMediaIds([]);
    });

    const allSelected = () => {
        const cards = props.mediaCards();
        if (cards.length === 0) return false;
        const selected = selectedSet();
        return cards.every(card => selected.has(card.id));
    };

    const toggleSelectMedia = (id: number) => {
        setSelectedMediaIds(prev => {
            const next = new Set(prev);
            if (next.has(id)) next.delete(id);
            else next.add(id);
            return Array.from(next);
        });
    };

    const toggleSelectAllMedia = () => {
        const cards = props.mediaCards();
        setSelectedMediaIds(allSelected() ? [] : cards.map(card => card.id));
    };

    const downloadSelectedMedia = () => {
        const selected = selectedSet();
        const medias = props.mediaCards().filter(media => selected.has(media.id));
        return props.onDownloadMediaList(medias, `已选择 ${medias.length} 个视频`);
    };

    const downloadAllMedia = () => props.onDownloadMediaList(props.mediaCards(), '全部视频');

    return (
        <div class="flex h-full min-h-0 gap-3 overflow-hidden">
            <aside class="flex w-56 shrink-0 flex-col overflow-hidden rounded-xl border border-base-300 bg-base-100">
                <div
                    class="flex shrink-0 items-center justify-between gap-2 border-b border-base-200 bg-linear-to-b from-base-200/35 to-base-100/80 px-3 py-2"
                >
                    <div class="flex min-w-0 items-center gap-1.5">
                        <span class="select-none truncate text-xs font-bold tracking-wide text-base-content/70">
                            {props.sidebarLabel()}
                        </span>
                        <span
                            class="shrink-0 rounded-full bg-success/10 px-2 py-0.5 text-xs font-bold tabular-nums text-success"
                            title={`共 ${props.sidebarCount()} 个`}
                        >
                            {props.sidebarCount()}
                        </span>
                    </div>
                    <button
                        type="button"
                        class="inline-flex h-7 shrink-0 items-center gap-1.5 rounded-full border border-base-300/50 bg-base-100/90 px-2.5 text-xs font-medium text-base-content/80 shadow-sm backdrop-blur-sm transition-[color,background-color,border-color,box-shadow] hover:border-primary/30 hover:bg-primary/5 hover:text-primary hover:shadow disabled:cursor-not-allowed disabled:opacity-55"
                        onClick={props.onRefresh}
                        disabled={props.sidebarLoading()}
                    >
                        <Show when={props.sidebarLoading()} fallback={<IconRefresh class="h-3.5 w-3.5 opacity-90"/>}>
                            <span class="loading loading-spinner loading-xs text-primary"/>
                        </Show>
                        <span>{props.sidebarLoading() ? '刷新中' : '刷新'}</span>
                    </button>
                </div>

                <Switch>
                    <Match when={props.sidebarLoading()}>
                        <div class="flex flex-1 items-center justify-center py-12">
                            <span class="loading loading-spinner loading-sm text-primary"></span>
                        </div>
                    </Match>
                    <Match when={props.sidebarItems().length > 0}>
                        <SidebarList
                            list={props.sidebarItems}
                            selectedId={props.selectedSidebarId}
                            onSelect={props.onSelectSidebar}
                            icon={props.sidebarIcon}
                        />
                    </Match>
                    <Match when={true}>
                        <EmptyState title={props.emptySidebarTitle} compact/>
                    </Match>
                </Switch>
            </aside>

            <main class="flex min-w-0 flex-1 flex-col overflow-hidden rounded-xl border border-base-300 bg-base-100">
                <Show
                    when={props.hasSelection()}
                    fallback={
                        <EmptyState
                            title={props.emptyContentTitle}
                            description="右侧将展示视频列表"
                        />
                    }
                >
                    <Switch>
                        <Match when={props.detailLoading()}>
                            <DetailLoading/>
                        </Match>
                        <Match when={!!props.detailError()}>
                            <DetailError message={props.detailError()} onRetry={props.onRetryDetail}/>
                        </Match>
                        <Match when={props.mediaCards().length === 0}>
                            <EmptyState title="暂无视频" description="可以切换到其他项查看"/>
                        </Match>
                        <Match when={true}>
                            <DetailToolbar
                                title={props.detailTitle()}
                                mediaCount={props.detailMediaCount()}
                                selectedCount={selectedMediaIds().length}
                                allSelected={allSelected()}
                                onToggleSelectAll={toggleSelectAllMedia}
                                onClearSelection={() => setSelectedMediaIds([])}
                                onDownloadSelected={() => void downloadSelectedMedia()}
                                onDownloadAll={() => void downloadAllMedia()}
                            />
                            <div class="min-h-0 flex-1 overflow-auto p-4">
                                <VideoCardGrid
                                    medias={props.mediaCards()}
                                    selectedSet={selectedSet}
                                    onToggleSelect={toggleSelectMedia}
                                    onDownloadOne={(media) => void props.onDownloadMediaList([media], `视频 ${media.title}`)}
                                />
                                <Show when={props.hasMore?.()}>
                                    <div class="mt-4 flex justify-center pb-2">
                                        <button
                                            class="btn btn-outline btn-sm"
                                            onClick={props.onLoadMore}
                                            disabled={props.loadingMore?.()}
                                        >
                                            {props.loadingMore?.() ? '加载中...' : '加载剩余视频'}
                                        </button>
                                    </div>
                                </Show>
                            </div>
                        </Match>
                    </Switch>
                </Show>
            </main>
        </div>
    );
}
