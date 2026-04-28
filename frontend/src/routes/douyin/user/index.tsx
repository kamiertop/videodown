import {createFileRoute, useNavigate} from '@tanstack/solid-router'
import {createResource, createSignal, For, type JSXElement, Match, Show, Switch} from "solid-js";
import {FollowList, ParseSecUserID, User} from "../../../../wailsjs/go/api/Douyin";
import {model} from "../../../../wailsjs/go/models";
import DetailError from "../../../components/DetailError.tsx";
import DetailLoading from "../../../components/DetailLoading.tsx";
import IconRefresh from "../../../components/icons/IconRefresh.tsx";
import IconUsers from "../../../components/icons/IconUsers.tsx";
import Toast from "../../../components/Toast.tsx";
import {useToast} from "../../../hooks/useToast.ts";

export const Route = createFileRoute('/douyin/user/')({
  component: DouyinUserIndexPage,
})

function DouyinUserIndexPage(): JSXElement {
  const navigate = useNavigate();
  const [parsing, setParsing] = createSignal(false);
  const [searchInput, setSearchInput] = createSignal("");
  const [parsedSecUserID, setParsedSecUserID] = createSignal("");
  const [parsedUser, setParsedUser] = createSignal<model.User | null>(null);
  const [loadingMore, setLoadingMore] = createSignal(false);
  const {message, type, showToast} = useToast();
  const [result, {mutate, refetch}] = createResource(async () => FollowList(0));

  const users = () => result()?.followings ?? [];

  function coverUrl(item: model.FollowItem): string {
    return item.avatar_larger.url_list[0] || item.avatar_medium.url_list[0] || item.avatar_thumb.url_list[0] || '';
  }

  function userAvatarUrl(user: model.User): string {
    return user.avatar_larger?.url_list?.[0]
      || user.avatar_medium?.url_list?.[0]
      || user.avatar_thumb?.url_list?.[0]
      || "";
  }

  function userTotal(): number {
    return result()?.total ?? 0;
  }

  // 是否还有更多用户数据可以加载
  function hasMore(): boolean {
    const data = result();
    if (!data) {
      return false;
    }

    return data.has_more;
  }

  async function reload(): Promise<void> {
    await refetch();
  }

  async function handleSearch(): Promise<void> {
    const raw = searchInput().trim();
    if (!raw || parsing()) return;

    setParsing(true);
    setParsedSecUserID("");
    setParsedUser(null);
    try {
      const secUserID = await ParseSecUserID(raw);
      const resp = await User(secUserID);
      setParsedSecUserID(secUserID);
      setParsedUser(resp.user);
    } catch (error) {
      showToast(error instanceof Error ? error.message : String(error), "error");
    } finally {
      setParsing(false);
    }
  }

  function handleKeyDown(e: KeyboardEvent): void {
    if (e.key === "Enter") void handleSearch();
  }

  function goToUser(secUserID: string): void {
    if (!secUserID) return;
    void navigate({to: '/douyin/user/$secUserId', params: {secUserId: secUserID}});
  }

  async function loadMore(): Promise<void> {
    const data = result();
    if (!data || !data.has_more || loadingMore()) {
      return;
    }

    setLoadingMore(true);
    try {
      const nextOffset = users().length;
      const next = await FollowList(nextOffset);
      mutate(model.FollowResponse.createFrom({
        ...next,
        followings: [...users(), ...(next.followings ?? [])],
        total: data.total ?? next.total,
      }));
    } finally {
      setLoadingMore(false);
    }
  }

  return (
    <section class="flex h-full flex-col p-4">
      {/*垂直布局*/}
      <header class="mb-3 rounded-lg border border-base-300 bg-base-100 px-4 py-3">
        <div class="flex items-center justify-between gap-3">
          <div class="flex items-center gap-2">
            <Show when={userTotal() > 0}>
              <span class="rounded-full bg-primary/10 px-2 py-0.5 text-xs font-bold tabular-nums text-primary">
                {userTotal()} 关注
              </span>
            </Show>
            <button
              class="flex h-8 w-8 items-center justify-center rounded transition-colors hover:bg-base-200"
              onClick={() => void reload()}
              title="刷新"
            >
              <IconRefresh class={`h-4 w-4 text-base-content/50 ${result.loading ? 'animate-spin' : ''}`}/>
            </button>
          </div>

          <div class="flex items-center gap-2">
            <input
              type="text"
              class="input input-sm input-bordered w-80"
              placeholder="输入抖音用户链接 / 分享链接 / sec_user_id"
              value={searchInput()}
              onInput={(e) => setSearchInput(e.currentTarget.value)}
              onKeyDown={handleKeyDown}
            />
            <button class="btn btn-primary btn-sm" onClick={() => void handleSearch()} disabled={parsing()}>
              <Show when={!parsing()} fallback={<span class="loading loading-spinner loading-xs"></span>}>
                解析
              </Show>
            </button>
          </div>
        </div>

        <Show when={parsedUser()}>
          <div class="mt-3 rounded-xl border border-base-300 bg-base-200/30 p-3">
            <div class="flex flex-wrap items-center gap-3">
              <div class="h-10 w-10 overflow-hidden rounded-full bg-base-200 ring-2 ring-base-200">
                <Show when={userAvatarUrl(parsedUser()!)} fallback={<div class="h-full w-full bg-base-200"/>}>
                  <img
                    src={userAvatarUrl(parsedUser()!)}
                    alt={parsedUser()!.nickname}
                    referrerPolicy="no-referrer"
                    class="h-full w-full object-cover"
                  />
                </Show>
              </div>
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-2">
                  <span
                    class="truncate text-sm font-bold text-base-content">{parsedUser()!.nickname || "抖音用户"}</span>
                  <span class="badge badge-outline badge-sm">作品 {parsedUser()!.aweme_count ?? 0}</span>
                  <span class="badge badge-outline badge-sm">粉丝 {parsedUser()!.follower_count ?? 0}</span>
                  <Show when={parsedUser()!.ip_location}>
                    <span class="badge badge-ghost badge-sm">{parsedUser()!.ip_location}</span>
                  </Show>
                </div>
                <div class="mt-0.5 line-clamp-1 text-xs text-base-content/60">
                  {parsedUser()!.signature?.trim() ? parsedUser()!.signature : "这个账号还没有留下简介。"}
                </div>
              </div>
              <div class="flex items-center gap-2">
                <button class="btn btn-outline btn-sm" onClick={() => {
                  setParsedSecUserID("");
                  setParsedUser(null);
                }}>
                  清除
                </button>
                <button class="btn btn-primary btn-sm" onClick={() => goToUser(parsedSecUserID())}>
                  进入详情
                </button>
              </div>
            </div>
          </div>
        </Show>
      </header>

      <div class="min-h-0 flex-1 overflow-auto rounded-lg border border-base-300 bg-base-100">
        <Switch>
          <Match when={result.loading}>
            <DetailLoading/>
          </Match>
          <Match when={result.error}>
            <DetailError message={String(result.error)} onRetry={() => void reload()}/>
          </Match>
          <Match when={users().length === 0}>
            <div class="flex h-full items-center justify-center text-base-content/40">
              <div class="text-center">
                <IconUsers class="mx-auto h-14 w-14"/>
                <p class="mt-3 text-sm font-semibold text-base-content/60">暂无用户数据</p>
                <p class="mt-1 text-xs text-base-content/50">可以通过上方搜索框直接解析用户链接</p>
              </div>
            </div>
          </Match>
          {/*渲染用户网格*/}
          <Match when={users().length > 0}>
            <div class="p-4">
              <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
                <For each={users()}>
                  {(item: model.FollowItem): JSXElement => (
                    <button
                      class="group flex flex-col items-center gap-2 rounded-xl border
                      border-base-300 bg-base-100 p-4 text-left transition-all duration-150 hover:-translate-y-px
                      hover:border-primary/40 hover:shadow-md active:scale-[0.98] justify-end"
                      type="button"
                      onClick={() => {
                        if (!item.sec_uid) return;
                        void navigate({to: '/douyin/user/$secUserId', params: {secUserId: item.sec_uid}});
                      }}
                      disabled={!item.sec_uid}
                    >
                      <div class="relative">
                        <div
                          class="h-16 w-16 overflow-hidden rounded-full bg-base-200 ring-2 ring-base-200 transition group-hover:ring-primary/30">
                          <img
                            src={coverUrl(item)}
                            alt={item.nickname}
                            class="h-full w-full object-cover"
                            loading="lazy"
                            referrerPolicy="no-referrer"
                          />
                        </div>
                      </div>
                      <span
                        class="max-w-full truncate text-sm font-semibold text-base-content group-hover:text-primary">
                        {item.nickname}
                      </span>
                      <span class="line-clamp-2 max-w-full text-center text-xs leading-relaxed text-base-content/50">
                          {item.signature || "这个账号还没有留下简介。"}
                      </span>
                      <div class="flex flex-wrap items-center justify-center gap-2">
                        <span class="badge badge-outline badge-sm">作品 {item.aweme_count}</span>
                        <span class="badge badge-outline badge-sm">粉丝 {item.follower_count}</span>
                      </div>
                    </button>
                  )}
                </For>
              </div>
              {/*当有更多数据可以加载时，显示加载更多按钮*/}
              <Show when={hasMore()}>
                <div class="mt-4 flex items-center justify-center">
                  <button
                    class="btn btn-outline btn-sm"
                    onClick={() => void loadMore()}
                    disabled={loadingMore()}
                  >
                    <Show when={!loadingMore()}>
                      加载更多用户
                    </Show>
                  </button>
                </div>
              </Show>
            </div>
          </Match>
        </Switch>
      </div>
      <Toast message={message()} type={type()}/>
    </section>
  )
}
