import {createFileRoute} from '@tanstack/solid-router'
import {createResource, For, type JSXElement, Match, Show, Switch} from "solid-js";
import {ClearDownloadHistory, DeleteDownloadHistory, DownloadHistory} from "../../../wailsjs/go/api/Douyin";
import {OpenDownloadLocation, OpenLocalFile} from "../../../wailsjs/go/utils/Settings";
import DetailError from "../../components/DetailError.tsx";
import IconChat from "../../components/icons/IconChat";
import IconEye from "../../components/icons/IconEye";
import IconFolderOpen from "../../components/icons/IconFolderOpen";
import IconPlayCircle from "../../components/icons/IconPlayCircle";
import IconRefresh from "../../components/icons/IconRefresh";
import NoCover from "../../components/NoCover.tsx";
import Toast from "../../components/Toast.tsx";
import {useToast} from "../../hooks/useToast.ts";
import {formatCount, formatDate, formatDuration} from "../../lib/format";
import {formatDownloadedAt} from "../../utils/format.ts";

export const Route = createFileRoute('/douyin/history')({
  component: History,
})


function History(): JSXElement {
  const {message, type, showToast} = useToast();
  const [items, {refetch, mutate}] = createResource(async () => {
    const data = await DownloadHistory();
    return data ?? [];
  });

  const historyItems = () => items() ?? [];

  const removeHistory = async (awemeId: string) => {
    if (!awemeId) return;
    try {
      await DeleteDownloadHistory(awemeId);
      mutate((current) => (current ?? []).filter((item) => item.awemeId !== awemeId));
      showToast("历史记录已删除", "success");
    } catch (error) {
      showToast(error instanceof Error ? error.message : String(error), "error");
    }
  };

  async function clearHistory(): Promise<void> {
    if (historyItems().length === 0) return;
    if (!window.confirm("确定要删除所有下载历史吗？本地文件不会被删除。")) return;
    try {
      await ClearDownloadHistory();
      mutate([]);
      showToast("下载历史已清空", "success");
    } catch (error) {
      showToast(error instanceof Error ? error.message : String(error), "error");
    }
  }

  async function openLocalFile(path: string): Promise<void> {
    try {
      await OpenLocalFile(path);
    } catch (error) {
      showToast(error instanceof Error ? error.message : String(error), "error");
    }
  }

  async function openLocation(path: string): Promise<void> {
    try {
      await OpenDownloadLocation(path);
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
            已记录 {historyItems().length} 个抖音作品
          </p>
        </div>
        <div class="flex items-center gap-2">
          <button class="btn btn-outline btn-sm gap-1.5" type="button" onClick={() => void refetch()}
                  disabled={items.loading}>
            <IconRefresh class={items.loading ? "h-4 w-4 animate-spin" : "h-4 w-4"}/>
            {items.loading ? "刷新中..." : "刷新"}
          </button>
          <button
            class="btn btn-outline btn-error btn-sm"
            type="button"
            onClick={() => void clearHistory()}
            disabled={items.loading || historyItems().length === 0}
          >
            删除全部历史记录
          </button>
        </div>
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
                  <article class="flex gap-3 p-3 flex-row">
                    {/*水平布局，左侧封面*/}
                    <div class="relative aspect-3/4 w-28 shrink-0 overflow-hidden rounded bg-base-200">
                      <Show when={item.cover} fallback={<NoCover/>}>
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
                        {item.isImageAlbum ? `图文 ${item.imageCount || 0}` : formatDuration(item.duration)}
                      </span>
                    </div>
                    <div class="min-w-0 flex-1 flex-col flex">
                      {/*右侧部分第一行：title+3个button*/}
                      <div class="flex gap-2">
                        <h3 class="line-clamp-2 flex-1 text-sm font-semibold leading-5">
                          {item.title}
                        </h3>
                        <button
                          class="btn btn-ghost btn-square btn-xs shrink-0"
                          title="本地打开"
                          onClick={() => void openLocalFile(item.path)}
                        >
                          <IconPlayCircle class="h-4 w-4"/>
                        </button>
                        <button
                          class="btn btn-ghost btn-square btn-xs shrink-0"
                          title="打开所在目录"
                          onClick={() => void openLocation(item.path)}
                        >
                          <IconFolderOpen class="h-4 w-4"/>
                        </button>
                        <button
                          class="btn btn-ghost btn-xs shrink-0 text-error"
                          onClick={() => void removeHistory(item.awemeId)}
                        >
                          删除记录
                        </button>
                      </div>
                      <div class="flex flex-col flex-1">
                        <div class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-base-content/60">
                          <span>{item.authorName || "未知作者"}</span>
                          <span>{item.awemeId}</span>
                          <Show when={item.publishTime}>
                            <span>发布 {formatDate(item.publishTime)}</span>
                          </Show>
                          <span>下载 {formatDownloadedAt(item.downloaded)}</span>
                        </div>
                        {/*点赞数+收藏数*/}
                        <div class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-base-content/60">
                        <span class="inline-flex items-center gap-1">
                          <IconEye class="h-3 w-3"/>{formatCount(item.diggCount)}
                        </span>
                          <span class="inline-flex items-center gap-1">
                          <IconChat class="h-3 w-3"/>{formatCount(item.collectCount)}
                        </span>
                        </div>
                        <p class="truncate mt-auto text-xs text-base-content/45 mb-2" title={item.path}>
                          {item.path}
                        </p>
                      </div>
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
