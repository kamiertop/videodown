import {useNavigate} from "@tanstack/solid-router";
import {createEffect, createMemo, createSignal, type JSXElement, Match, Switch} from "solid-js";
import {createStore} from "solid-js/store";
import * as model from "@bindings/github.com/kamiertop/videodown/douyin/model/models";
import {awemeCoverURLCandidates, awemeToDownloadItem} from "../../lib/douyin/aweme.ts";
import {type DouyinBatchPageLoader, startDouyinBatch} from "../../lib/douyin/batchDownload.ts";
import {addDouyinVideos, type DouyinDownloadItem} from "../../lib/douyin/store.ts";
import {formatDate, formatDuration} from "../../lib/format.ts";
import DetailError from "../DetailError.tsx";
import DetailLoading from "../DetailLoading.tsx";
import EmptyState from "../EmptyState.tsx";
import VideoGrid, {type DouyinVideoCardItem} from "./VideoGrid.tsx";

export type DouyinVideoContentKind =
    | "favorite-video"
    | "user-video"
    | "favorite-collection"
    | "favorite-mix"
    | "user-mix"
    | "follow-dynamic";

// kind 是这个组件的核心入口：页面类型确定后，默认标题、空态、下载来源都可推导。
function defaultTitle(kind: DouyinVideoContentKind): string {
  switch (kind) {
    case "favorite-video":
      return "收藏视频";
    case "user-video":
      return "全部作品";
    case "follow-dynamic":
      return "关注动态";
    case "favorite-collection":
      return "收藏夹视频";
    case "favorite-mix":
    case "user-mix":
      return "合集视频";
  }
}

function emptyTitle(kind: DouyinVideoContentKind): string {
  switch (kind) {
    case "favorite-video":
      return "暂无收藏视频";
    case "user-video":
      return "暂无作品";
    case "follow-dynamic":
      return "暂无关注动态";
    case "favorite-collection":
      return "暂无收藏夹视频";
    case "favorite-mix":
    case "user-mix":
      return "暂无合集视频";
  }
}

function emptyDescription(kind: DouyinVideoContentKind): string {
  switch (kind) {
    case "favorite-video":
      return "请先确认账号已登录，或稍后重试。";
    case "user-video":
      return "该用户暂未返回可展示的视频。";
    case "follow-dynamic":
      return "关注的用户暂未发布新作品。";
    case "favorite-collection":
      return "该收藏夹暂未返回可展示的视频。";
    case "favorite-mix":
    case "user-mix":
      return "该合集暂未返回可展示的视频。";
  }
}

// 统一的视频内容面板：收藏视频、用户作品、收藏夹详情和合集详情都使用这一套选择/下载/分页 UI。
export default function VideoContentPanel(props: {
  kind: DouyinVideoContentKind;
  // 外层负责拉数据；本组件只负责展示、选择和加入下载队列。
  loading: boolean;
  error?: string;
  onRetry?: () => void;
  title?: string;
  items: readonly model.AwemeItem[];
  /** 后端返回的总作品数（用户详情页使用 aweme_count）。 */
  totalCount?: number;
  // sourceName 是具体来源名，例如某个合集名或用户昵称。
  sourceName: string;
  fallbackAuthor: string;
  showToast: (message: string, type?: "success" | "error" | "warning" | "info") => void;
  refreshing?: boolean;
  onRefresh?: () => void;
  hasMore?: boolean;
  loadingMore?: boolean;
  onLoadMore?: () => void;
  /**
   * 一键下载全部的批量会话描述。传入后点击按钮会立即跳转下载页并自动开始：
   * 当前已加载内容作为首批，后续分页由 loader 在会话里继续加载。
   * 不传时按钮退化为“把已加载内容入队并跳转”。
   */
  batchDownload?: {
    title: string;
    totalCount?: number;
    createLoader: () => DouyinBatchPageLoader;
  };
}): JSXElement {
  const navigate = useNavigate();
  // allSelected=true 时表示“当前已加载视频默认都选中”，selectedMap 存的是排除项。
  // allSelected=false 时，selectedMap 存的是用户逐个点选的视频 id。
  const [allSelected, setAllSelected] = createSignal(false);
  // createStore 方便按 id 删除单个选择项；值固定为 true，不存额外数据。
  const [selectedMap, setSelectedMap] = createStore<Record<string, true>>({});

  function clearSelection(): void {
    setAllSelected(false);
    for (const id of Object.keys(selectedMap)) setSelectedMap(id, undefined!);
  }

  const videoItems = createMemo<DouyinVideoCardItem[]>(() =>
      props.items.map((item, index): DouyinVideoCardItem => {
        // 后端 AwemeItem 很大，网格只需要轻量视图模型和下载任务；
        // 清晰度等解析统一走 awemeToDownloadItem，与链接解析、批量翻页保持一致。
        const downloadItem = awemeToDownloadItem(item, props.sourceName, props.fallbackAuthor, index);

        return {
          id: downloadItem.awemeId,
          cover: downloadItem.cover,
          coverCandidates: awemeCoverURLCandidates(item),
          title: downloadItem.title,
          author: downloadItem.authorName,
          publishText: formatDate(item.create_time ?? 0),
          durationText: formatDuration(downloadItem.duration),
          isTop: item.is_top === 1,
          downloadItem,
          mediaBadge: downloadItem.mediaBadge,
        };
      }),
  );

  // 翻页或刷新后剔除已经不存在的选择项，避免把旧页面的视频加入下载队列。
  createEffect(() => {
    const validIds = new Set(videoItems().map((item) => item.id));
    for (const id of Object.keys(selectedMap)) {
      if (!validIds.has(id)) setSelectedMap(id, undefined!);
    }
  });

  const selectedKeys = createMemo(() => Object.keys(selectedMap));

  // 全选模式下 selectedMap 记录“排除项”；普通模式下记录“选中项”，这样大量视频全选时不用写满所有 id。
  const selectedCount = createMemo(() => {
    if (allSelected()) return Math.max(0, videoItems().length - selectedKeys().length);
    return selectedKeys().length;
  });

  function allVideosSelected(): boolean {
    // 只判断当前已加载数据；还没点“加载更多”的下一页不会隐式加入选择。
    return videoItems().length > 0 && selectedCount() === videoItems().length;
  }

  function isVideoSelected(id: string): boolean {
    if (allSelected()) return !selectedMap[id];
    return selectedMap[id];
  }

  function toggleSelect(id: string): void {
    if (selectedMap[id]) setSelectedMap(id, undefined!);
    else setSelectedMap(id, true);
  }

  function toggleSelectAll(): void {
    if (allVideosSelected()) {
      clearSelection();
      return;
    }
    // 不逐个写入所有 id，靠 allSelected + 排除项表达全选，列表很长时更轻。
    setAllSelected(true);
    for (const id of Object.keys(selectedMap)) setSelectedMap(id, undefined!);
  }

  function videoCardClass(id: string): string {
    // 全选模式下 selectedMap 表示排除项，因此排除项显示成未选/弱化样式。
    if (allSelected()) {
      return selectedMap[id]
          ? "border-2 border-base-300 bg-base-100 opacity-70"
          : "border-2 border-primary bg-primary/5 shadow-sm shadow-primary/15";
    }
    return selectedMap[id]
        ? "border-2 border-primary bg-primary/5 shadow-sm shadow-primary/15"
        : "border-2 border-transparent bg-base-100 ring-1 ring-base-300";
  }

  const selectedDownloadItems = createMemo(() =>
      videoItems().filter((item) => isVideoSelected(item.id)).map((item) => item.downloadItem),
  );

  async function enqueueAndGoDownload(items: DouyinDownloadItem[]): Promise<void> {
    if (items.length === 0) {
      props.showToast("请先选择要加入下载页的视频", "info");
      return;
    }
    // 下载页使用全局 store 跨路由承接任务，加入后直接跳转过去。
    addDouyinVideos(items);
    clearSelection();
    await navigate({to: "/douyin/download"});
  }

  // 一键下载全部：立即跳转下载页并自动开始；后续分页由批量会话在下载页继续加载。
  // 只有配置了 batchDownload（可翻页的来源）才渲染按钮；关注动态等无翻页来源不提供。
  async function startBatchDownload(): Promise<void> {
    const batch = props.batchDownload;
    if (!batch) return;

    const started = startDouyinBatch({
      title: batch.title,
      totalCount: batch.totalCount,
      initialItems: videoItems().map((item) => item.downloadItem),
      loader: batch.createLoader(),
    });
    if (!started) {
      props.showToast("已有下载任务进行中，请稍后再试", "warning");
      return;
    }
    clearSelection();
    await navigate({to: "/douyin/download"});
  }

  return (
      <Switch>
        <Match when={props.loading}>
          {/* 外层正在拉第一页数据时，视频区域整体显示加载态。 */}
          <DetailLoading/>
        </Match>
        <Match when={props.error}>
          {/* 错误来源可能是收藏视频、用户作品或集合详情接口，由外层传入 retry。 */}
          <DetailError message={props.error!} onRetry={props.onRetry ?? (() => undefined)}/>
        </Match>
        <Match when={videoItems().length === 0}>
          {/* 空态文案从 kind 推导，调用方不用传重复字符串*/}
          <div class="flex h-full min-h-0 flex-1 items-center justify-center p-6">
            <EmptyState title={emptyTitle(props.kind)} description={emptyDescription(props.kind)}/>
          </div>
        </Match>
        <Match when={true}>
          {/* 具体网格只接收轻量卡片模型和选择回调，不知道后端 AwemeItem。 */}
          <VideoGrid
              title={props.title ?? defaultTitle(props.kind)}
              items={videoItems()}
              totalCount={props.totalCount}
              selectedCount={selectedCount()}
              allSelected={allVideosSelected()}
              selectedClass={videoCardClass}
              onToggleItem={toggleSelect}
              onToggleAll={toggleSelectAll}
              onClearSelection={clearSelection}
              onDownloadSelected={() => void enqueueAndGoDownload(selectedDownloadItems())}
              onDownloadAll={props.batchDownload ? () => void startBatchDownload() : undefined}
              refreshing={props.refreshing}
              onRefresh={props.onRefresh}
              hasMore={props.hasMore}
              loadingMore={props.loadingMore}
              onLoadMore={props.onLoadMore}
          />
        </Match>
      </Switch>
  );
}
