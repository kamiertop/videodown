import {
  ClearDownloadHistory,
  DeleteDownloadHistory,
  DownloadHistoryPage
} from "@bindings/github.com/kamiertop/videodown/bilibili/api/bilibili";
import * as model from "@bindings/github.com/kamiertop/videodown/bilibili/model/models";
import {OpenDownloadLocation, OpenLocalFile} from "@bindings/github.com/kamiertop/videodown/utils/settings";
import {createFileRoute} from '@tanstack/solid-router'
import {createSignal, For, type JSXElement, Match, onMount, Show, Switch} from "solid-js";
import DetailError from "../../components/DetailError.tsx";
import IconChat from "../../components/icons/IconChat";
import IconEye from "../../components/icons/IconEye";
import IconFolderOpen from "../../components/icons/IconFolderOpen";
import IconPlayCircle from "../../components/icons/IconPlayCircle";
import NoCover from "../../components/NoCover.tsx";
import Toast from "../../components/Toast.tsx";
import {useToast} from "../../hooks/useToast.ts";
import {formatCount, formatDate, formatDuration} from "../../lib/format";
import {formatDownloadedAt} from "../../utils/format.ts";

type DownloadHistoryItem = model.DownloadHistoryItem;

export const Route = createFileRoute('/bilibili/history')({
  component: History,
})

// 历史按页加载：后端分页 + 关键字过滤，避免记录多时全量跨桥、全量渲染。
const PAGE_SIZE = 50;

function History(): JSXElement {
  const {message, type, showToast} = useToast();
  const [items, setItems] = createSignal<DownloadHistoryItem[]>([]);
  const [total, setTotal] = createSignal(0);
  const [hasMore, setHasMore] = createSignal(false);
  const [loading, setLoading] = createSignal(false);
  // 区分“还在加载”与“确实没有历史”，首次加载完成前不显示空态文案。
  const [loaded, setLoaded] = createSignal(false);
  const [errorMsg, setErrorMsg] = createSignal("");
  const [searchValue, setSearchValue] = createSignal<string>("");
  // 竞态防护：搜索词快速变化时只认最后一次请求的结果。
  let requestSeq = 0;
  let searchTimer: number | undefined;

  async function loadPage(reset: boolean): Promise<void> {
    const seq = ++requestSeq;
    setLoading(true);
    try {
      const offset = reset ? 0 : items().length;
      const page = await DownloadHistoryPage(offset, PAGE_SIZE, searchValue().trim());
      if (seq !== requestSeq) return;
      setErrorMsg("");
      setTotal(page.total ?? 0);
      setHasMore(page.hasMore);
      setItems(reset ? page.items ?? [] : [...items(), ...(page.items ?? [])]);
    } catch (error) {
      if (seq === requestSeq) {
        const msg = error instanceof Error ? error.message : String(error);
        setErrorMsg(msg);
        showToast(msg, "error");
      }
    } finally {
      if (seq === requestSeq) {
        setLoading(false);
        setLoaded(true);
      }
    }
  }

  function searchInput(value: string): void {
    setSearchValue(value);
    // 防抖：输入停顿后再发起后端搜索。
    window.clearTimeout(searchTimer);
    searchTimer = window.setTimeout(() => void loadPage(true), 300);
  }

  onMount(() => void loadPage(true));

  // 移除下载历史记录
  async function removeHistory(cid: number): Promise<void> {
    try {
      await DeleteDownloadHistory(cid);
      setItems((current) => current.filter((item) => item.cid !== cid));
      setTotal((current) => Math.max(0, current - 1));
      showToast("历史记录已删除", "success");
    } catch (error) {
      showToast(error instanceof Error ? error.message : String(error), "error");
    }
  }

  // 删除全部下载历史记录
  async function clearHistory(): Promise<void> {
    if (total() === 0 && items().length === 0) return;
    if (!window.confirm("确定要删除所有下载历史吗？本地文件不会被删除")) return;
    try {
      await ClearDownloadHistory();
      setItems([]);
      setTotal(0);
      setHasMore(false);
      showToast("下载历史已清空", "success");
    } catch (error) {
      showToast(error instanceof Error ? error.message : String(error), "error");
    }
  }

  // 打开已下载内容的本地文件
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

  function historyBadge(item: DownloadHistoryItem): string {
    if (item.downloadKind === "cover") return "封面";
    return formatDuration(item.duration);
  }

  function Card(item: DownloadHistoryItem): JSXElement {
    return (
        <article class="flex gap-3 p-3">
          <div class="relative aspect-video w-44 shrink-0 overflow-hidden rounded bg-base-200">
            <Show when={item.cover} fallback={<NoCover/>}>
              <img
                  class="h-full w-full object-cover"
                  src={item.cover}
                  alt={item.title}
                  referrerPolicy="no-referrer"
                  loading="lazy"
              />
            </Show>
            <span class="absolute bottom-1 right-1 rounded bg-black/65 px-1 py-0.5 text-xs tabular-nums text-white">
              {historyBadge(item)}
            </span>
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-start gap-2">
              <h3 class="line-clamp-2 flex-1 text-sm font-semibold leading-5">{item.title}</h3>
              <button
                  class="btn btn-ghost btn-square btn-xs shrink-0"
                  type="button"
                  title="本地打开"
                  onClick={() => void openLocalFile(item.path)}
              >
                <IconPlayCircle class="h-4 w-4"/>
              </button>
              <button
                  class="btn btn-ghost btn-square btn-xs shrink-0"
                  type="button"
                  title="打开所在目录"
                  onClick={() => void openLocation(item.path)}
              >
                <IconFolderOpen class="h-4 w-4"/>
              </button>
              <button
                  class="btn btn-ghost btn-xs shrink-0 text-error"
                  type="button"
                  onClick={() => void removeHistory(item.cid)}
              >
                删除
              </button>
            </div>
            <div class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-base-content/60">
              <span>{item.upperName || "未知 UP"}</span>
              <Show when={item.downloadKind !== "cover"}>
                <span>{item.bvid}</span>
              </Show>
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
            <p class="mt-2 truncate text-xs text-base-content/45" title={item.path}>{item.path}</p>
          </div>
        </article>
    )
  }

  return (
      <div class="flex h-full flex-col p-4">
        <section class="mb-3 flex items-center justify-between rounded-lg border border-base-300 bg-base-100 px-4 py-3">
          <div>
            <h2 class="text-base font-bold">下载历史</h2>
            <p class="text-sm text-base-content/60">
              <Show when={searchValue().trim()} fallback={<>已记录 {total()} 个视频</>}>
                匹配 {total()} 个视频
              </Show>
            </p>
          </div>
          <input type="text"
                 placeholder="输入UP主名字或视频标题模糊搜索"
                 class="input"
                 value={searchValue()}
                 onInput={(e) => searchInput(e.currentTarget.value)}
          />
          <div class="flex items-center gap-2">
            <button
                class="btn btn-outline btn-sm"
                type="button"
                onClick={() => void loadPage(true)}
                disabled={loading()}
            >
              {loading() ? "刷新中..." : "刷新"}
            </button>
            <button
                class="btn btn-outline btn-error btn-sm"
                type="button"
                onClick={() => void clearHistory()}
                disabled={loading() || (total() === 0 && items().length === 0)}
            >
              删除全部
            </button>
          </div>
        </section>

        <section class="min-h-0 flex-1 overflow-y-auto rounded-lg border border-base-300 bg-base-100">
          <Switch>
            <Match when={!loaded()}>
              <div class="flex h-full items-center justify-center">
                <span class="loading loading-spinner loading-md text-primary"/>
              </div>
            </Match>
            <Match when={errorMsg() && items().length === 0}>
              <DetailError message={errorMsg()} onRetry={() => void loadPage(true)}/>
            </Match>
            <Match when={items().length === 0}>
              <div class="flex h-full items-center justify-center text-sm text-base-content/50">
                {searchValue().trim() ? "没有匹配的下载历史" : "暂无下载历史"}
              </div>
            </Match>
            <Match when={items().length > 0}>
              <div class="divide-y divide-base-200">
                <For each={items()}>
                  {
                    (item: DownloadHistoryItem): JSXElement => (
                        <Card {...item}/>
                    )
                  }
                </For>
              </div>
              <Show when={hasMore()}>
                <div class="flex items-center justify-center p-4">
                  <button class="btn btn-outline btn-sm" type="button" disabled={loading()}
                          onClick={() => void loadPage(false)}>
                    {loading() ? "加载中..." : `加载更多（已加载 ${items().length}/${total()}）`}
                  </button>
                </div>
              </Show>
            </Match>
          </Switch>
        </section>

        <Toast message={message()} type={type()}/>
      </div>
  )
}
