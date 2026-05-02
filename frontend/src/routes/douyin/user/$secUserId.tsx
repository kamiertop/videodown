import {createFileRoute, Link} from '@tanstack/solid-router'
import {createResource, createSignal, type JSXElement, Match, Show, Switch} from "solid-js";
import {CollectionList, User, UserSeries, UserVideoList} from "../../../../wailsjs/go/api/Douyin";
import {model} from "../../../../wailsjs/go/models";
import CollectionVideoPanel, {
  type DouyinListItem,
  type ListPage,
  type VideoPage,
} from "../../../components/douyin/CollectionVideoPanel.tsx";
import VideoContentPanel from "../../../components/douyin/VideoContentPanel.tsx";
import Toast from "../../../components/Toast.tsx";
import {useToast} from "../../../hooks/useToast.ts";
import {formatCount} from "../../../lib/format.ts";

export const Route = createFileRoute('/douyin/user/$secUserId')({
  component: DouyinUserPage,
})

const USER_VIDEO_PAGE_SIZE = 16;
const USER_MIX_PAGE_SIZE = 12;
const USER_MIX_VIDEO_PAGE_SIZE = 20;

type UserTab = "video" | "series";
type DouyinMixItem = model.CollectionItem | model.SeriesInfoItem;

function isUserSeries(item: DouyinMixItem): item is model.SeriesInfoItem {
  return "series_id" in item;
}

function mixId(item: DouyinMixItem): string {
  return isUserSeries(item) ? item.series_id : item.mix_id;
}

async function loadUserMixes(secUserId: string, cursor: number): Promise<ListPage> {
  const data = await UserSeries(secUserId, cursor, USER_MIX_PAGE_SIZE);
  const items = data.series_infos ?? [];
  return {
    items,
    cursor: data.cursor ?? cursor + items.length,
    hasMore: Number(data.has_more ?? 0) > 0,
  };
}

async function loadUserMixVideos(secUserId: string, item: DouyinListItem, cursor: number): Promise<VideoPage> {
  const mixItem = item as DouyinMixItem;
  const seriesID = mixId(mixItem);
  if (!seriesID) throw new Error("合集 ID 无效");

  const data = await CollectionList(secUserId || mixItem.author?.sec_uid || "", seriesID, cursor, USER_MIX_VIDEO_PAGE_SIZE);
  return {
    items: data.aweme_list ?? [],
    hasMore: Number(data.has_more ?? 0) > 0,
  };
}

function avatarUrl(user: model.User | undefined): string {
  return user?.avatar_larger?.url_list?.[0]
    ?? user?.avatar_medium?.url_list?.[0]
    ?? user?.avatar_thumb?.url_list?.[0]
    ?? "";
}

function UserHeader(props: {
  user: model.User | undefined;
  videoCount: number;
  loading: boolean;
}): JSXElement {
  return (
    <header
      class="flex h-12 shrink-0 items-center gap-3 overflow-hidden rounded-xl border border-base-300 bg-base-100 px-4">
      <div class="flex min-w-0 flex-1 items-center gap-2">
        <Link to="/douyin/user" class="btn btn-ghost btn-sm shrink-0">返回用户</Link>
        <div class="h-5 w-px bg-base-300"></div>
        <h2 class="shrink-0 text-sm font-bold text-base-content">抖音用户</h2>
        <span class="min-w-0 truncate rounded-full bg-base-200 px-2 py-0.5 text-xs text-base-content/60">
          {props.user?.signature || "全部作品"}
        </span>
      </div>

      <div class="flex shrink-0 items-center gap-2">
        <Switch>
          <Match when={props.loading}>
            <div class="flex items-center gap-2">
              <span class="loading loading-spinner loading-xs text-primary"></span>
              <span class="text-xs text-base-content/50">获取用户信息...</span>
            </div>
          </Match>
          <Match when={!props.loading}>
            <div class="flex min-w-0 items-center gap-2">
              <div class="h-8 w-8 shrink-0 overflow-hidden rounded-full bg-base-200 ring-2 ring-base-200">
                <Show when={avatarUrl(props.user)} fallback={<div class="h-full w-full bg-base-200"/>}>
                  <img src={avatarUrl(props.user)} alt="" class="h-full w-full object-cover"
                       referrerPolicy="no-referrer"/>
                </Show>
              </div>
              <div class="flex min-w-0 items-center gap-2">
                <span class="max-w-56 truncate text-sm font-black text-base-content">
                  {props.user?.nickname || "抖音用户"}
                </span>
                <span
                  class="badge badge-outline badge-sm">作品 {formatCount(props.user?.aweme_count ?? props.videoCount)}</span>
                <span class="badge badge-outline badge-sm">粉丝 {formatCount(props.user?.follower_count ?? 0)}</span>
                <Show when={props.user?.ip_location}>
                  <span class="badge badge-ghost badge-sm">{props.user?.ip_location}</span>
                </Show>
              </div>
            </div>
          </Match>
        </Switch>
      </div>
    </header>
  );
}

function DouyinUserPage(): JSXElement {
  const params = Route.useParams();
  const secUserId = () => params().secUserId;
  const [activeTab, setActiveTab] = createSignal<UserTab>("video");
  const [loadingMore, setLoadingMore] = createSignal(false);
  const [seriesRefreshKey, setSeriesRefreshKey] = createSignal(0);
  const {message, type, showToast} = useToast();
  const [userResult, {refetch: refetchUser}] = createResource(secUserId, User);
  const [videoResult, {refetch: refetchVideos, mutate: mutateVideos}] = createResource(
    secUserId,
    (id) => UserVideoList(id, USER_VIDEO_PAGE_SIZE, 0),
  );

  const user = () => userResult()?.user;
  const videos = () => videoResult()?.aweme_list ?? [];

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
    if (activeTab() === "series") {
      setSeriesRefreshKey((value) => value + 1);
      await refetchUser();
      return;
    }
    await Promise.all([refetchUser(), refetchVideos()]);
  }

  return (
    <section class="flex h-full min-h-0 flex-col gap-3 overflow-hidden bg-base-200/40 p-3">
      <UserHeader
        user={user()}
        videoCount={videos().length}
        loading={userResult.loading}
      />

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
            <div classList={{
              "flex h-full min-h-0 w-full min-w-0 flex-1 flex-col overflow-hidden": activeTab() === "video",
              "hidden": activeTab() !== "video",
            }}>
              <VideoContentPanel
                kind="user-video"
                loading={userResult.loading || videoResult.loading}
                error={(userResult.error || videoResult.error) ? String(userResult.error || videoResult.error) : ""}
                onRetry={() => void reload()}
                items={videos()}
                sourceName={user()?.nickname || "用户作品"}
                fallbackAuthor={user()?.nickname || "未知作者"}
                showToast={showToast}
                refreshing={userResult.loading || videoResult.loading}
                onRefresh={() => void reload()}
                hasMore={hasMore()}
                loadingMore={loadingMore()}
                onLoadMore={() => void loadMore()}
              />
            </div>

            <CollectionVideoPanel
              active={activeTab() === "series"}
              kind="user-mix"
              sourceKey={`user:${secUserId()}`}
              refreshKey={seriesRefreshKey()}
              showToast={showToast}
              loadList={(cursor) => loadUserMixes(secUserId(), cursor)}
              loadVideos={(item, cursor) => loadUserMixVideos(secUserId(), item, cursor)}
            />
          </div>
        </div>
      </main>
      <Toast message={message()} type={type()}/>
    </section>
  );
}
