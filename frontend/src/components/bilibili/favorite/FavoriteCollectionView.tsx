import {type JSXElement, Match, Show, Switch} from "solid-js";
import type {MediaCardItem} from "../../../lib/model.ts";
import DetailError from "../../DetailError";
import DetailLoading from "../../DetailLoading";
import EmptyState from "../../EmptyState";
import IconRefresh from "../../icons/IconRefresh";
import SidebarList, {type SidebarListItem} from "../../SidebarList";
import VideoListSection from "../../VideoListSection";

export type sidebarLabelType = "收藏夹" | "合集";

export default function FavoriteCollectionView<T extends SidebarListItem>(props: {
    sidebarItems: () => readonly T[];
    selectedSidebarId: () => number | null;
    onSelectSidebar: (item: T) => void;
    sidebarIcon: JSXElement;
    /** 侧栏左上角标题，如「收藏夹」「合集」 */
    sidebarLabel: sidebarLabelType;
    /** 对应列表数量 */
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
}): JSXElement {
    function emptySidebarTitle(): string {
        switch (props.sidebarLabel) {
            case "合集":
                return "暂无合集";
            case "收藏夹":
                return "暂无收藏夹";
        }
    }

    function emptyContentTitle(): string {
        switch (props.sidebarLabel) {
            case "合集":
                return "暂无合集内容";
            case "收藏夹":
                return "暂无收藏夹内容";
        }
    }

    return (
        <div class="flex h-full min-h-0 gap-3 overflow-hidden">
            <aside class="flex w-56 shrink-0 flex-col overflow-hidden rounded-xl border border-base-300 bg-base-100">
                <div
                    class="flex shrink-0 items-center justify-between gap-2 border-b border-base-200 bg-linear-to-b from-base-200/35 to-base-100/80 px-3 py-2"
                >
                    <div class="flex min-w-0 items-center gap-1.5">
                        <span class="select-none truncate text-xs font-bold tracking-wide text-base-content/70">
                            {props.sidebarLabel}
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
                        <EmptyState title={emptySidebarTitle()} compact/>
                    </Match>
                </Switch>
            </aside>

            <main class="flex min-w-0 flex-1 flex-col overflow-hidden rounded-xl border border-base-300 bg-base-100">
                <Show
                    when={props.hasSelection()}
                    fallback={
                        <EmptyState
                            title={emptyContentTitle()}
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
                            <VideoListSection
                                title={props.detailTitle()}
                                mediaCount={props.detailMediaCount()}
                                medias={props.mediaCards}
                                selectionResetKey={() => `${props.selectedSidebarId() ?? ""}-${props.detailVersion()}`}
                                hasMore={props.hasMore}
                                loadingMore={props.loadingMore}
                                onLoadMore={props.onLoadMore}
                            />
                        </Match>
                    </Switch>
                </Show>
            </main>
        </div>
    );
}
