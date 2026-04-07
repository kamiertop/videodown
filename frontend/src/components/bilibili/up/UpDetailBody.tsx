import {Match, Show, Switch, type JSXElement} from "solid-js";
import type {MediaCardItem} from "../../VideoCardGrid";
import DetailError from "../../DetailError";
import DetailLoading from "../../DetailLoading";
import DetailToolbar from "../../DetailToolbar";
import EmptyState from "../../EmptyState";
import SidebarList from "../../SidebarList";
import VideoCardGrid from "../../VideoCardGrid";
import IconRefresh from "../../icons/IconRefresh";
import IconBook from "../../icons/IconBook";

export type UpTab = 'videos' | 'lists';

/**
 * Bilibili UP 详情页主体（Tab + 内容区）：
 * - 只负责渲染 UI，不直接发请求
 * - 所有数据/回调由路由页传入，便于拆分与测试
 *
 * 布局/滚动策略：
 * - 本组件内部也使用 `flex + overflow-hidden`，并在内容区使用 `overflow-auto`
 * - 目的：无论是“全部视频”还是“合集|系列”，都只让右侧内容区滚动，避免整页滚动抖动
 */
export default function UpDetailBody(props: {
    // Tab
    activeTab: () => UpTab;
    setActiveTab: (tab: UpTab) => void;

    // 视频 Tab
    videoLoading: () => boolean;
    videoError: () => string;
    videoCards: () => MediaCardItem[];
    videoTotal: () => number;
    selectedMediaIds: () => number[];
    allSelected: () => boolean;
    hasMoreVideos: () => boolean;
    videoLoadingMore: () => boolean;
    selectedSet: () => Set<number>;
    onToggleSelectAll: () => void;
    onClearSelection: () => void;
    onDownloadSelected: () => void;
    onDownloadAll: () => void;
    onToggleSelect: (id: number) => void;
    onDownloadOne: (m: MediaCardItem) => void;
    onRetryVideo: () => void;
    onLoadMoreVideos: () => void;

    // 合集/系列 Tab（Sidebar + 列表）
    ssLoadedOnce: () => boolean;
    ssLoading: () => boolean;
    ssError: () => string;
    // 注意：SidebarList 需要函数形式的 list；这里保持 route 侧的 memo 原样透传
    listSidebarItems: () => any[];
    selectedListId: () => number | null;
    listDetailLoading: () => boolean;
    listDetailError: () => string;
    listCards: () => MediaCardItem[];
    listTotal: () => number;
    listLoadingMore: () => boolean;
    hasMoreListVideos: () => boolean;
    selectedListItem: () => {subtitle: string; title: string; count?: number} | null;
    onEnsureLoadSeasonsSeries: () => void;
    onRefreshSeasonsSeries: () => void;
    onSelectListItem: (item: any) => void;
    onRetryListDetail: () => void;
    onLoadMoreList: () => void;
}): JSXElement {
    return (
        <section class="flex h-full min-h-0 gap-3 overflow-hidden bg-base-200/40 p-3">
            <main class="flex min-w-0 flex-1 flex-col overflow-hidden rounded-xl border border-base-300 bg-base-100">
                {/* Tab 切换（全部视频 / 合集|系列） */}
                <div class="flex shrink-0 border-b border-base-300">
                    <button
                        class={`flex-1 py-3 text-center text-sm font-bold transition-colors ${
                            props.activeTab() === 'videos'
                                ? 'border-b-2 border-success text-success'
                                : 'text-base-content/60 hover:text-base-content'
                        }`}
                        onClick={() => props.setActiveTab('videos')}
                    >
                        全部视频
                        <span class="ml-1 text-xs font-normal tabular-nums">
                            {props.videoTotal() || props.videoCards().length}
                        </span>
                    </button>
                    <button
                        class={`flex-1 py-3 text-center text-sm font-bold transition-colors ${
                            props.activeTab() === 'lists'
                                ? 'border-b-2 border-success text-success'
                                : 'text-base-content/60 hover:text-base-content'
                        }`}
                        onClick={() => {
                            props.setActiveTab('lists');
                            // 懒加载：首次切换到“合集|系列”时再去拉取 sidebar 数据，避免首屏请求过多。
                            props.onEnsureLoadSeasonsSeries();
                        }}
                    >
                        合集 | 系列
                        <span class="ml-1 text-xs font-normal tabular-nums">{props.listSidebarItems().length}</span>
                    </button>
                </div>

                <Switch>
                    {/* 全部视频 */}
                    <Match when={props.activeTab() === 'videos'}>
                        <Switch>
                            <Match when={props.videoLoading()}>
                                <DetailLoading/>
                            </Match>
                            <Match when={!!props.videoError()}>
                                <DetailError message={props.videoError()} onRetry={props.onRetryVideo}/>
                            </Match>
                            <Match when={!props.videoLoading() && props.videoCards().length === 0}>
                                <EmptyState title="暂无视频" description="该 UP 主暂无投稿视频或接口返回为空"/>
                            </Match>
                            <Match when={true}>
                                <DetailToolbar
                                    title="全部投稿视频"
                                    mediaCount={props.videoTotal() || props.videoCards().length}
                                    selectedCount={props.selectedMediaIds().length}
                                    allSelected={props.allSelected()}
                                    onToggleSelectAll={props.onToggleSelectAll}
                                    onClearSelection={props.onClearSelection}
                                    onDownloadSelected={props.onDownloadSelected}
                                    onDownloadAll={props.onDownloadAll}
                                />
                                <div class="min-h-0 flex-1 overflow-auto p-4">
                                    <VideoCardGrid
                                        medias={props.videoCards()}
                                        selectedSet={props.selectedSet}
                                        onToggleSelect={props.onToggleSelect}
                                        onDownloadOne={m => props.onDownloadOne(m)}
                                    />
                                    <Show when={props.hasMoreVideos()}>
                                        <div class="mt-4 flex justify-center pb-2">
                                            <button
                                                class="btn btn-outline btn-sm"
                                                onClick={props.onLoadMoreVideos}
                                                disabled={props.videoLoadingMore()}
                                            >
                                                {props.videoLoadingMore() ? '加载中...' : '加载更多视频'}
                                            </button>
                                        </div>
                                    </Show>
                                </div>
                            </Match>
                        </Switch>
                    </Match>

                    {/* 合集 | 系列 */}
                    <Match when={props.activeTab() === 'lists'}>
                        <div class="flex min-h-0 flex-1 overflow-hidden">
                            {/* 左侧 sidebar（合集/系列列表） */}
                            <aside class="flex w-64 shrink-0 flex-col overflow-hidden border-r border-base-300 bg-base-100">
                                <div class="flex shrink-0 items-center justify-between border-b border-base-200 px-3 py-2">
                                    <span class="text-xs font-bold text-base-content/70">列表</span>
                                    <button
                                        class="flex h-7 w-7 shrink-0 items-center justify-center rounded transition-colors hover:bg-base-200 disabled:cursor-not-allowed"
                                        onClick={props.onRefreshSeasonsSeries}
                                        disabled={props.ssLoading()}
                                        title="刷新"
                                    >
                                        <IconRefresh
                                            class={`h-3.5 w-3.5 text-base-content/50 ${props.ssLoading() ? 'animate-spin' : ''}`}
                                        />
                                    </button>
                                </div>
                                <Switch>
                                    <Match when={props.ssLoading()}>
                                        <div class="flex flex-1 items-center justify-center py-12">
                                            <span class="loading loading-spinner loading-sm text-primary"></span>
                                        </div>
                                    </Match>
                                    <Match when={!!props.ssError()}>
                                        <div class="p-3">
                                            <DetailError message={props.ssError()} onRetry={props.onRefreshSeasonsSeries}/>
                                        </div>
                                    </Match>
                                    <Match when={props.listSidebarItems().length === 0}>
                                        <EmptyState title="暂无合集/系列" compact/>
                                    </Match>
                                    <Match when={true}>
                                        <SidebarList
                                            list={() => props.listSidebarItems() as any}
                                            selectedId={props.selectedListId}
                                            onSelect={props.onSelectListItem}
                                            icon={() => (
                                                <IconBook class="h-3 w-3"/>
                                            )}
                                        />
                                    </Match>
                                </Switch>
                            </aside>

                            {/* 右侧视频内容区（当前选中的合集/系列的视频） */}
                            <div class="flex min-w-0 flex-1 flex-col overflow-hidden">
                                <Show
                                    when={!!props.selectedListItem()}
                                    fallback={<EmptyState title="选择一个合集/系列查看内容" description="右侧将展示视频列表"/>}
                                >
                                    <Switch>
                                        <Match when={props.listDetailLoading()}>
                                            <DetailLoading/>
                                        </Match>
                                        <Match when={!!props.listDetailError()}>
                                            <DetailError message={props.listDetailError()} onRetry={props.onRetryListDetail}/>
                                        </Match>
                                        <Match when={!props.listDetailLoading() && props.listCards().length === 0}>
                                            <EmptyState title="暂无视频" description="该合集/系列暂无可用视频或接口返回为空"/>
                                        </Match>
                                        <Match when={true}>
                                            <DetailToolbar
                                                title={`${props.selectedListItem()!.subtitle}: ${props.selectedListItem()!.title}`}
                                                mediaCount={props.listTotal() || props.selectedListItem()!.count || props.listCards().length}
                                                selectedCount={props.selectedMediaIds().length}
                                                allSelected={props.allSelected()}
                                                onToggleSelectAll={props.onToggleSelectAll}
                                                onClearSelection={props.onClearSelection}
                                                onDownloadSelected={props.onDownloadSelected}
                                                onDownloadAll={props.onDownloadAll}
                                            />
                                            <div class="min-h-0 flex-1 overflow-auto p-4">
                                                <VideoCardGrid
                                                    medias={props.listCards()}
                                                    selectedSet={props.selectedSet}
                                                    onToggleSelect={props.onToggleSelect}
                                                    onDownloadOne={m => props.onDownloadOne(m)}
                                                />
                                                <Show when={props.hasMoreListVideos()}>
                                                    <div class="mt-4 flex justify-center pb-2">
                                                        <button
                                                            class="btn btn-outline btn-sm"
                                                            onClick={props.onLoadMoreList}
                                                            disabled={props.listLoadingMore()}
                                                        >
                                                            {props.listLoadingMore() ? '加载中...' : '加载更多视频'}
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
    );
}

