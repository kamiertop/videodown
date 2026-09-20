import {type Task} from "@bindings/github.com/kamiertop/videodown/douyin/download"
import {DownloadVideos} from "@bindings/github.com/kamiertop/videodown/douyin/download/service.ts";
import {Events} from "@wailsio/runtime";
import {createSignal, onCleanup} from "solid-js";
import {type DouyinDownloadItem, douyinVideoList, removeDouyinVideo} from "./store.ts";

type ToastType = "error" | "success" | "info" | "warning";
type ShowToast = (message: string, type?: ToastType) => void;

export type DouyinDownloadPhase = "video" | "image" | "sleep" | "music" | "done" | "error";

export interface DouyinDownloadProgress {
  awemeId: string;
  title: string;
  phase: DouyinDownloadPhase;
  downloaded: number;
  total: number;
  percent: number;
  sleepRemaining?: number;
  sleepTotal?: number;
}

export interface DouyinDownloadRunResult {
  success: number;
  failed: number;
  failedItems: DouyinDownloadItem[];
}

// downloading 是全局独占锁：手动“开始下载”、单个下载和批量会话共用，
// 保证同一时刻只有一个下载流程在跑，避免同一条视频被并发写盘。
export const [downloading, setDownloading] = createSignal(false);
const [downloadingByID, setDownloadingByID] = createSignal<Record<string, boolean>>({});
const [progressByID, setProgressByID] = createSignal<Record<string, DouyinDownloadProgress>>({});

let progressListenerReady = false;
let activeToast: ShowToast | null = null;

function notify(message: string, type?: ToastType): void {
  activeToast?.(message, type);
}

/** 供批量会话等模块级调用方复用下载页注册的 toast 通道。 */
export function notifyDouyinDownload(message: string, type?: ToastType): void {
  notify(message, type);
}

function ensureProgressListener(): void {
  if (progressListenerReady) return;
  progressListenerReady = true;

  Events.On("douyin-download-progress", ({data: payload}) => {
    if (!payload?.awemeId) return;
    setProgressByID((prev) => ({
      ...prev,
      [payload.awemeId]: {
        ...payload,
        percent: Math.max(0, Math.min(100, Number(payload.percent) || 0)),
      },
    }));
  });
}

function hasDownloadURL(item: DouyinDownloadItem): boolean {
  // 普通视频走 videoURL；图文/动图走有序素材列表，配乐是附加资源。
  if (item.mediaBadge) return (item.assets?.length ?? 0) > 0 || (item.imageURLs?.length ?? 0) > 0;
  return !!item.videoURL;
}

function toTask(item: DouyinDownloadItem): Task {
  // 前端只提交最终选择好地下载地址和来源元数据；落盘目录由后端统一判断。
  return ({
    awemeId: item.awemeId,
    sourceName: item.sourceName ?? "",
    title: item.title,
    cover: item.cover,
    duration: item.duration,
    authorName: item.authorName,
    publishTime: item.publishTime ?? 0,
    diggCount: item.diggCount ?? 0,
    collectCount: item.collectCount ?? 0,
    videoURL: item.videoURL ?? "",
    imageURLs: item.mediaBadge === "image" ? item.imageURLs ?? [] : [],
    assets: item.mediaBadge ? item.assets ?? [] : [],
    musicURL: item.mediaBadge ? item.musicURL ?? "" : "",
  });
}

// 后端按 awemeId 推送实时进度；成功后该条会被移出列表，失败则留在列表供用户重试。
function buildTasks(items: readonly DouyinDownloadItem[]): Task[] {
  const seen = new Set<string>();
  return items
      .filter((item) => {
        const key = item.awemeId.trim();
        // 同一 awemeId 在同一批次只提交一次，避免并发下载同一个文件。
        if (!key || seen.has(key) || !hasDownloadURL(item)) return false;
        seen.add(key);
        return true;
      })
      .map(toTask);
}

/**
 * 执行一轮下载，不含自动重试，也不持有全局锁：成功项移出列表，失败项留在返回值里
 * 由调用方决定去留。批量会话用它逐批推进，失败项攒到最后统一重试。
 */
export async function runDouyinDownloadTasks(items: readonly DouyinDownloadItem[]): Promise<DouyinDownloadRunResult> {
  const tasks = buildTasks(items);
  if (tasks.length === 0) {
    if (items.length > 0) notify("暂无可用下载地址，请稍后重试", "warning");
    return {success: 0, failed: items.length, failedItems: [...items]};
  }

  for (const task of tasks) {
    setDownloadingByID((prev) => ({...prev, [task.awemeId]: true}));
  }

  try {
    const batch = await DownloadVideos(tasks);
    // 后端返回的 awemeId 可能带有首尾空白；统一按规范化 ID 匹配，
    // 否则成功结果虽已返回，前端会因 Map 未命中而无法移除列表项。
    const byID = new Map(items.map((item) => [item.awemeId.trim(), item]));
    const failedItems: DouyinDownloadItem[] = [];

    // 和 B 站一致：成功项自动移除，失败项保留并展示错误。
    for (const result of batch.results ?? []) {
      const resultID = String(result.awemeId ?? "").trim();
      const item = byID.get(resultID);
      if (!item) continue;

      if (result.error) {
        notify(`下载失败：${item.title}，${result.error}`, "error");
        failedItems.push(item);
      } else {
        removeDouyinVideo(resultID);
      }
    }

    return {success: batch.success ?? 0, failed: batch.failed ?? 0, failedItems};
  } catch (error) {
    notify(error instanceof Error ? error.message : String(error), "error");
    return {success: 0, failed: tasks.length, failedItems: [...items]};
  } finally {
    for (const task of tasks) {
      setDownloadingByID((prev) => ({...prev, [task.awemeId]: false}));
      setProgressByID((prev) => {
        const next = {...prev};
        delete next[task.awemeId];
        return next;
      });
    }
  }
}

/**
 * 获取全局下载锁后执行 fn；锁被占用时直接返回 undefined。
 * 批量会话全程持锁，手动按钮因此在整个批量期间保持禁用。
 */
export async function withDouyinDownloadLock<T>(fn: () => Promise<T>): Promise<T | undefined> {
  if (downloading()) return undefined;
  setDownloading(true);
  try {
    return await fn();
  } finally {
    setDownloading(false);
  }
}

/** 手动“开始下载”入口：一轮下载后对失败项立即补一轮重试。 */
export async function startDouyinDownloadQueue(items: readonly DouyinDownloadItem[] = douyinVideoList()): Promise<number> {
  const result = await withDouyinDownloadLock(async () => {
    if (items.length === 0) {
      notify("暂无可下载内容", "warning");
      return 0;
    }

    let {success, failed, failedItems} = await runDouyinDownloadTasks(items);
    if (failedItems.length > 0) {
      notify(`有 ${failedItems.length} 个内容下载失败，其他任务完成后将自动重试`, "warning");
      const retry = await runDouyinDownloadTasks(failedItems);
      success += retry.success;
      failed = retry.failed;
    }
    if (failed === 0) {
      notify(`下载完成：成功 ${success} 个`, "success");
      return success;
    }
    notify(`下载完成：成功 ${success} 个，失败 ${failed} 个`, "warning");
    return success;
  });
  return result ?? 0;
}

async function downloadOneItem(item: DouyinDownloadItem): Promise<number> {
  if (downloadingByID()[item.awemeId]) return 0;
  const result = await withDouyinDownloadLock(async () => {
    let {success, failed, failedItems} = await runDouyinDownloadTasks([item]);
    if (failedItems.length > 0) {
      notify("下载失败，正在自动重试", "warning");
      const retry = await runDouyinDownloadTasks(failedItems);
      success += retry.success;
      failed = retry.failed;
    }
    if (failed === 0) {
      notify(`下载完成：${item.title}`, "success");
      return 1;
    }
    return 0;
  });
  return result ?? 0;
}

export function useDouyinDownloadQueue(showToast: ShowToast) {
  ensureProgressListener();
  activeToast = showToast;

  onCleanup(() => {
    if (activeToast === showToast) {
      activeToast = null;
    }
  });

  function canDownload(item: DouyinDownloadItem): boolean {
    return hasDownloadURL(item);
  }

  function isDownloading(item: DouyinDownloadItem): boolean {
    return downloadingByID()[item.awemeId];
  }

  function progressFor(item: DouyinDownloadItem): DouyinDownloadProgress | undefined {
    return progressByID()[item.awemeId];
  }

  return {
    canDownload,
    downloadOne: downloadOneItem,
    downloading,
    isDownloading,
    progressFor,
    startDownload: startDouyinDownloadQueue,
  };
}
