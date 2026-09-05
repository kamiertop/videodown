import {useNavigate} from "@tanstack/solid-router";
import {createEffect, createMemo, createSignal, type JSXElement, Match, Show, Switch} from "solid-js";
import {createStore} from "solid-js/store";
import * as model from "@bindings/github.com/kamiertop/videodown/douyin/model/models";
import {
  defaultDouyinVideoOption,
  douyinCoverCandidates,
  douyinDownloadAssets,
  douyinImageURLs,
  douyinMediaBadge,
  douyinMusicURL,
  douyinVideoOptions,
} from "../../lib/douyin/media.ts";
import {addDouyinVideos, type DouyinDownloadItem} from "../../lib/douyin/store.ts";
import {formatDate, formatDuration} from "../../lib/format.ts";
import DetailError from "../DetailError.tsx";
import DetailLoading from "../DetailLoading.tsx";
import EmptyState from "../EmptyState.tsx";
import VideoGrid, {type DouyinVideoCardItem} from "./VideoGrid.tsx";

function normalizeDouyinDuration(value?: number): number {
  // 抖音部分接口返回毫秒，部分字段可能已经是秒；展示层统一转成秒。
  if (!value || value <= 0) return 0;
  return value >= 1000 ? Math.floor(value / 1000) : value;
}

function awemeKey(item: model.AwemeItem, index: number): string {
  // aweme_id 最稳定；其他 ID 只用于接口缺字段时兜底，避免列表 key 为空。
  return item.aweme_id || item.group_id || item.sec_item_id || `${item.author_user_id || "item"}-${index}`;
}

function awemeCover(item: model.AwemeItem): string {
  return awemeCoverCandidates(item)[0] ?? "";
}

function awemeCoverCandidates(item: model.AwemeItem): string[] {
  const candidates = [
    ...(item.video?.raw_cover?.url_list ?? []),
    ...(item.video?.cover?.url_list ?? []),
    ...(item.video?.origin_cover?.url_list ?? []),
    ...(item.images?.[0]?.url_list ?? []),
  ];
  return [...new Set(candidates.filter(Boolean))];
}

function awemeTitle(item: model.AwemeItem): string {
  return item.item_title || item.desc || item.caption || `作品 ${item.aweme_id || ""}`.trim();
}

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
  /** 一键下载前自动加载完剩余分页；完成后使用最新 items 组建下载队列。 */
  prepareDownloadAll?: (onStatus?: (message: string) => void) => Promise<void>;
}): JSXElement {
  const navigate = useNavigate();
  // allSelected=true 时表示“当前已加载视频默认都选中”，selectedMap 存的是排除项。
  // allSelected=false 时，selectedMap 存的是用户逐个点选的视频 id。
  const [allSelected, setAllSelected] = createSignal(false);
  // createStore 方便按 id 删除单个选择项；值固定为 true，不存额外数据。
  const [selectedMap, setSelectedMap] = createStore<Record<string, true>>({});
  const [downloadAllLoading, setDownloadAllLoading] = createSignal(false);
  const [downloadAllStatus, setDownloadAllStatus] = createSignal("正在加载全部视频");

  function clearSelection(): void {
    setAllSelected(false);
    for (const id of Object.keys(selectedMap)) setSelectedMap(id, undefined!);
  }

  const videoItems = createMemo<DouyinVideoCardItem[]>(() =>
      props.items.map((item, index): DouyinVideoCardItem => {
        // 后端 AwemeItem 很大，网格只需要轻量视图模型和下载任务。
        const duration = normalizeDouyinDuration(item.video?.duration ?? item.duration ?? 0);
        const awemeId = item.aweme_id || item.group_id || item.sec_item_id || `${item.author_user_id}-${index}`;
        const title = awemeTitle(item);
        const cover = awemeCover(item);
        const author = item.author?.nickname || item.author?.uid || props.fallbackAuthor;
        // 清晰度选项在进入下载页前就解析好，下载页只需要让用户切换最终 URL。
        const videoOptions = douyinVideoOptions(item);
        const selectedVideoOption = defaultDouyinVideoOption(videoOptions);
        const imageURLs = douyinImageURLs(item);
        const mediaBadge = douyinMediaBadge(item);

        return {
          id: awemeKey(item, index),
          cover,
          coverCandidates: awemeCoverCandidates(item),
          title,
          author,
          publishText: formatDate(item.create_time ?? 0),
          durationText: formatDuration(duration),
          isTop: item.is_top === 1,
          downloadItem: {
            awemeId,
            sourceName: props.sourceName,
            title,
            cover,
            coverCandidates: douyinCoverCandidates(item),
            duration,
            authorName: author,
            publishTime: item.create_time ?? 0,
            diggCount: item.statistics?.digg_count ?? 0,
            collectCount: item.statistics?.collect_count ?? 0,
            link: awemeId ? `https://www.douyin.com/video/${awemeId}` : undefined,
            videoURL: selectedVideoOption?.url,
            videoOptions,
            selectedVideoOptionId: selectedVideoOption?.id,
            imageURLs,
            assets: mediaBadge ? douyinDownloadAssets(item) : undefined,
            musicURL: mediaBadge ? douyinMusicURL(item) : undefined,
            mediaBadge,
          },
          mediaBadge,
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
              onDownloadAll={async () => {
                if (downloadAllLoading()) return;
                setDownloadAllLoading(true);
                try {
                  setDownloadAllStatus("正在加载全部视频");
                  await props.prepareDownloadAll?.(setDownloadAllStatus);
                  await enqueueAndGoDownload(videoItems().map((item) => item.downloadItem));
                } finally {
                  setDownloadAllLoading(false);
                }
              }}
              refreshing={props.refreshing}
              onRefresh={props.onRefresh}
              hasMore={props.hasMore}
              loadingMore={props.loadingMore}
              onLoadMore={props.onLoadMore}
          />
          <Show when={downloadAllLoading()}>
            <div class="fixed inset-0 z-40 grid place-items-center bg-base-300/45 p-4 backdrop-blur-[2px]">
              <div class="w-full max-w-sm rounded-xl border border-base-300 bg-base-100 p-5 text-center shadow-2xl">
                <h3 class="text-base font-semibold text-base-content">{downloadAllStatus()}</h3>
                <p class="mt-1 text-xs text-base-content/60">请稍候，完成后将自动加入下载队列</p>
                <progress class="progress progress-primary mt-4 w-full" />
                <p class="mt-2 text-xs tabular-nums text-base-content/55">已加载 {videoItems().length} 个视频</p>
              </div>
            </div>
          </Show>
        </Match>
      </Switch>
  );
}
