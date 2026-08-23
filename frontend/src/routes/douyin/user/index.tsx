import {createFileRoute, useNavigate} from '@tanstack/solid-router'
import {createEffect, createSignal, For, onMount, type JSXElement, Match, onCleanup, Show, Switch} from "solid-js";
import {ParseSecUserID, SearchFollow, User, FollowList} from "../../../../wailsjs/go/api/Douyin";
import {model} from "../../../../wailsjs/go/models";
import DetailError from "../../../components/DetailError.tsx";
import DetailLoading from "../../../components/DetailLoading.tsx";
import IconRefresh from "../../../components/icons/IconRefresh.tsx";
import IconUsers from "../../../components/icons/IconUsers.tsx";
import Toast from "../../../components/Toast.tsx";
import {useToast} from "../../../hooks/useToast.ts";
import {
  appendFollowData,
  followData,
  followLoadingMore,
  followSearchInput,
  followSearchResults,
  followSearching,
  resetFollowSearch,
  setFollowLoadingMore,
  setFollowSearchInput,
  setFollowSearchResults,
  setFollowSearching,
  updateFollowData,
} from "../../../lib/douyin/followStore.ts";

export const Route = createFileRoute('/douyin/user/')({
  component: DouyinUserIndexPage,
})

type FollowCardItem = {
  secUid: string;
  avatar: string;
  nickname: string;
  signature: string;
  awemeCount?: number;
  followerCount?: number;
  isSearchResult: boolean;
};

function DouyinUserIndexPage(): JSXElement {
  const navigate = useNavigate();
  const {message, type, showToast} = useToast();
  const [initialLoading, setInitialLoading] = createSignal(!followData());
  const [loadError, setLoadError] = createSignal("");
  const [parsing, setParsing] = createSignal(false);
  const [parseInput, setParseInput] = createSignal("");
  const [parsedSecUserID, setParsedSecUserID] = createSignal("");
  const [parsedUser, setParsedUser] = createSignal<model.User | null>(null);
  let followSearchSeq = 0;

  const users = () => followData()?.followings ?? [];

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
    return followData()?.total ?? 0;
  }

  const followCards = () => {
    if (followSearchResults() !== null) {
      return followSearchResults()!.map((item): FollowCardItem => ({
        secUid: item.rich_sug_sec_uid,
        avatar: item.rich_sug_avatar_uri,
        nickname: item.rich_sug_nickname || item.rich_sug_user_id || "未命名用户",
        signature: item.rich_sug_follow_status === "follow"
            ? "已关注"
            : item.rich_sug_relation_type || item.rich_sug_follow_status || "搜索结果",
        isSearchResult: true,
      }));
    }

    return users().map((item): FollowCardItem => ({
      secUid: item.sec_uid,
      avatar: coverUrl(item),
      nickname: item.nickname || item.uid || "未命名用户",
      signature: item.signature || "这个账号还没有留下简介。",
      awemeCount: item.aweme_count,
      followerCount: item.follower_count,
      isSearchResult: false,
    }));
  };

  function hasMore(): boolean {
    const data = followData();
    if (!data) return false;
    return data.has_more;
  }

  async function loadInitial(): Promise<void> {
    if (followData()) {
      setInitialLoading(false);
      return;
    }

    setInitialLoading(true);
    setLoadError("");
    try {
      updateFollowData(await FollowList(0));
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : String(error));
    } finally {
      setInitialLoading(false);
    }
  }

  async function reload(): Promise<void> {
    setInitialLoading(true);
    setLoadError("");
    try {
      updateFollowData(await FollowList(0));
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : String(error));
    } finally {
      setInitialLoading(false);
    }
  }

  createEffect(() => {
    const raw = followSearchInput().trim();
    const seq = ++followSearchSeq;

    if (!raw) {
      resetFollowSearch();
      return;
    }

    setFollowSearching(true);
    const timer = window.setTimeout(() => {
      void (async () => {
        try {
          const resp = await SearchFollow(raw);
          if (seq === followSearchSeq) {
            setFollowSearchResults(resp);
          }
        } catch (error) {
          if (seq === followSearchSeq) {
            showToast(error instanceof Error ? error.message : String(error), "error");
            setFollowSearchResults([]);
          }
        } finally {
          if (seq === followSearchSeq) {
            setFollowSearching(false);
          }
        }
      })();
    }, 450);

    onCleanup(() => window.clearTimeout(timer));
  });

  async function loadMore(): Promise<void> {
    const data = followData();
    if (!data || !data.has_more || followLoadingMore()) return;

    setFollowLoadingMore(true);
    try {
      const nextOffset = users().length;
      const next = await FollowList(nextOffset);
      appendFollowData(next, data.total ?? next.total, users());
    } catch (error) {
      showToast(error instanceof Error ? error.message : String(error), "error");
    } finally {
      setFollowLoadingMore(false);
    }
  }

  async function handleParseUser(): Promise<void> {
    const raw = parseInput().trim();
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

  function handleParseKeyDown(e: KeyboardEvent): void {
    if (e.key === "Enter") void handleParseUser();
  }

  function goToUser(secUserID: string): void {
    if (!secUserID) return;
    void navigate({to: '/douyin/user/$secUserId', params: {secUserId: secUserID}});
  }

  onMount(() => {
    void loadInitial();
  });

  return (
      <section class="flex h-full flex-col p-4">
        <header class="mb-3 rounded-lg border border-base-300 bg-base-100 px-4 py-3">
          <div class="flex items-center gap-3">
            <div class="flex shrink-0 items-center gap-2">
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
                <IconRefresh class={`h-4 w-4 text-base-content/50 ${initialLoading() ? 'animate-spin' : ''}`}/>
              </button>
            </div>

            <div class="flex min-w-0 flex-1 justify-center">
              <div class="flex min-w-0 max-w-3xl flex-wrap items-center justify-center gap-2">
                <input
                    type="text"
                    class="input input-sm input-bordered w-56 sm:w-64"
                    placeholder="输入昵称/抖音号从关注列表中搜索博主"
                    value={followSearchInput()}
                    onInput={(e) => setFollowSearchInput(e.currentTarget.value)}
                />
                <Show when={followSearching()}>
                  <span class="loading loading-spinner loading-xs text-primary"></span>
                </Show>
              </div>
            </div>

            <div class="ml-auto flex shrink-0 items-center gap-2">
              <input
                  type="text"
                  class="input input-sm input-bordered w-72 sm:w-80"
                  placeholder="输入抖音用户链接 / 分享链接 / sec_user_id"
                  value={parseInput()}
                  onInput={(e) => setParseInput(e.currentTarget.value)}
                  onKeyDown={handleParseKeyDown}
              />
              <button class="btn btn-primary btn-sm" onClick={() => void handleParseUser()} disabled={parsing()}>
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
                    <span class="truncate text-sm font-bold text-base-content">{parsedUser()!.nickname || "抖音用户"}</span>
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
            <Match when={initialLoading()}>
              <DetailLoading/>
            </Match>
            <Match when={loadError()}>
              <DetailError message={loadError()} onRetry={() => void reload()}/>
            </Match>
            <Match when={followSearching()}>
              <div class="flex h-full items-center justify-center text-base-content/40">
                <div class="text-center">
                  <span class="loading loading-spinner loading-md text-primary"></span>
                  <p class="mt-3 text-sm font-semibold text-base-content/60">正在搜索关注</p>
                  <p class="mt-1 text-xs text-base-content/50">输入停止后会自动刷新结果</p>
                </div>
              </div>
            </Match>
            <Match when={followCards().length === 0}>
              <div class="flex h-full items-center justify-center text-base-content/40">
                <div class="text-center">
                  <IconUsers class="mx-auto h-14 w-14"/>
                  <p class="mt-3 text-sm font-semibold text-base-content/60">
                    {followSearchResults() !== null ? "未找到匹配关注" : "暂无用户数据"}
                  </p>
                  <p class="mt-1 text-xs text-base-content/50">
                    {followSearchResults() !== null ? "可以换个关键词继续搜索" : "可以通过上方搜索框直接解析用户链接"}
                  </p>
                </div>
              </div>
            </Match>
            <Match when={followCards().length > 0}>
              <div class="p-4">
                <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
                  <For each={followCards()}>
                    {(item): JSXElement => (
                        <button
                            class="group flex flex-col items-center gap-2 rounded-xl border border-base-300 bg-base-100 p-4 text-left transition-all duration-150 hover:-translate-y-px hover:border-primary/40 hover:shadow-md active:scale-[0.98] justify-end"
                            type="button"
                            onClick={() => goToUser(item.secUid)}
                            disabled={!item.secUid}
                        >
                          <div class="relative">
                            <div class="h-16 w-16 overflow-hidden rounded-full bg-base-200 ring-2 ring-base-200 transition group-hover:ring-primary/30">
                              <img
                                  src={item.avatar}
                                  alt={item.nickname}
                                  class="h-full w-full object-cover"
                                  loading="lazy"
                                  referrerPolicy="no-referrer"
                              />
                            </div>
                          </div>
                          <span class="max-w-full truncate text-sm font-semibold text-base-content group-hover:text-primary">
                            {item.nickname}
                          </span>
                          <Show when={!item.isSearchResult}>
                            <span class="line-clamp-2 max-w-full text-center text-xs leading-relaxed text-base-content/50">
                              {item.signature}
                            </span>
                            <div class="flex flex-wrap items-center justify-center gap-2">
                              <span class="badge badge-outline badge-sm">作品 {item.awemeCount}</span>
                              <span class="badge badge-outline badge-sm">粉丝 {item.followerCount}</span>
                            </div>
                          </Show>
                        </button>
                    )}
                  </For>
                </div>
                <Show when={followSearchResults() === null && hasMore()}>
                  <div class="mt-4 flex items-center justify-center">
                    <button
                        class="btn btn-outline btn-sm"
                        onClick={() => void loadMore()}
                        disabled={followLoadingMore()}
                    >
                      <Show when={!followLoadingMore()}>
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
  );
}
