import {createFileRoute, Link} from '@tanstack/solid-router'
import {createResource, createSignal, type JSXElement, Match, Show, Switch} from "solid-js";
import {User, UserVideoList} from "../../../../wailsjs/go/api/Douyin";
import {model} from "../../../../wailsjs/go/models";
import MixPanel from "../../../components/douyin/MixPanel.tsx";
import VideoContentPanel from "../../../components/douyin/VideoContentPanel.tsx";
import IconRefresh from "../../../components/icons/IconRefresh.tsx";
import Toast from "../../../components/Toast.tsx";
import {useToast} from "../../../hooks/useToast.ts";
import {formatCount} from "../../../lib/format.ts";

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

function DouyinUserPage(): JSXElement {
  const params = Route.useParams();
  const secUserId = () => params().secUserId;
  const [activeTab, setActiveTab] = createSignal<UserTab>("video");
  const [loadingMore, setLoadingMore] = createSignal(false);
  // 用户合集列表由 MixPanel 拉取；刷新按钮通过 refreshKey 通知它重置并重新请求。
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
      // 用户作品接口使用 max_cursor 分页，追加时保留前面已经加载的作品。
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
      // 合集 tab 的数据不在当前路由组件里，靠 refreshKey 让子组件自行刷新。
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
                    class="badge badge-outline badge-sm">作品 {formatCount(user()?.aweme_count ?? videos().length)}</span>
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
                {/* 用户作品和收藏视频共用 VideoContentPanel，区别只在 kind/sourceName/fallbackAuthor。 */}
                <VideoContentPanel
                  kind="user-video"
                  loading={userResult.loading || videoResult.loading}
                  error={(userResult.error || videoResult.error) ? String(userResult.error || videoResult.error) : ""}
                  onRetry={() => void reload()}
                  items={videos()}
                  sourceName={user()?.nickname || "用户作品"}
                  fallbackAuthor={user()?.nickname || "未知作者"}
                  showToast={showToast}
                  hasMore={hasMore()}
                  loadingMore={loadingMore()}
                  onLoadMore={() => void loadMore()}
                />
              </Match>
              {/*右侧合集网格*/}
              <Match when={activeTab() === "series"}>
                {/* 传入 secUserId 后，MixPanel 会切到“用户合集”接口和文案。 */}
                <MixPanel
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
