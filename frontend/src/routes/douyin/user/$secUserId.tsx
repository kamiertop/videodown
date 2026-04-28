import {createFileRoute, Link, useNavigate} from '@tanstack/solid-router'
import {createMemo, createResource, createSignal, type JSXElement, Match, Show, Switch} from "solid-js";
import {createStore} from "solid-js/store";
import {User, UserVideoList} from "../../../../wailsjs/go/api/Douyin";
import {model} from "../../../../wailsjs/go/models";
import DetailError from "../../../components/DetailError.tsx";
import DetailLoading from "../../../components/DetailLoading.tsx";
import DouyinMixPanel from "../../../components/douyin/DouyinMixPanel.tsx";
import type {DouyinVideoCardItem} from "../../../components/douyin/DouyinVideoCard.tsx";
import DouyinVideoGrid from "../../../components/douyin/DouyinVideoGrid.tsx";
import EmptyState from "../../../components/EmptyState.tsx";
import IconRefresh from "../../../components/icons/IconRefresh.tsx";
import Toast from "../../../components/Toast.tsx";
import {useToast} from "../../../hooks/useToast.ts";
import {defaultDouyinVideoOption, douyinImageURLs, douyinVideoOptions, isDouyinImageAlbum} from "../../../lib/douyinMedia.ts";
import {addDouyinVideos, type DouyinDownloadItem} from "../../../lib/douyinStore.ts";
import {formatCount, formatDate, formatDuration} from "../../../lib/format.ts";

export const Route = createFileRoute('/douyin/user/$secUserId')({
  component: DouyinUserPage,
})

const USER_VIDEO_PAGE_SIZE = 20;

type UserTab = "video" | "series";

function avatarUrl(user: model.User | undefined): string {
  return user?.avatar_larger?.url_list?.[0]
    ?? user?.avatar_medium?.url_list?.[0]
    ?? user?.avatar_thumb?.url_list?.[0]
    ?? "";
}

function normalizeDouyinDuration(value?: number): number {
  if (!value || value <= 0) return 0;
  return value >= 1000 ? Math.floor(value / 1000) : value;
}

function awemeKey(item: model.AwemeItem, index: number): string {
  return item.aweme_id || item.group_id || item.sec_item_id || `${item.author_user_id || "item"}-${index}`;
}

function awemeCover(item: model.AwemeItem): string {
  return item.video.cover.url_list[0] || item.video.origin_cover.url_list[0] || "";
}

function awemeTitle(item: model.AwemeItem): string {
  return item.item_title || item.desc || item.caption || `作品 ${item.aweme_id || ""}`.trim();
}

function DouyinUserPage(): JSXElement {
  const params = Route.useParams();
  const secUserId = () => params().secUserId;
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = createSignal<UserTab>("video");
  const [loadingMore, setLoadingMore] = createSignal(false);
  const [seriesRefreshKey, setSeriesRefreshKey] = createSignal(0);
  const [allSelected, setAllSelected] = createSignal(false);
  const [selectedMap, setSelectedMap] = createStore<Record<string, true>>({});
  const {message, type, showToast} = useToast();
  const [userResult, {refetch: refetchUser}] = createResource(secUserId, User);
  const [videoResult, {refetch: refetchVideos, mutate: mutateVideos}] = createResource(
    secUserId,
    (id) => UserVideoList(id, USER_VIDEO_PAGE_SIZE, 0),
  );

  const user = () => userResult()?.user;
  const videos = () => videoResult()?.aweme_list ?? [];
  const videoItems = createMemo<DouyinVideoCardItem[]>(() =>
    videos().map((item, index): DouyinVideoCardItem => {
      const duration = normalizeDouyinDuration(item.video?.duration ?? item.duration ?? 0);
      const awemeId = item.aweme_id || item.group_id || item.sec_item_id || `${item.author_user_id}-${index}`;
      const title = awemeTitle(item);
      const cover = awemeCover(item);
      const authorName = item.author?.nickname || user()?.nickname || "未知作者";
      const videoOptions = douyinVideoOptions(item);
      const selectedVideoOption = defaultDouyinVideoOption(videoOptions);
      return {
        id: awemeKey(item, index),
        cover,
        title,
        author: authorName,
        publishText: formatDate(item.create_time ?? 0),
        durationText: formatDuration(duration),
        downloadItem: {
          awemeId,
          sourceKind: "用户作品",
          sourceName: user()?.nickname || authorName,
          title,
          cover,
          duration,
          authorName,
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
      }
    }),
  );
  const selectedKeys = createMemo(() => Object.keys(selectedMap));
  const selectedCount = createMemo(() => {
    if (allSelected()) return Math.max(0, videoItems().length - selectedKeys().length);
    return selectedKeys().length;
  });

  function isSelected(id: string): boolean {
    if (allSelected()) return !selectedMap[id];
    return selectedMap[id];
  }

  function allVideosSelected(): boolean {
    return videoItems().length > 0 && selectedCount() === videoItems().length;
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

  function cardClass(id: string): string {
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
    videoItems().filter(item => isSelected(item.id)).map(item => item.downloadItem),
  );

  async function enqueueAndGoDownload(items: DouyinDownloadItem[]): Promise<void> {
    if (items.length === 0) return;
    addDouyinVideos(items);
    clearSelection();
    await navigate({to: "/douyin/download"});
  }

  function hasMore(): boolean {
    return Number(videoResult()?.has_more ?? 0) > 0;
  }

  async function loadMore(): Promise<void> {
    const current = videoResult();
    if (!current || !hasMore() || loadingMore()) return;

    setLoadingMore(true);
    try {
      const next = await UserVideoList(secUserId(), USER_VIDEO_PAGE_SIZE, current.max_cursor ?? videos().length);
      mutateVideos(model.UserVideoListResponse.createFrom({
        ...next,
        aweme_list: [...videos(), ...(next.aweme_list ?? [])],
      }));
    } finally {
      setLoadingMore(false);
    }
  }

  async function reload(): Promise<void> {
    clearSelection();
    if (activeTab() === "series") {
      setSeriesRefreshKey((value) => value + 1);
      await refetchUser();
      return;
    }
    await Promise.all([refetchUser(), refetchVideos()]);
  }

  return (
    <section class="flex h-full min-h-0 flex-col gap-3 overflow-hidden bg-base-200/40 p-3">
      <header
        class="flex h-12 shrink-0 items-center gap-3 overflow-hidden rounded-xl border border-base-300 bg-base-100 px-4">
        <div class="flex min-w-0 flex-1 items-center gap-2">
          <Link to="/douyin/user" class="btn btn-ghost btn-sm shrink-0">返回用户</Link>
          <div class="h-5 w-px bg-base-300"></div>
          <h2 class="shrink-0 text-sm font-bold text-base-content">抖音用户</h2>
          <span class="min-w-0 truncate rounded-full bg-base-200 px-2 py-0.5 text-xs text-base-content/60">
            {user()?.signature || "全部作品"}
          </span>
        </div>

        <div class="flex shrink-0 items-center gap-2">
          <Switch>
            <Match when={userResult.loading}>
              <div class="flex items-center gap-2">
                <span class="loading loading-spinner loading-xs text-primary"></span>
                <span class="text-xs text-base-content/50">获取用户信息...</span>
              </div>
            </Match>
            <Match when={!userResult.loading}>
              <div class="flex min-w-0 items-center gap-2">
                <div class="h-8 w-8 shrink-0 overflow-hidden rounded-full bg-base-200 ring-2 ring-base-200">
                  <Show when={avatarUrl(user())} fallback={<div class="h-full w-full bg-base-200"/>}>
                    <img src={avatarUrl(user())} alt="" class="h-full w-full object-cover"
                         referrerPolicy="no-referrer"/>
                  </Show>
                </div>
                <div class="flex min-w-0 items-center gap-2">
                  <span class="max-w-56 truncate text-sm font-black text-base-content">
                    {user()?.nickname || "抖音用户"}
                  </span>
                  <span
                    class="badge badge-outline badge-sm">作品 {formatCount(user()?.aweme_count ?? videoItems().length)}</span>
                  <span class="badge badge-outline badge-sm">粉丝 {formatCount(user()?.follower_count ?? 0)}</span>
                  <Show when={user()?.ip_location}>
                    <span class="badge badge-ghost badge-sm">{user()?.ip_location}</span>
                  </Show>
                </div>
              </div>
            </Match>
          </Switch>
          <button class="btn btn-ghost btn-square btn-sm" onClick={() => void reload()} title="刷新">
            <IconRefresh
              class={`h-4 w-4 text-base-content/50 ${
                userResult.loading || videoResult.loading ? "animate-spin" : ""
              }`}/>
          </button>
        </div>
      </header>

      <main class="min-h-0 flex-1 overflow-hidden rounded-xl border border-base-300 bg-base-100">
        <div class="flex h-full min-h-0 flex-col">
          <nav class="grid shrink-0 grid-cols-2 border-b border-base-300 bg-base-100 p-1" role="tablist"
               aria-label="用户内容">
            <button
              class={`min-h-8 rounded-lg text-sm font-semibold ${
                activeTab() === "video" ? "bg-primary/12 text-primary ring-1 ring-primary/25" : "text-base-content/55 hover:bg-base-200"
              }`}
              type="button"
              role="tab"
              aria-selected={activeTab() === "video"}
              onClick={() => setActiveTab("video")}
            >
              视频
            </button>
            <button
              class={`min-h-8 rounded-lg text-sm font-semibold ${
                activeTab() === "series" ? "bg-primary/12 text-primary ring-1 ring-primary/25" : "text-base-content/55 hover:bg-base-200"
              }`}
              type="button"
              role="tab"
              aria-selected={activeTab() === "series"}
              onClick={() => {
                setActiveTab("series");
              }}
            >
              合集
            </button>
          </nav>

          <div class="min-h-0 flex-1 overflow-hidden">
            <Switch>
              {/*左侧视频网格*/}
              <Match when={activeTab() === "video"}>
                <Switch>
                  <Match when={userResult.loading || videoResult.loading}>
                    <DetailLoading/>
                  </Match>
                  <Match when={userResult.error || videoResult.error}>
                    <DetailError message={String(userResult.error || videoResult.error)} onRetry={() => void reload()}/>
                  </Match>
                  <Match when={videoItems().length === 0}>
                    <EmptyState title="暂无作品" description="该用户暂未返回可展示的视频。"/>
                  </Match>
                  <Match when={true}>
                    <DouyinVideoGrid
                      title="全部作品"
                      countText={`${videoItems().length} 个已加载`}
                      items={videoItems()}
                      selectedCount={selectedCount()}
                      allSelected={allVideosSelected()}
                      selectedClass={cardClass}
                      onToggleItem={toggleSelect}
                      onToggleAll={toggleSelectAll}
                      onClearSelection={clearSelection}
                      onDownloadSelected={() => void enqueueAndGoDownload(selectedDownloadItems())}
                      onDownloadAll={() => void enqueueAndGoDownload(videoItems().map(item => item.downloadItem))}
                      hasMore={hasMore()}
                      loadingMore={loadingMore()}
                      onLoadMore={() => void loadMore()}
                    />
                  </Match>
                </Switch>
              </Match>
              {/*右侧合集网格*/}
              <Match when={activeTab() === "series"}>
                <DouyinMixPanel
                  active={activeTab() === "series"}
                  userSecUserId={secUserId()}
                  refreshKey={seriesRefreshKey()}
                  showToast={showToast}
                />
              </Match>
            </Switch>
          </div>
        </div>
      </main>
      <Toast message={message()} type={type()}/>
    </section>
  );
}
