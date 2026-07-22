import {createFileRoute, useNavigate} from '@tanstack/solid-router'
import {createSignal, For, type JSXElement, Match, onMount, Show, Switch} from "solid-js";
import {DynamicVideosPage} from "../../../wailsjs/go/api/BiliBili";
import {model} from "../../../wailsjs/go/models";
import DetailError from "../../components/DetailError";
import NoCover from "../../components/NoCover";
import Toast from "../../components/Toast";
import IconRefresh from "../../components/icons/IconRefresh";
import {useToast} from "../../hooks/useToast";
import {parseBilibiliLengthToSeconds} from "../../lib/format";
import {addVideos} from "../../lib/bilibili/store";
import type {MediaCardItem} from "../../lib/model";

type DynamicArchiveItem = model.DynamicArchiveItem;

export const Route = createFileRoute('/bilibili/dynamic')({
  component: Dynamic,
})

function formatDynamicTime(item: DynamicArchiveItem): string {
  const relative = item.pub_time?.trim();
  if (relative) return relative;
  if (!item.pub_ts) return "";

  const date = new Date(item.pub_ts * 1000);
  const now = new Date();
  if (date.getFullYear() === now.getFullYear()) {
    return `${date.getMonth() + 1}-${String(date.getDate()).padStart(2, "0")} ${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
  }
  return `${date.getFullYear()}-${date.getMonth() + 1}-${String(date.getDate()).padStart(2, "0")}`;
}

function mediaID(bvid: string): number {
  let hash = 0;
  for (let i = 0; i < bvid.length; i += 1) {
    hash = ((hash << 5) - hash + bvid.charCodeAt(i)) | 0;
  }
  return Math.abs(hash) || Date.now();
}

function toMediaCardItem(item: DynamicArchiveItem): MediaCardItem {
  return {
    id: mediaID(item.bvid),
    title: item.title,
    cover: item.cover,
    duration: parseBilibiliLengthToSeconds(item.duration_text),
    bvid: item.bvid,
    link: `https://www.bilibili.com/video/${item.bvid}`,
    upperName: item.author_name,
    pubtime: item.pub_ts,
    sourceListName: "关注动态",
    sourceListKind: "关注动态",
  };
}

function Dynamic(): JSXElement {
  const navigate = useNavigate();
  const {message, type, showToast} = useToast();
  const [items, setItems] = createSignal<DynamicArchiveItem[]>([]);
  const [offset, setOffset] = createSignal("");
  const [hasMore, setHasMore] = createSignal(false);
  const [loading, setLoading] = createSignal(false);
  const [loadingMore, setLoadingMore] = createSignal(false);
  const [error, setError] = createSignal("");

  async function loadPage(nextOffset: string, append: boolean): Promise<void> {
    if (append ? loadingMore() : loading()) return;
    append ? setLoadingMore(true) : setLoading(true);
    setError("");

    try {
      const page = await DynamicVideosPage(nextOffset);
      const nextItems = page.items ?? [];
      setItems((current) => append ? [...current, ...nextItems] : nextItems);
      setOffset(page.offset ?? "");
      setHasMore(!!page.has_more);
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      setError(msg);
      showToast(msg, "error");
    } finally {
      append ? setLoadingMore(false) : setLoading(false);
    }
  }

  async function refresh(): Promise<void> {
    setItems([]);
    setOffset("");
    setHasMore(false);
    await loadPage("", false);
  }

  async function enqueueAndGoDownload(targets: DynamicArchiveItem[]): Promise<void> {
    const downloadable = targets.filter((item) => item.bvid?.trim());
    if (downloadable.length === 0) {
      showToast("没有可下载的视频", "warning");
      return;
    }

    addVideos(downloadable.map(toMediaCardItem));
    await navigate({to: "/bilibili/download"});
  }

  onMount(() => {
    void loadPage("", false);
  });

  return (
      <section class="flex h-full min-h-0 flex-col bg-base-200/40 p-4">
        <header class="mb-3 flex shrink-0 items-center justify-between rounded-lg border border-base-300 bg-base-100 px-4 py-3">
          <div class="min-w-0">
            <h2 class="text-base font-bold text-base-content">关注动态</h2>
            <p class="text-sm text-base-content/60">
              已加载 {items().length} 个视频
            </p>
          </div>
          <div class="flex items-center gap-2">
            <button
                class="btn btn-outline btn-sm"
                type="button"
                onClick={() => void refresh()}
                disabled={loading()}
            >
              <IconRefresh class={`h-3.5 w-3.5 ${loading() ? "animate-spin" : ""}`}/>
              {loading() ? "刷新中" : "刷新"}
            </button>
            <button
                class="btn btn-primary btn-sm"
                type="button"
                onClick={() => void enqueueAndGoDownload(items())}
                disabled={items().length === 0 || loading()}
            >
              下载全部
            </button>
          </div>
        </header>

        <section class="min-h-0 flex-1 overflow-hidden rounded-lg border border-base-300 bg-base-100">
          <Switch>
            <Match when={loading()}>
              <div class="flex h-full items-center justify-center">
                <span class="loading loading-spinner loading-md text-primary"/>
              </div>
            </Match>
            <Match when={error()}>
              <DetailError message={error()} onRetry={() => void refresh()}/>
            </Match>
            <Match when={items().length === 0}>
              <div class="flex h-full items-center justify-center text-sm text-base-content/50">暂无关注视频动态</div>
            </Match>
            <Match when={items().length > 0}>
              <div class="flex h-full min-h-0 flex-col">
                <div class="grid h-10 shrink-0 grid-cols-[minmax(0,1fr)_9rem_8rem_7rem] items-center border-b border-base-300 bg-base-200/45 px-4 text-xs font-bold text-base-content/60">
                  <span>视频</span>
                  <span>作者</span>
                  <span>时间</span>
                  <span class="text-right">操作</span>
                </div>
                <div class="min-h-0 flex-1 overflow-y-auto divide-y divide-base-200">
                  <For each={items()}>
                    {(item): JSXElement => (
                        <article class="grid grid-cols-[minmax(0,1fr)_9rem_8rem_7rem] items-center gap-3 px-4 py-3 transition-colors hover:bg-base-200/35">
                          <div class="flex min-w-0 items-center gap-3">
                            <div class="relative aspect-video w-32 shrink-0 overflow-hidden rounded bg-base-200">
                              <Show when={item.cover} fallback={<NoCover/>}>
                                <img
                                    class="h-full w-full object-cover"
                                    src={item.cover}
                                    alt={item.title}
                                    referrerPolicy="no-referrer"
                                    loading="lazy"
                                />
                              </Show>
                              <span class="absolute bottom-1 right-1 rounded bg-black/65 px-1 py-0.5 text-[11px] tabular-nums text-white">
                                {item.duration_text || "--:--"}
                              </span>
                            </div>
                            <div class="min-w-0 flex-1">
                              <h3 class="line-clamp-2 text-sm font-semibold leading-5 text-base-content" title={item.title}>
                                {item.title || "未命名视频"}
                              </h3>
                              <div class="mt-1 truncate text-xs text-base-content/40">{item.bvid}</div>
                            </div>
                          </div>
                          <span class="truncate text-sm text-base-content/70" title={item.author_name}>
                            {item.author_name || "未知 UP"}
                          </span>
                          <span class="text-sm tabular-nums text-base-content/60">{formatDynamicTime(item)}</span>
                          <div class="flex justify-end">
                            <button
                                class="btn btn-primary btn-sm"
                                type="button"
                                onClick={() => void enqueueAndGoDownload([item])}
                                disabled={!item.bvid}
                            >
                              下载
                            </button>
                          </div>
                        </article>
                    )}
                  </For>
                </div>
                <Show when={hasMore()}>
                  <div class="shrink-0 border-t border-base-300 bg-base-100 px-4 py-3 text-center">
                    <button
                        class="btn btn-outline btn-sm"
                        type="button"
                        disabled={loadingMore()}
                        onClick={() => void loadPage(offset(), true)}
                    >
                      {loadingMore() ? "加载中..." : "加载更多"}
                    </button>
                  </div>
                </Show>
              </div>
            </Match>
          </Switch>
        </section>

        <Toast message={message()} type={type()}/>
      </section>
  )
}
