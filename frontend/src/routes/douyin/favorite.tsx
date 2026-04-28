import {createFileRoute, useNavigate} from '@tanstack/solid-router'
import {
  createEffect,
  createMemo,
  createResource,
  createSignal,
  For,
  type JSXElement,
  Match,
  onCleanup,
  onMount,
  Show,
  Switch,
} from "solid-js";
import {createStore} from "solid-js/store";
import {CollectList, FavoriteVideo} from "../../../wailsjs/go/api/Douyin";
import {model} from "../../../wailsjs/go/models";
import DetailError from "../../components/DetailError.tsx";
import DetailLoading from "../../components/DetailLoading.tsx";
import DouyinMixPanel from "../../components/douyin/DouyinMixPanel.tsx";
import type {DouyinVideoCardItem} from "../../components/douyin/DouyinVideoCard.tsx";
import DouyinVideoGrid from "../../components/douyin/DouyinVideoGrid.tsx";
import EmptyState from "../../components/EmptyState.tsx";
import Toast from "../../components/Toast.tsx";
import {useToast} from "../../hooks/useToast.ts";
import {defaultDouyinVideoOption, douyinImageURLs, douyinVideoOptions, isDouyinImageAlbum} from "../../lib/douyinMedia.ts";
import {addDouyinVideos, type DouyinDownloadItem} from "../../lib/douyinStore.ts";
import {formatCount, formatDate, formatDuration} from "../../lib/format.ts";

export const Route = createFileRoute('/douyin/favorite')({
  component: DouyinFavoritePage,
})

type FavoriteTab = 'collection' | 'video' | 'mix';

function normalizeDouyinDuration(value?: number): number {
  if (!value || value <= 0) {
    return 0;
  }
  // 抖音这类接口的 duration 实际上常见是毫秒，这里兼容秒/毫秒两种情况。
  return value >= 1000 ? Math.floor(value / 1000) : value;
}

function awemeKey(item: model.AwemeItem, index: number): string {
  return item.aweme_id || item.group_id || item.sec_item_id || `${item.author_user_id || "item"}-${index}`;
}

function awemeCover(item: model.AwemeItem): string {
  return item.video?.cover?.url_list?.[0]
    ?? item.video?.origin_cover?.url_list?.[0]
    ?? item.images?.[0]?.url_list?.[0]
    ?? "";
}

function awemeTitle(item: model.AwemeItem): string {
  return item.item_title || item.desc || item.caption || `作品 ${item.aweme_id || ""}`.trim();
}

function collectionKey(item: model.CollectsList, index: number): string {
  if (item.collects_id_str) {
    return item.collects_id_str;
  }
  if (item.collects_id) {
    return `c${item.collects_id}`;
  }
  return `i${index}`;
}

function tabLabel(tab: FavoriteTab): string {
  const labels: Record<FavoriteTab, string> = {
    collection: "收藏夹",
    video: "视频",
    mix: "合集",
  };
  return labels[tab];
}

function DouyinFavoritePage(): JSXElement {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = createSignal<FavoriteTab>('collection');
  const [collectionLoadingMore, setCollectionLoadingMore] = createSignal(false);
  const [videoLoadingMore, setVideoLoadingMore] = createSignal(false);
  const [videoEnabled, setVideoEnabled] = createSignal(false);
  const [allVideoSelected, setAllVideoSelected] = createSignal(false);
  const [selectedVideoMap, setSelectedVideoMap] = createStore<Record<string, true>>({});
  const {message, type, showToast} = useToast();

  // 收藏夹首屏直接拉取；视频列表仅在切到“视频”标签后再触发，避免无意义请求。
  const [collectionResult, {refetch: refetchCollections, mutate: mutateCollections}] = createResource(
    async () => CollectList(0),
  );
  const [videoResult, {refetch: refetchVideos, mutate: mutateVideos}] = createResource(
    () => videoEnabled(),
    async (enabled) => {
      if (!enabled) return undefined;
      return FavoriteVideo(20, 0);
    },
  );

  const collections = () => collectionResult()?.collects_list ?? [];
  const videos = () => videoResult()?.aweme_list ?? [];

  // 将后端返回的深层对象提前压成轻量视图模型，渲染层只消费字符串。
  function toVideoItem(item: model.AwemeItem, index: number, fallbackAuthor = "—"): DouyinVideoCardItem {
    const duration = normalizeDouyinDuration(item.video?.duration ?? item.duration ?? 0);
    const awemeId = item.aweme_id || item.group_id || item.sec_item_id || `${item.author_user_id}-${index}`;
    const title = awemeTitle(item);
    const cover = awemeCover(item);
    const author = item.author?.nickname || item.author?.uid || fallbackAuthor;
    const videoOptions = douyinVideoOptions(item);
    const selectedVideoOption = defaultDouyinVideoOption(videoOptions);
    return {
      id: awemeKey(item, index),
      cover,
      title,
      author,
      publishText: formatDate(item.create_time ?? 0),
      durationText: formatDuration(duration),
      downloadItem: {
        awemeId,
        sourceKind: "收藏视频",
        sourceName: "收藏视频",
        title,
        cover,
        duration,
        authorName: author,
        publishTime: item.create_time ?? 0,
        diggCount: item.statistics?.digg_count ?? 0,
        collectCount: item.statistics?.collect_count ?? 0,
        link: awemeId ? `https://www.douyin.com/video/${awemeId}` : undefined,
        videoURL: selectedVideoOption?.url,
        videoOptions,
        selectedVideoOptionId: selectedVideoOption?.id,
        imageURLs: douyinImageURLs(item),
      },
      showImgLabel: isDouyinImageAlbum(item),
    };
  }

  const videoItems = createMemo<DouyinVideoCardItem[]>(() =>
    videos().map((item, index): DouyinVideoCardItem => toVideoItem(item, index)),
  );
  const selectedVideoKeys = createMemo(() => Object.keys(selectedVideoMap));

  // 全选模式下 selectedVideoMap 存“被排除项”；普通模式下存“被选中项”。
  const selectedVideoCount = createMemo(() => {
    if (allVideoSelected()) {
      return Math.max(0, videoItems().length - selectedVideoKeys().length);
    }
    return selectedVideoKeys().length;
  });

  function allVideosSelected(): boolean {
    return videoItems().length > 0 && selectedVideoCount() === videoItems().length;
  }

  function isVideoSelected(id: string): boolean {
    if (allVideoSelected()) {
      return !selectedVideoMap[id];
    }
    return selectedVideoMap[id];
  }

  function videoCardClass(id: string): string {
    if (allVideoSelected()) {
      return selectedVideoMap[id]
        ? "border-2 border-base-300 bg-base-100 opacity-70"
        : "border-2 border-transparent bg-base-100 ring-1 ring-base-300";
    }
    return selectedVideoMap[id]
      ? "border-2 border-success bg-success/5 shadow-sm shadow-success/15"
      : "border-2 border-transparent bg-base-100 ring-1 ring-base-300";
  }

  createEffect(() => {
    const validIds = new Set(videoItems().map((item) => item.id));
    for (const id of Object.keys(selectedVideoMap)) {
      if (!validIds.has(id)) {
        setSelectedVideoMap(id, undefined!);
      }
    }
  });

  function switchTab(event: KeyboardEvent): void {
    if (event.key === 'ArrowLeft') {
      event.preventDefault();
      setActiveTab((prev) => {
        if (prev === 'video') return 'collection';
        if (prev === 'mix') return 'video';
        return 'collection';
      });
      return;
    }

    if (event.key === 'ArrowRight') {
      event.preventDefault();
      setActiveTab((prev) => {
        if (prev === 'collection') return 'video';
        if (prev === 'video') return 'mix';
        return 'mix';
      });
    }
  }

  onMount(() => {
    window.addEventListener("keydown", switchTab);
  });
  onCleanup(() => {
    window.removeEventListener("keydown", switchTab);
  });

  createEffect(() => {
    if (activeTab() === 'video' && !videoEnabled()) {
      setVideoEnabled(true);
    }
  });

  function toggleSelectVideo(id: string): void {
    if (selectedVideoMap[id]) {
      setSelectedVideoMap(id, undefined!);
      return;
    }
    setSelectedVideoMap(id, true);
  }

  function clearSelectedVideos(): void {
    setAllVideoSelected(false);
    for (const id of Object.keys(selectedVideoMap)) {
      setSelectedVideoMap(id, undefined!);
    }
  }

  function toggleSelectAllVideos(): void {
    if (allVideosSelected()) {
      setAllVideoSelected(false);
      for (const id of Object.keys(selectedVideoMap)) {
        setSelectedVideoMap(id, undefined!);
      }
      return;
    }
    setAllVideoSelected(true);
    for (const id of Object.keys(selectedVideoMap)) {
      setSelectedVideoMap(id, undefined!);
    }
  }

  const selectedDownloadItems = createMemo(() => {
    return videoItems()
      .filter((item) => isVideoSelected(item.id))
      .map((item) => item.downloadItem);
  });

  async function enqueueAndGoDownload(items: DouyinDownloadItem[]): Promise<void> {
    if (items.length === 0) {
      showToast("请先选择要加入下载页的视频", "info");
      return;
    }
    addDouyinVideos(items);
    setAllVideoSelected(false);
    for (const id of Object.keys(selectedVideoMap)) {
      setSelectedVideoMap(id, undefined!);
    }
    await navigate({to: "/douyin/download"});
  }

  function collectionHasMore(): boolean {
    return collectionResult()?.has_more ?? false;
  }

  function videoHasMore(): boolean {
    return Number(videoResult()?.has_more ?? 0) > 0;
  }

  async function loadMoreCollections(): Promise<void> {
    const current = collectionResult();
    if (!current || !current.has_more || collectionLoadingMore()) {
      return;
    }

    setCollectionLoadingMore(true);
    try {
      const next = await CollectList(current.cursor ?? collections().length);
      // 追加分页时直接在当前资源上拼接，避免再引入一套额外 signal。
      mutateCollections(model.CollectListResponse.createFrom({
        ...next,
        collects_list: [...collections(), ...(next.collects_list ?? [])],
        total_number: current.total_number ?? next.total_number,
      }));
    } catch (error) {
      showToast(error instanceof Error ? error.message : String(error), 'warning');
    } finally {
      setCollectionLoadingMore(false);
    }
  }

  async function loadMoreVideos(): Promise<void> {
    const current = videoResult();
    if (!current || Number(current.has_more ?? 0) <= 0 || videoLoadingMore()) {
      return;
    }

    setVideoLoadingMore(true);
    try {
      const next = await FavoriteVideo(20, current.cursor ?? videos().length);
      mutateVideos(model.FavoriteVideoResponse.createFrom({
        ...next,
        aweme_list: [...videos(), ...(next.aweme_list ?? [])],
        uid: current.uid ?? next.uid,
        sec_uid: current.sec_uid ?? next.sec_uid,
      }));
    } catch (error) {
      showToast(error instanceof Error ? error.message : String(error), 'warning');
    } finally {
      setVideoLoadingMore(false);
    }
  }

  return (
    <section class="flex h-full min-h-0 flex-col p-2">
      <header class="shrink-0">
        <nav
          class="mb-2 flex flex-row justify-between rounded-2xl border border-base-300/90 bg-linear-to-b from-base-100 via-base-100 to-base-200/20 p-1.5 shadow-sm"
          role="tablist"
          aria-label="收藏分类"
        >
          <For each={['collection', 'video', 'mix'] as const}>
            {(tab) => (
              <button
                class={`min-h-6 flex-1 rounded-xl px-2 text-center text-xs font-semibold sm:text-sm ${
                  activeTab() === tab
                    ? "bg-primary/12 text-primary shadow-sm ring-1 ring-primary/25"
                    : "text-base-content/50 hover:bg-base-200/80 hover:text-base-content"
                } ${tab === 'mix' ? "opacity-90" : ""}`}
                type="button"
                role="tab"
                aria-selected={activeTab() === tab}
                onClick={() => setActiveTab(tab)}
              >
                {tabLabel(tab)}
              </button>
            )}
          </For>
        </nav>
      </header>

      {/* 外层只负责标签和边框，具体视频网格统一交给 DouyinVideoGrid 渲染。 */}
      <div class="min-h-0 flex-1 overflow-hidden rounded-xl border border-base-300 bg-base-100 shadow-sm">
        <div class="flex h-full min-h-0 flex-col">
          <Switch>
            <Match when={activeTab() === 'collection'}>
              <Switch>
                <Match when={collectionResult.loading}>
                  <DetailLoading/>
                </Match>
                <Match when={collectionResult.error}>
                  <DetailError message={String(collectionResult.error)} onRetry={() => void refetchCollections()}/>
                </Match>
                <Match when={collections().length === 0}>
                  <EmptyState title="暂无收藏夹" description="请先确认账号已登录且收藏夹可见。"/>
                </Match>
                <Match when={collections().length > 0}>
                  <div class="min-h-0 flex-1 overflow-auto">
                    <div class="divide-y divide-base-200/90">
                      <For each={collections()}>
                        {(item, index) => (
                          <article
                            class="flex gap-3 p-3 transition-colors hover:bg-base-200/20 sm:gap-4 sm:p-4"
                            data-key={collectionKey(item, index())}
                          >
                            <div
                              class="relative h-18 w-16 shrink-0 overflow-hidden rounded-xl border border-base-300/70 bg-base-200 sm:h-20 sm:w-20">
                              <Show
                                when={item.collects_cover?.url_list?.[0]}
                                fallback={<div
                                  class="flex h-full items-center justify-center text-[0.65rem] font-bold text-base-content/35">无封面</div>}
                              >
                                <img
                                  src={item.collects_cover.url_list[0]}
                                  alt=""
                                  class="h-full w-full object-cover"
                                  loading="lazy"
                                  decoding="async"
                                  referrerPolicy="no-referrer"
                                />
                              </Show>
                            </div>
                            <div class="min-w-0 flex-1 self-center">
                              <h3 class="line-clamp-2 text-sm font-semibold text-base-content sm:leading-snug">
                                {item.collects_name || "未命名收藏夹"}
                              </h3>
                              <p class="mt-1 line-clamp-1 text-xs text-base-content/50">
                                {item.user_info?.nickname || item.user_id_str || "未知用户"}
                              </p>
                              <div class="mt-2 flex flex-wrap gap-2 text-xs text-base-content/55">
                                <span class="rounded-full bg-base-200 px-2 py-1 tabular-nums">
                                  {item.total_number ?? 0} 个作品
                                </span>
                                <span class="rounded-full bg-base-200 px-2 py-1 tabular-nums">
                                  {formatCount(item.play_count ?? 0)} 次播放
                                </span>
                                <Show when={!item.is_normal_status}>
                                  <span
                                    class="inline-flex items-center rounded-full bg-warning/12 px-2 py-1 text-warning">
                                    状态异常
                                  </span>
                                </Show>
                              </div>
                            </div>
                          </article>
                        )}
                      </For>
                    </div>
                    <Show when={collectionHasMore()}>
                      <div class="border-t border-base-200/90 p-2">
                        <button
                          class="btn btn-ghost btn-sm h-9 w-full"
                          type="button"
                          onClick={() => void loadMoreCollections()}
                          disabled={collectionLoadingMore()}
                        >
                          <Show
                            when={!collectionLoadingMore()}
                            fallback={<span class="loading loading-spinner loading-sm text-primary"/>}
                          >
                            加载更多
                          </Show>
                        </button>
                      </div>
                    </Show>
                  </div>
                </Match>
              </Switch>
            </Match>

            <Match when={activeTab() === 'video'}>
              <Switch>
                <Match when={videoResult.loading}>
                  <DetailLoading/>
                </Match>
                <Match when={videoResult.error}>
                  <DetailError message={String(videoResult.error)} onRetry={() => void refetchVideos()}/>
                </Match>
                <Match when={videoItems().length === 0}>
                  <EmptyState title="暂无收藏视频" description="请先确认账号已登录，或稍后重试。"/>
                </Match>
                <Match when={videoItems().length > 0}>
                  <DouyinVideoGrid
                    title="收藏视频"
                    countText={`${videoItems().length} 个视频`}
                    items={videoItems()}
                    selectedCount={selectedVideoCount()}
                    allSelected={allVideosSelected()}
                    selectedClass={videoCardClass}
                    onToggleItem={toggleSelectVideo}
                    onToggleAll={toggleSelectAllVideos}
                    onClearSelection={clearSelectedVideos}
                    onDownloadSelected={() => void enqueueAndGoDownload(selectedDownloadItems())}
                    hasMore={videoHasMore()}
                    loadingMore={videoLoadingMore()}
                    onLoadMore={() => void loadMoreVideos()}
                  />
                </Match>
              </Switch>
            </Match>
            <Match when={activeTab() === 'mix'}>
              <DouyinMixPanel active={activeTab() === 'mix'} showToast={showToast}/>
            </Match>
          </Switch>
        </div>
      </div>
      <Toast message={message()} type={type()}/>
    </section>
  );
}
