import {createSignal, onCleanup} from "solid-js";
import {DownloadVideos} from "../../wailsjs/go/api/Douyin";
import {api} from "../../wailsjs/go/models";
import {EventsOn} from "../../wailsjs/runtime";
import {type DouyinDownloadItem, douyinVideoList, removeDouyinVideo} from "./douyinStore.ts";

type ToastType = "error" | "success" | "info" | "warning";
type ShowToast = (message: string, type?: ToastType) => void;

export type DouyinDownloadPhase = "video" | "image" | "sleep" | "music"|"done" | "error";

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

type BackendTask = api.DouyinDownloadTask;

const [downloading, setDownloading] = createSignal(false);
const [downloadingByID, setDownloadingByID] = createSignal<Record<string, boolean>>({});
const [progressByID, setProgressByID] = createSignal<Record<string, DouyinDownloadProgress>>({});

let progressListenerReady = false;
let activeToast: ShowToast | null = null;

function notify(message: string, type?: ToastType): void {
  activeToast?.(message, type);
}

function ensureProgressListener(): void {
  if (progressListenerReady) return;
  progressListenerReady = true;

  EventsOn("douyin-download-progress", (payload: DouyinDownloadProgress) => {
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

export function formatDouyinSource(item: DouyinDownloadItem): string {
  if (!item.sourceName || item.sourceName === item.sourceKind) {
    return item.sourceKind || "抖音";
  }
  return `${item.sourceKind}「${item.sourceName}」`;
}

function hasDownloadURL(item: DouyinDownloadItem): boolean {
  // 普通视频走 videoURL；图文/动图走有序素材列表，配乐是附加资源。
  if (item.mediaBadge) return (item.assets?.length ?? 0) > 0 || (item.imageURLs?.length ?? 0) > 0;
  return !!item.videoURL;
}

function toBackendTask(item: DouyinDownloadItem): BackendTask {
  // 前端只提交最终选择好的下载地址，后端不再重新解析清晰度，行为和 B 站下载队列保持一致。
  return api.DouyinDownloadTask.createFrom({
    awemeId: item.awemeId,
    sourceKind: item.sourceKind,
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

export function useDouyinDownloadQueue(showToast: ShowToast) {
  ensureProgressListener();
  activeToast = showToast;

  onCleanup(() => {
    if (activeToast === showToast) {
      activeToast = null;
    }
  });

  const listSourceSummary = () => {
    const labels = [...new Set(douyinVideoList().map(formatDouyinSource).filter(Boolean))];
    if (labels.length === 0) return "";
    if (labels.length === 1) return `来源：${labels[0]}`;
    return `来源：${labels.join("、")}`;
  };

  // 后端按 awemeId 推送实时进度；成功后该条会被移出列表，失败则留在列表供用户重试。
  function buildTasks(items: DouyinDownloadItem[]): BackendTask[] {
    const seen = new Set<string>();
    return items
      .filter((item) => {
        const key = item.awemeId.trim();
        // 同一 awemeId 在同一批次只提交一次，避免并发下载同一个文件。
        if (!key || seen.has(key) || !hasDownloadURL(item)) return false;
        seen.add(key);
        return true;
      })
      .map(toBackendTask);
  }

  async function runTasks(items: DouyinDownloadItem[]): Promise<{ success: number; failed: number }> {
    const tasks = buildTasks(items);
    if (tasks.length === 0) {
      notify("暂无可用下载地址，请稍后重试", "warning");
      return {success: 0, failed: items.length};
    }

    for (const task of tasks) {
      setDownloadingByID((prev) => ({...prev, [task.awemeId]: true}));
    }

    try {
      const batch = await DownloadVideos(tasks);
      const byID = new Map(items.map((item) => [item.awemeId, item]));

      // 和 B 站一致：成功项自动移除，失败项保留并展示错误。
      for (const result of batch.results ?? []) {
        const item = byID.get(result.awemeId);
        if (!item) continue;

        if (result.error) {
          notify(`下载失败：${item.title}，${result.error}`, "error");
        } else {
          removeDouyinVideo(item.awemeId);
        }
      }

      return {success: batch.success ?? 0, failed: batch.failed ?? 0};
    } catch (error) {
      notify(error instanceof Error ? error.message : String(error), "error");
      return {success: 0, failed: tasks.length};
    } finally {
      for (const task of tasks) {
        setDownloadingByID((prev) => ({...prev, [task.awemeId]: false}));
        setProgressByID((prev) => {
          const next = {...prev};
          delete next[task.awemeId];
          return next;
        });
      }
      setDownloading(false);
    }
  }

  async function startDownload(items = douyinVideoList()): Promise<void> {
    if (downloading()) return;
    if (items.length === 0) {
      notify("暂无可下载内容", "warning");
      return;
    }

    setDownloading(true);
    const {success, failed} = await runTasks(items);
    if (failed === 0) {
      notify(`下载完成：成功 ${success} 个`, "success");
      return;
    }
    notify(`下载完成：成功 ${success} 个，失败 ${failed} 个`, "warning");
  }

  async function downloadOne(item: DouyinDownloadItem): Promise<void> {
    if (downloading() || downloadingByID()[item.awemeId]) return;
    setDownloading(true);
    const {failed} = await runTasks([item]);
    if (failed === 0) {
      notify(`下载完成：${item.title}`, "success");
    }
  }

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
    downloadOne,
    downloading,
    isDownloading,
    progressFor,
    listSourceSummary,
    startDownload,
  };
}
