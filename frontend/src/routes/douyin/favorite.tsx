import {createFileRoute, useNavigate} from '@tanstack/solid-router'
import {
  createEffect,
  createMemo,
  createResource,
  createSignal,
  For,
  Index,
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
import EmptyState from "../../components/EmptyState.tsx";
import IconRefresh from "../../components/icons/IconRefresh.tsx";
import Toast from "../../components/Toast.tsx";
import {useToast} from "../../hooks/useToast.ts";
import {addDouyinVideos, type DouyinDownloadItem} from "../../lib/douyinStore.ts";
import {formatCount, formatDate, formatDuration} from "../../lib/format.ts";

export const Route = createFileRoute('/douyin/favorite')({
  component: DouyinFavoritePage,
})

type FavoriteTab = 'collection' | 'video' | 'music' | 'mix';

type VideoListItemView = {
  id: string;
  awemeId: string;
  cover: string;
  title: string;
  author: string;
  publishText: string;
  durationText: string;
  diggText: string;
  collectText: string;
  downloadItem: DouyinDownloadItem;
};

function normalizeDouyinDuration(value?: number): number {
  if (!value || value <= 0) {
    return 0;
  }
  // 抖音这类接口的 duration 实际上常见是毫秒，这里兼容秒/毫秒两种情况。
  return value >= 1000 ? Math.floor(value / 1000) : value;
}

function awemeKey(item: model.AwemeItem): string {
  return item.aweme_id || item.group_id || item.sec_item_id || String(item.author_user_id || "");
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
    music: "音乐",
    mix: "合集",
  };
  return labels[tab];
}

function tabHint(tab: FavoriteTab): string {
  const hints: Record<FavoriteTab, string> = {
    collection: "已同步的公开收藏夹",
    video: "已收藏的作品",
    music: "需后端接口",
    mix: "需后端接口",
  };
  return hints[tab];
}

function DouyinFavoritePage(): JSXElement {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = createSignal<FavoriteTab>('collection');
  const [collectionLoadingMore, setCollectionLoadingMore] = createSignal(false);
  const [videoLoadingMore, setVideoLoadingMore] = createSignal(false);
  const [videoEnabled, setVideoEnabled] = createSignal(false);
  const [allVideoSelected, setAllVideoSelected] = createSignal(false);
  const [selectedVideoMap, setSelectedVideoMap] = createStore<Record<string, true>>({});
  const [collectionReloadKey] = createSignal(0);
  const [videoReloadKey] = createSignal(0);
  const {message, type, showToast} = useToast();

  // 收藏夹首屏直接拉取；视频列表仅在切到“视频”标签后再触发，避免无意义请求。
  const [collectionResult, {refetch: refetchCollections, mutate: mutateCollections}] = createResource(
    collectionReloadKey,
    async () => CollectList(0),
  );
  const [videoResult, {refetch: refetchVideos, mutate: mutateVideos}] = createResource(
    () => videoEnabled() ? videoReloadKey() : null,
    async (key) => {
      if (key === null) {
        return undefined;
      }
      return FavoriteVideo(20, 0);
    },
  );

  const collections = () => collectionResult()?.collects_list ?? [];
  const videos = () => videoResult()?.aweme_list ?? [];
  // 将后端返回的深层对象提前压成轻量视图模型，渲染层只消费字符串。
  const videoItems = createMemo<VideoListItemView[]>(() =>
    videos().map((item, index) => {
      const duration = normalizeDouyinDuration(item.video?.duration ?? 0);
      const awemeId = item.aweme_id || item.group_id || item.sec_item_id || `${item.author_user_id}-${index}`;
      return {
        id: awemeKey(item),
        awemeId,
        cover: item.video?.cover?.url_list?.[0] ?? item.video?.origin_cover?.url_list?.[0],
        title: item.item_title || item.desc || item.caption || `作品 ${item.aweme_id}`,
        author: item.author?.nickname || item.author?.uid || "—",
        publishText: formatDate(item.create_time ?? 0),
        durationText: formatDuration(duration),
        diggText: formatCount(item.statistics?.digg_count ?? 0),
        collectText: formatCount(item.statistics?.collect_count ?? 0),
        downloadItem: {
          awemeId,
          title: item.item_title || item.desc || item.caption || `作品 ${item.aweme_id}`,
          cover: item.video?.cover?.url_list?.[0] ?? item.video?.origin_cover?.url_list?.[0] ?? "",
          duration,
          authorName: item.author?.nickname || item.author?.uid || "—",
          publishTime: item.create_time ?? 0,
          diggCount: item.statistics?.digg_count ?? 0,
          collectCount: item.statistics?.collect_count ?? 0,
          link: awemeId ? `https://www.douyin.com/video/${awemeId}` : undefined,
        },
      }
    }),
  );
  const selectedVideoKeys = createMemo(() => Object.keys(selectedVideoMap));
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
    return !!selectedVideoMap[id];
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
        if (prev === 'music') return 'video';
        if (prev === 'mix') return 'music';
        return 'collection';
      });
      return;
    }

    if (event.key === 'ArrowRight') {
      event.preventDefault();
      setActiveTab((prev) => {
        if (prev === 'collection') return 'video';
        if (prev === 'video') return 'music';
        if (prev === 'music') return 'mix';
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

  function refreshCurrentTab(): void {
    const tab = activeTab();
    if (tab === 'collection') {
      void refetchCollections();
      return;
    }
    if (tab === 'video') {
      if (!videoEnabled()) {
        setVideoEnabled(true);
        return;
      }
      void refetchVideos();
      return;
    }
    showToast('当前标签暂未接入后端接口', 'info');
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
    <section class="flex h-full min-h-0 flex-col p-4">
      <header class="shrink-0">
        <nav
          class="mb-3 grid grid-cols-4 gap-1 rounded-2xl border border-base-300/90 bg-linear-to-b from-base-100 via-base-100 to-base-200/20 p-1.5 shadow-sm"
          role="tablist"
          aria-label="收藏分类"
        >
          <For each={['collection', 'video', 'music', 'mix'] as const}>
            {(tab) => (
              <button
                class={`min-h-10 flex-1 rounded-xl px-2 text-center text-xs font-semibold sm:text-sm ${
                  activeTab() === tab
                    ? "bg-primary/12 text-primary shadow-sm ring-1 ring-primary/25"
                    : "text-base-content/50 hover:bg-base-200/80 hover:text-base-content"
                } ${tab === 'music' || tab === 'mix' ? "opacity-90" : ""}`}
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

        <div
          class="mb-2 flex min-h-16 items-center justify-between gap-3 rounded-xl border border-base-300 bg-linear-to-r from-base-100 to-base-100/90 px-3 py-3 shadow-sm sm:px-4">
          <div class="min-w-0">
            <div class="flex items-center gap-2">
              <h2 class="truncate text-base font-bold text-base-content">{tabLabel(activeTab())}</h2>
              <Show when={activeTab() === 'collection' && collections().length > 0}>
                <span
                  class="rounded-full bg-base-200 px-2 py-0.5 text-xs font-semibold tabular-nums text-base-content/65">
                  {collections().length}
                </span>
              </Show>
            </div>
            <p class="mt-1 text-xs text-base-content/50 sm:text-sm">
              {tabHint(activeTab())}
            </p>
          </div>
          <button
            class="btn btn-ghost btn-square btn-sm shrink-0 rounded-xl border border-transparent hover:border-base-300 hover:bg-base-200/70"
            type="button"
            title="刷新"
            onClick={refreshCurrentTab}
          >
            <IconRefresh class={`h-4 w-4 text-base-content/50 ${
              (activeTab() === 'collection' && collectionResult.loading) || (activeTab() === 'video' && videoResult.loading)
                ? "animate-spin"
                : ""
            }`}
            />
          </button>
        </div>
      </header>

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
                  <div class="flex min-h-0 flex-1 flex-col">
                    <div class="flex shrink-0 items-center gap-2 border-b border-base-300 px-4 py-3">
                      <div class="min-w-0 flex-1">
                        <h3 class="truncate text-sm font-bold text-base-content">收藏视频</h3>
                        <p class="text-xs text-base-content/55">{videoItems().length} 个视频</p>
                      </div>
                      <button class="btn btn-ghost btn-sm" onClick={toggleSelectAllVideos}>
                        {allVideosSelected() ? "取消全选" : "全选"}
                      </button>
                      <Show when={selectedVideoCount() > 0}>
                        <button class="btn btn-ghost btn-sm text-error" onClick={clearSelectedVideos}>
                          取消已选
                        </button>
                      </Show>
                      <button
                        class="btn btn-outline btn-primary btn-sm"
                        onClick={() => void enqueueAndGoDownload(selectedDownloadItems())}
                        disabled={selectedVideoCount() === 0}
                      >
                        去下载 ({selectedVideoCount()})
                      </button>
                    </div>

                    <div class="min-h-0 flex-1 overflow-auto">
                      <div
                        class="grid grid-cols-2 gap-2 p-2 sm:grid-cols-3 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7 2xl:grid-cols-8">
                        <Index each={videoItems()}>
                          {(item) => (
                            <article
                              class={`flex h-full cursor-pointer flex-col overflow-hidden rounded-lg transition-colors ${
                                isVideoSelected(item().id)
                                  ? "border-2 border-success bg-success/5 shadow-sm shadow-success/15"
                                  : "border-2 border-transparent bg-base-100 ring-1 ring-base-300"
                              }`}
                              style={{"content-visibility": "auto", "contain-intrinsic-size": "260px"}}
                              onClick={() => toggleSelectVideo(item().id)}
                            >
                              <div
                                class="relative w-full overflow-hidden bg-base-200 aspect-3/5">
                                <Show
                                  when={item().cover}
                                  fallback={<div
                                    class="absolute inset-0 flex items-center justify-center text-xs text-base-content/35">无封面</div>}
                                >
                                  <img
                                    src={item().cover}
                                    class="h-full w-full object-cover"
                                    alt=""
                                    loading="lazy"
                                    decoding="async"
                                    fetchpriority="low"
                                    referrerPolicy="no-referrer"
                                  />
                                </Show>
                                <span
                                  class="absolute bottom-1 right-1 max-w-[calc(100%-0.5rem)] rounded-md bg-black/65 px-1.5 py-1 text-[0.65rem] font-medium tabular-nums leading-none text-white">
                                {item().durationText}
                              </span>
                              </div>
                              <div class="flex flex-1 flex-col p-2">
                                <h3 class="line-clamp-2 text-[12px] font-semibold leading-4.5 text-base-content">
                                  {item().title}
                                </h3>
                                <p class="mt-1 line-clamp-1 text-[10px] text-base-content/50">@{item().author}</p>
                                <p class="mt-1.5 text-[10px] text-base-content/45">
                                  发布 {item().publishText}
                                </p>
                                <div class="mt-auto flex items-center gap-1 pt-1.5 text-[10px] text-base-content/55">
                                  <span class="bg-base-200 px-1.5 py-0.5 tabular-nums">赞 {item().diggText}</span>
                                  <span class="bg-base-200 px-1.5 py-0.5 tabular-nums">收藏 {item().collectText}</span>
                                </div>
                              </div>
                            </article>
                          )}
                        </Index>
                      </div>
                      <Show when={videoHasMore()}>
                        <div class="border-t border-base-200/90 p-2">
                          <button
                            class="btn btn-ghost btn-sm h-9 w-full"
                            type="button"
                            onClick={() => void loadMoreVideos()}
                            disabled={videoLoadingMore()}
                          >
                            加载更多
                          </button>
                        </div>
                      </Show>
                    </div>

                  </div>
                </Match>
              </Switch>
            </Match>

            <Match when={activeTab() === 'music'}>
              <div class="p-6 sm:p-10">
                <EmptyState title="收藏音乐待接入" description="后端暂未提供音乐收藏接口，这里先保留标签位。"/>
              </div>
            </Match>

            <Match when={activeTab() === 'mix'}>
              <div class="p-6 sm:p-10">
                <EmptyState title="收藏合集待接入" description="后端暂未提供合集收藏接口，这里先保留标签位。"/>
              </div>
            </Match>
          </Switch>
        </div>
      </div>
      <Toast message={message()} type={type()}/>
    </section>
  );
}
