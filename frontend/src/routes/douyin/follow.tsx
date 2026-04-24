import {createFileRoute} from '@tanstack/solid-router'
import {createResource, createSignal, For, type JSXElement, Match, Show, Switch} from "solid-js";
import {FollowList} from "../../../wailsjs/go/api/Douyin";
import {model} from "../../../wailsjs/go/models";
import DetailError from "../../components/DetailError.tsx";
import DetailLoading from "../../components/DetailLoading.tsx";
import IconRefresh from "../../components/icons/IconRefresh.tsx";
import IconUsers from "../../components/icons/IconUsers.tsx";

export const Route = createFileRoute('/douyin/follow')({
  component: DouyinFollowPage,
})

function DouyinFollowPage(): JSXElement {
  const [loadingMore, setLoadingMore] = createSignal(false);
  const [result, {mutate, refetch}] = createResource(async () => FollowList(0));

  const followings = () => result()?.followings ?? [];

  function coverUrl(item: model.FollowItem): string {
    return item.avatar_larger.url_list[0] || item.avatar_medium.url_list[0] || item.avatar_thumb.url_list[0] || '';
  }

  function followTotal(): number {
    return result()?.total ?? 0;
  }

  // 是否还有更多关注数据可以加载
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

  async function loadMore(): Promise<void> {
    const data = result();
    if (!data || !data.has_more || loadingMore()) {
      return;
    }

    setLoadingMore(true);
    try {
      const nextOffset = followings().length;
      const next = await FollowList(nextOffset);
      mutate(model.FollowResponse.createFrom({
        ...next,
        followings: [...followings(), ...(next.followings ?? [])],
        total: data.total ?? next.total,
      }));
    } finally {
      setLoadingMore(false);
    }
  }

  return (
    <section class="flex h-full flex-col p-4">
      {/*垂直布局*/}
      <header class="mb-3 flex items-center justify-between rounded-lg border border-base-300 bg-base-100 px-4 py-3">
        <div class="flex items-center gap-2">
          <h2 class="text-base font-bold">关注列表</h2>
          <Show when={followTotal() > 0}>
            <span class="rounded-full bg-primary/10 px-2 py-0.5 text-xs font-bold tabular-nums text-primary">
              {followTotal()}
            </span>
          </Show>
        </div>
        <button
          class="flex h-8 w-8 items-center justify-center rounded transition-colors hover:bg-base-200"
          onClick={() => void reload()}
          title="刷新"
        >
          <IconRefresh class={`h-4 w-4 text-base-content/50 ${result.loading ? 'animate-spin' : ''}`}/>
        </button>
      </header>

      <div class="min-h-0 flex-1 overflow-auto rounded-lg border border-base-300 bg-base-100">
        <Switch>
          <Match when={result.loading}>
            <DetailLoading/>
          </Match>
          <Match when={result.error}>
            <DetailError message={String(result.error)} onRetry={() => void reload()}/>
          </Match>
          <Match when={followings().length === 0}>
            <div class="flex h-full items-center justify-center text-base-content/40">
              <div class="text-center">
                <IconUsers class="mx-auto h-14 w-14"/>
                <p class="mt-3 text-sm font-semibold text-base-content/60">暂无关注数据</p>
                <p class="mt-1 text-xs text-base-content/50">请先确认 Cookie 可用，或稍后重试</p>
              </div>
            </div>
          </Match>
          {/*渲染关注用户网格*/}
          <Match when={followings().length > 0}>
            <div class="p-4">
              <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
                <For each={followings()}>
                  {(item: model.FollowItem): JSXElement => (
                    <button
                      class="group flex flex-col items-center gap-2 rounded-xl border
                      border-base-300 bg-base-100 p-4 text-left transition-all duration-150 hover:-translate-y-px
                      hover:border-primary/40 hover:shadow-md active:scale-[0.98] justify-end"
                      type="button"
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
                      加载剩余关注
                    </Show>
                  </button>
                </div>
              </Show>
            </div>
          </Match>
        </Switch>
      </div>
    </section>
  )
}
