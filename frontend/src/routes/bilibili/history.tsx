import {createFileRoute} from '@tanstack/solid-router'
import {createResource, For, type JSXElement, Match, Show, Switch} from "solid-js";
import {DeleteDownloadHistory, DownloadHistory} from "../../../wailsjs/go/api/BiliBili";
import DetailError from "../../components/DetailError.tsx";
import IconChat from "../../components/icons/IconChat";
import IconEye from "../../components/icons/IconEye";
import Toast from "../../components/Toast.tsx";
import {useToast} from "../../hooks/useToast.ts";
import {formatCount, formatDate, formatDuration} from "../../lib/format";

export const Route = createFileRoute('/bilibili/history')({
  component: History,
})

function formatDownloadedAt(value: any): string {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  const hh = String(date.getHours()).padStart(2, "0");
  const mm = String(date.getMinutes()).padStart(2, "0");
  const ss = String(date.getSeconds()).padStart(2, "0");
  return `${y}-${m}-${d} ${hh}:${mm}:${ss}`;
}

function History(): JSXElement {
  const {message, type, showToast} = useToast();
  const [items, {refetch, mutate}] = createResource(async () => {
    const data = await DownloadHistory();
    return data ?? [];
  });

  const historyItems = () => items() ?? [];

  // 移除下载历史记录
  async function removeHistory(cid: number): Promise<void> {
    try {
      await DeleteDownloadHistory(cid);
      mutate((current) => (current ?? []).filter((item) => item.cid !== cid));
      showToast("历史记录已删除", "success");
    } catch (error) {
      showToast(error instanceof Error ? error.message : String(error), "error");
    }
  }

  return (
    <div class="flex h-full flex-col p-4">
      <section class="mb-3 flex items-center justify-between rounded-lg border border-base-300 bg-base-100 px-4 py-3">
        <div>
          <h2 class="text-base font-bold">下载历史</h2>
          <p class="text-sm text-base-content/60">
            已记录 {historyItems().length} 个视频
          </p>
        </div>
        <button class="btn btn-outline btn-sm" type="button" onClick={() => void refetch()} disabled={items.loading}>
          {items.loading ? "刷新中..." : "刷新"}
        </button>
      </section>

      <section class="min-h-0 flex-1 overflow-y-auto rounded-lg border border-base-300 bg-base-100">
        <Switch>
          <Match when={items.loading}>
            <div class="flex h-full items-center justify-center">
              <span class="loading loading-spinner loading-md text-primary"/>
            </div>
          </Match>
          <Match when={items.error}>
            <DetailError message={String(items.error)} onRetry={() => void refetch()}/>
          </Match>
          <Match when={historyItems().length === 0}>
            <div class="flex h-full items-center justify-center text-sm text-base-content/50">
              暂无下载历史
            </div>
          </Match>
          <Match when={historyItems().length > 0}>
            <div class="divide-y divide-base-200">
              <For each={historyItems()}>
                {(item) => (
                  <article class="flex gap-3 p-3">
                    <div class="relative aspect-video w-44 shrink-0 overflow-hidden rounded bg-base-200">
                      <Show
                        when={item.cover}
                        fallback={
                          <div class="flex h-full items-center justify-center text-xs text-base-content/40">无封面</div>
                        }
                      >
                        <img
                          class="h-full w-full object-cover"
                          src={item.cover}
                          alt={item.title}
                          referrerPolicy="no-referrer"
                          loading="lazy"
                        />
                      </Show>
                      <span
                        class="absolute bottom-1 right-1 rounded bg-black/65 px-1 py-0.5 text-xs tabular-nums text-white">
                          {formatDuration(item.duration)}
                      </span>
                    </div>
                    <div class="min-w-0 flex-1">
                      <div class="flex items-start gap-2">
                        <h3 class="line-clamp-2 flex-1 text-sm font-semibold leading-5">
                          {item.title}
                        </h3>
                        <button
                          class="btn btn-ghost btn-xs shrink-0 text-error"
                          onClick={() => void removeHistory(item.cid)}
                        >
                          删除
                        </button>
                      </div>
                      <div class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-base-content/60">
                        <span>{item.upperName || "未知 UP"}</span>
                        <span>{item.bvid}</span>
                        <Show when={item.pubtime}>
                          <span>发布 {formatDate(item.pubtime)}</span>
                        </Show>
                        <span>下载 {formatDownloadedAt(item.downloaded)}</span>
                      </div>
                      <div class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-base-content/60">
                        <span class="inline-flex items-center gap-1">
                            <IconEye class="h-3 w-3"/>{formatCount(item.play)}
                        </span>
                        <span class="inline-flex items-center gap-1">
                            <IconChat class="h-3 w-3"/>{formatCount(item.danmaku)}
                        </span>
                      </div>
                      <p class="mt-2 truncate text-xs text-base-content/45" title={item.path}>
                        {item.path}
                      </p>
                    </div>
                  </article>
                )}
              </For>
            </div>
          </Match>
        </Switch>
      </section>

      <Toast message={message()} type={type()}/>
    </div>
  )
}
