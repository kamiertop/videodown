import {useNavigate} from "@tanstack/solid-router";
import {createEffect, createMemo, createSignal, For, type JSXElement, Match, Show, Switch,} from "solid-js";
import {createStore} from "solid-js/store";
import {Collection as FavoriteMixCollection, CollectionList, UserSeries} from "../../../wailsjs/go/api/Douyin";
import {model} from "../../../wailsjs/go/models";
import type {DouyinDownloadItem} from "../../lib/douyinStore.ts";
import {addDouyinVideos} from "../../lib/douyinStore.ts";
import {formatDate, formatDuration} from "../../lib/format.ts";
import {defaultDouyinVideoOption, douyinImageURLs, douyinVideoOptions, isDouyinImageAlbum} from "../../lib/douyinMedia.ts";
import DetailError from "../DetailError.tsx";
import DetailLoading from "../DetailLoading.tsx";
import EmptyState from "../EmptyState.tsx";
import type {DouyinVideoCardItem} from "./DouyinVideoCard.tsx";
import DouyinVideoGrid from "./DouyinVideoGrid.tsx";

const MIX_PAGE_SIZE = 12;
const MIX_VIDEO_PAGE_SIZE = 20;
const EMPTY_MIX_LIST: readonly DouyinMixItem[] = [];

type DouyinMixItem = model.CollectionItem | model.SeriesInfoItem;

function normalizeDouyinDuration(value?: number): number {
  if (!value || value <= 0) return 0;
  return value >= 1000 ? Math.floor(value / 1000) : value;
}

function awemeKey(item: model.AwemeItem, index: number): string {
  return item.aweme_id || item.group_id || item.sec_item_id || `${item.author_user_id || "item"}-${index}`;
}

function awemeCover(item: model.AwemeItem): string {
  return item.video?.cover?.url_list?.[0]
    ?? item.video?.origin_cover?.url_list?.[0]
    ?? "";
}

function awemeTitle(item: model.AwemeItem): string {
  return item.item_title || item.desc || item.caption || `作品 ${item.aweme_id || ""}`.trim();
}

function mixId(item: DouyinMixItem): string {
  return "series_id" in item ? item.series_id : item.mix_id;
}

function mixKey(item: DouyinMixItem, index: number): string {
  return mixId(item) || `mix-${index}`;
}

function mixCover(item: DouyinMixItem): string {
  return item.cover_url?.url_list?.[0] ?? "";
}

function mixTitle(item: DouyinMixItem): string {
  return "series_id" in item
    ? item.series_name || item.real_name
    : item.mix_name;
}

function mixEpisodeCount(item: DouyinMixItem | null | undefined): number {
  if (!item) return 0;
  if ("series_id" in item) {
    return item.stats?.updated_to_episode ?? item.stats?.total_episode ?? 0;
  }
  return item.statis?.updated_to_episode ?? item.statis?.has_updated_episode ?? 0;
}

export default function DouyinMixPanel(props: {
  active: boolean;
  showToast: (message: string, type?: "success" | "error" | "warning" | "info") => void;
  userSecUserId?: string;
  refreshKey?: unknown;
}): JSXElement {
  const navigate = useNavigate();

  // --- mix list state ---
  const [mixLoading, setMixLoading] = createSignal(false);
  const [mixError, setMixError] = createSignal("");
  const [mixes, setMixes] = createSignal<readonly DouyinMixItem[]>(EMPTY_MIX_LIST);
  const [mixCursor, setMixCursor] = createSignal(0);
  const [mixHasMore, setMixHasMore] = createSignal(false);
  const [mixLoadingMore, setMixLoadingMore] = createSignal(false);
  const [mixLoadedOnce, setMixLoadedOnce] = createSignal(false);

  // --- selected mix & video detail state ---
  const [selectedMix, setSelectedMix] = createSignal<DouyinMixItem | null>(null);
  const [detailLoading, setDetailLoading] = createSignal(false);
  const [detailError, setDetailError] = createSignal("");
  const [detail, setDetail] = createSignal<model.CollectionListResponse | null>(null);
  const [detailHasMore, setDetailHasMore] = createSignal(false);
  const [detailLoadingMore, setDetailLoadingMore] = createSignal(false);

  // --- video selection state ---
  const [allSelected, setAllSelected] = createSignal(false);
  const [selectedMap, setSelectedMap] = createStore<Record<string, true>>({});

  let mixRequestSeq = 0;
  let detailRequestSeq = 0;
  let loadedSourceKey = "";
  let observedRefreshKey: unknown = props.refreshKey;

  const sourceKey = () => props.userSecUserId ? `user:${props.userSecUserId}` : "favorite";
  const listTitle = () => props.userSecUserId ? "全部合集" : "收藏合集";
  const emptyTitle = () => props.userSecUserId ? "暂无合集" : "暂无收藏合集";

  const sidebarItems = () => mixes();

  const mixVideos = () => detail()?.aweme_list ?? [];

  // 合集详情与收藏视频使用同一张视频卡片，只在这里做接口数据到卡片模型的压平。
  const videoItems = createMemo<DouyinVideoCardItem[]>(() =>
    mixVideos().map((item, index): DouyinVideoCardItem => {
      const duration = normalizeDouyinDuration(item.video?.duration ?? item.duration ?? 0);
      const awemeId = item.aweme_id || item.group_id || item.sec_item_id || `${item.author_user_id}-${index}`;
      const title = awemeTitle(item);
      const cover = awemeCover(item);
      const author = item.author?.nickname || item.author?.uid || selectedMix()?.author?.nickname || "未知作者";
      const videoOptions = douyinVideoOptions(item);
      const selectedVideoOption = defaultDouyinVideoOption(videoOptions);
      const sourceName = selectedMix() ? mixTitle(selectedMix()!) : listTitle();
      return {
        id: awemeKey(item, index),
        cover,
        title,
        author,
        publishText: formatDate(item.create_time ?? 0),
        durationText: formatDuration(duration),
        downloadItem: {
          awemeId,
          sourceKind: props.userSecUserId ? "用户合集" : "收藏合集",
          sourceName,
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
    }),
  );

  // 合集页是左侧列表、右侧详情的固定布局；左侧分页只追加列表数据。
  function resetMixState(): void {
    setMixes(EMPTY_MIX_LIST);
    setMixCursor(0);
    setMixHasMore(false);
    setMixLoadingMore(false);
    setMixLoadedOnce(false);
    setSelectedMix(null);
    setDetail(null);
    setDetailError("");
    setDetailHasMore(false);
    clearSelection();
  }

  const loadMixes = async (append = false) => {
    if (append) {
      setMixLoadingMore(true);
    } else {
      setMixLoading(true);
      setMixError("");
      setMixLoadingMore(false);
      setMixHasMore(false);
    }

    const seq = ++mixRequestSeq;
    try {
      const cursor = append ? mixCursor() : 0;
      const data = props.userSecUserId
        ? await UserSeries(props.userSecUserId, cursor, MIX_PAGE_SIZE)
        : await FavoriteMixCollection(MIX_PAGE_SIZE, cursor);
      if (seq !== mixRequestSeq) return;
      const nextItems = props.userSecUserId
        ? (data as model.UserSeriesResponse).series_infos ?? []
        : (data as model.CollectionResponse).mix_infos ?? [];
      if (append) {
        setMixes([...sidebarItems(), ...nextItems]);
      } else {
        setMixes(nextItems);
      }
      setMixHasMore(Number(data.has_more ?? 0) > 0);
      setMixCursor(data.cursor ?? (append ? sidebarItems().length : nextItems.length));
      setMixLoadedOnce(true);
    } catch (error) {
      if (seq !== mixRequestSeq) return;
      const msg = error instanceof Error ? error.message : String(error);
      if (append) props.showToast(`加载更多合集失败: ${msg}`, "warning");
      else setMixError(msg);
    } finally {
      if (seq === mixRequestSeq) {
        if (append) setMixLoadingMore(false);
        else setMixLoading(false);
      }
    }
  };

  // 右侧详情切换合集时重置选择；加载更多时保留当前选择并追加视频。
  const loadMixDetail = async (item: DouyinMixItem, append = false) => {
    setSelectedMix(item);
    if (append) {
      setDetailLoadingMore(true);
    } else {
      setDetailLoadingMore(false);
      setDetailLoading(true);
      setDetailError("");
      setDetail(null);
      setDetailHasMore(false);
      clearSelection();
    }

    const seq = ++detailRequestSeq;
    try {
      const data = await CollectionList(
        props.userSecUserId || item.author?.sec_uid || "",
        mixId(item),
        append ? mixVideos().length : 0,
        MIX_VIDEO_PAGE_SIZE,
      );
      if (seq !== detailRequestSeq) return;
      const receivedCount = data.aweme_list?.length ?? 0;
      if (append) {
        const prev = mixVideos();
        const next = model.CollectionListResponse.createFrom({
          ...data,
          aweme_list: [...prev, ...(data.aweme_list ?? [])],
        });
        setDetail(next);
        setDetailHasMore(receivedCount > 0 && Number(next.has_more ?? 0) > 0);
      } else {
        setDetail(data);
        setDetailHasMore(receivedCount > 0 && Number(data.has_more ?? 0) > 0);
      }
    } catch (error) {
      if (seq !== detailRequestSeq) return;
      const msg = error instanceof Error ? error.message : String(error);
      if (append) props.showToast(`加载更多视频失败: ${msg}`, "warning");
      else setDetailError(msg);
    } finally {
      if (seq === detailRequestSeq) {
        if (append) setDetailLoadingMore(false);
        else setDetailLoading(false);
      }
    }
  };

  // 收藏页初始停在“收藏夹”，合集列表等用户切到该标签时再加载。
  createEffect(() => {
    if (!props.active) return;
    const currentSourceKey = sourceKey();
    if (loadedSourceKey !== currentSourceKey) {
      loadedSourceKey = currentSourceKey;
      resetMixState();
    }
    if (observedRefreshKey !== props.refreshKey) {
      observedRefreshKey = props.refreshKey;
      resetMixState();
    }
    if (!mixLoadedOnce()) {
      void loadMixes();
    }
  });

  const selectedKeys = createMemo(() => Object.keys(selectedMap));

  // 全选模式下 selectedMap 存“被排除项”；普通模式下存“被选中项”。
  const selectedCount = createMemo(() => {
    if (allSelected()) return Math.max(0, videoItems().length - selectedKeys().length);
    return selectedKeys().length;
  });

  function allVideosSelected(): boolean {
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

  function clearSelection(): void {
    setAllSelected(false);
    for (const id of Object.keys(selectedMap)) setSelectedMap(id, undefined!);
  }

  function toggleSelectAll(): void {
    if (allVideosSelected()) {
      clearSelection();
      return;
    }
    setAllSelected(true);
    for (const id of Object.keys(selectedMap)) setSelectedMap(id, undefined!);
  }

  function videoCardClass(id: string): string {
    if (allSelected()) {
      return selectedMap[id]
        ? "border-2 border-base-300 bg-base-100 opacity-70"
        : "border-2 border-transparent bg-base-100 ring-1 ring-base-300";
    }
    return selectedMap[id]
      ? "border-2 border-success bg-success/5 shadow-sm shadow-success/15"
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
    addDouyinVideos(items);
    clearSelection();
    await navigate({to: "/douyin/download"});
  }

  return (
    <div class={props.active ? "flex h-full min-h-0 flex-1" : "hidden"}>
      <Switch>
        <Match when={mixLoading()}>
          <DetailLoading/>
        </Match>
        <Match when={!!mixError()}>
          <DetailError message={mixError()} onRetry={() => void loadMixes()}/>
        </Match>
        <Match when={sidebarItems().length === 0}>
          <EmptyState title={emptyTitle()} description="请先确认账号已登录，或稍后重试。"/>
        </Match>
        <Match when={sidebarItems().length > 0}>
          {/* 左侧固定宽度目录，右侧详情区独立滚动，避免两块内容互相挤压。 */}
          <aside class="flex w-48 shrink-0 flex-col border-r border-base-300 bg-base-100">
            <div class="flex shrink-0 items-center justify-between gap-2 border-b border-base-200 px-3 py-2">
              <div class="min-w-0">
                <h3 class="truncate text-sm font-bold text-base-content">{listTitle()}</h3>
                <p class="text-xs text-base-content/50">{sidebarItems().length} 个已加载</p>
              </div>
            </div>
            <div class="min-h-0 flex-1 overflow-auto">
              <For each={sidebarItems()}>
                {(item, index) => {
                  const isSelected = () => {
                    const selected = selectedMix();
                    return !!selected && mixId(selected) === mixId(item);
                  };
                  return (
                    <button
                      type="button"
                      class={`flex w-full gap-3 border-b border-base-200/80 p-3 text-left transition-colors ${
                        isSelected() ? "bg-primary/10" : "hover:bg-base-200/35"
                      }`}
                      onClick={() => {
                        const selected = selectedMix();
                        if (selected && mixId(selected) === mixId(item) && detail() && !detailLoading()) return;
                        void loadMixDetail(item);
                      }}
                      data-key={mixKey(item, index())}
                    >
                      <div class="h-16 w-16 shrink-0 overflow-hidden rounded-xl border border-base-300 bg-base-200">
                        <Show
                          when={mixCover(item)}
                          fallback={<div
                            class="flex h-full items-center justify-center text-[0.65rem] font-bold text-base-content/35">无封面</div>}
                        >
                          <img
                            src={mixCover(item)}
                            alt=""
                            class="h-full w-full object-cover"
                            loading="lazy"
                            decoding="async"
                            referrerPolicy="no-referrer"
                          />
                        </Show>
                      </div>
                      <div class="min-w-0 flex-1 self-center">
                        <h4
                          class={`line-clamp-2 text-sm font-semibold ${isSelected() ? "text-primary" : "text-base-content"}`}>
                          {mixTitle(item) || "未命名合集"}
                        </h4>
                        <p class="mt-1 truncate text-xs text-base-content/50">
                          {item.author?.nickname || item.author?.uid || "未知作者"}
                        </p>
                        <p class="mt-1 text-xs text-base-content/45">
                          {mixEpisodeCount(item)} 个作品
                        </p>
                      </div>
                    </button>
                  );
                }}
              </For>
            </div>
            <Show when={mixHasMore()}>
              <div class="shrink-0 border-t border-base-200 p-2">
                <button
                  class="btn btn-ghost btn-sm h-9 w-full"
                  type="button"
                  onClick={() => void loadMixes(true)}
                  disabled={mixLoadingMore()}
                >
                  <Show
                    when={!mixLoadingMore()}
                    fallback={<span class="loading loading-spinner loading-sm text-primary"/>}
                  >
                    加载更多
                  </Show>
                </button>
              </div>
            </Show>
          </aside>

          {/* 详情区只挂载当前合集的视频网格，切换合集时重新请求并重置选择。 */}
          <main class="flex min-w-0 flex-1 flex-col overflow-hidden">
            <Show
              when={selectedMix()}
              fallback={<EmptyState title="选择一个合集" description="右侧将展示合集内视频列表。"/>}
            >
              <div class="flex h-full min-h-0 flex-col">
                <Switch>
                  <Match when={detailLoading()}>
                    <DetailLoading/>
                  </Match>
                  <Match when={!!detailError()}>
                    <DetailError message={detailError()} onRetry={() => {
                      const item = selectedMix();
                      if (item) void loadMixDetail(item);
                    }}/>
                  </Match>
                  <Match when={videoItems().length === 0}>
                    <div class="flex flex-1 items-center justify-center p-6">
                      <EmptyState title="暂无合集视频" description="该合集暂未返回可展示的视频。"/>
                    </div>
                  </Match>
                  <Match when={true}>
                    <DouyinVideoGrid
                      title="合集视频"
                      countText={`${videoItems().length} 个已加载`}
                      items={videoItems()}
                      selectedCount={selectedCount()}
                      allSelected={allVideosSelected()}
                      selectedClass={videoCardClass}
                      onToggleItem={toggleSelect}
                      onToggleAll={toggleSelectAll}
                      onClearSelection={clearSelection}
                      onDownloadSelected={() => void enqueueAndGoDownload(selectedDownloadItems())}
                      onDownloadAll={() => void enqueueAndGoDownload(videoItems().map(v => v.downloadItem))}
                      hasMore={detailHasMore()}
                      loadingMore={detailLoadingMore()}
                      onLoadMore={() => {
                        const item = selectedMix();
                        if (item) void loadMixDetail(item, true);
                      }}
                    />
                  </Match>
                </Switch>
              </div>
            </Show>
          </main>
        </Match>
      </Switch>
    </div>
  );
}
