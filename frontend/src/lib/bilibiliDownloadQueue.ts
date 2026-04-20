import {createEffect, createMemo, createSignal, onCleanup} from "solid-js";
import {DownloadVideosByDash} from "../../wailsjs/go/api/BiliBili";
import {api} from "../../wailsjs/go/models";
import {EventsOn} from "../../wailsjs/runtime";
import {
  BilibiliPlayResolveError,
  bilibiliPlayResolveKey,
  resolveBilibiliPlayUrl,
  type ResolvedPlayInfo,
  streamBaseUrl,
  switchResolvedAudio,
  switchResolvedPlayAtQn,
  type VideoAccessInfo,
} from "./bilibiliPlayResolve.ts";
import {removeVideoAfterDownloadSuccess, videoList} from "./bilibiliStore.ts";
import type {MediaCardItem} from "./model.ts";

type ToastType = "error" | "success" | "info" | "warning";
type ShowToast = (message: string, type?: ToastType) => void;

export type PlayResolveEntry =
  | { status: "loading" }
  | { status: "done"; data: ResolvedPlayInfo }
  | { status: "error"; message: string; accessInfo?: VideoAccessInfo };

export type DownloadPhase = "video" | "audio" | "merge" | "done" | "error";

export interface DownloadProgress {
  bvid: string;
  cid?: number;
  title: string;
  phase: DownloadPhase;
  downloaded: number;
  total: number;
  percent: number;
}

interface DownloadTask {
  item: MediaCardItem;
  videoURL: string;
  audioURL: string;
}

type DashDownloadTask = api.DashDownloadTask;

// 解析播放地址是异步副作用。这个 Set 防止 Solid effect 重跑时对同一个 BV 重复发起解析请求。
const playResolveInFlight = new Set<string>();

function containsPlayKey(key: string): boolean {
  return videoList().some((v) => bilibiliPlayResolveKey(v) === key);
}

// 下载目录名的生成规则：
// - 如果来源是「全部投稿」，目录名为 UP 主名字，方便用户区分不同 UP 的视频；
// - 其他来源（合集、收藏夹等）则沿用原来的来源名称，保持和列表展示一致。
function downloadDirName(item: MediaCardItem): string {
  if (item.sourceListKind === "全部投稿") {
    return item.upperName;
  }
  if (item.sourceListKind === "解析结果") {
    return "";
  }

  return item.sourceListName ?? "";
}

export function formatListSource(item: MediaCardItem): string {
  const kind = item.sourceListKind;
  const name = item.sourceListName;
  if (kind === "全部投稿") {
    return kind;
  }
  if (kind === "分P") {
    return `分P「${name}」`;
  }

  return `${kind}「${name}」`;
}

export function useBilibiliDownloadQueue(showToast: ShowToast) {
  // playResolveByBvid：每个 BV 的 DASH 解析状态，卡片会根据它显示 loading/error/画质音质选择器。
  const [playResolveByBvid, setPlayResolveByBvid] = createSignal<Record<string, PlayResolveEntry>>({});
  // downloading 是整批下载的全局锁；downloadingByBvid 用来控制单张卡片的按钮和进度条。
  const [downloading, setDownloading] = createSignal<boolean>(false);
  const [downloadingByBvid, setDownloadingByBvid] = createSignal<Record<string, boolean>>({});
  const [progressByBvid, setProgressByBvid] = createSignal<Record<string, DownloadProgress>>({});

  const listSourceSummary = createMemo(() => {
    const labels = [...new Set(videoList().map((i) => formatListSource(i)).filter(Boolean))];
    if (labels.length === 0) return "";
    if (labels.length === 1) return `来源：${labels[0]}`;
    return `来源：${labels.join("、")}`;
  });

  createEffect(() => {
    // 这个 effect 是下载页的核心入口：
    // 只要全局 videoList 变化，就自动清理已移除视频的解析状态，并为新增视频解析 DASH。
    const list = videoList();
    const bvSet = new Set(
      list.map((i) => bilibiliPlayResolveKey(i)).filter((k): k is string => !!k),
    );

    setPlayResolveByBvid((prev) => {
      const next: Record<string, PlayResolveEntry> = {...prev};
      for (const k of Object.keys(next)) {
        if (!bvSet.has(k)) {
          // 视频被用户移出列表后，解析缓存和 in-flight 标记都要同步清掉。
          playResolveInFlight.delete(k);
          delete next[k];
        }
      }
      return next;
    });

    const map = playResolveByBvid();

    for (const item of list) {
      const key = bilibiliPlayResolveKey(item);
      if (!key || playResolveInFlight.has(key)) continue;

      const cur = map[key];
      if (cur?.status === "loading" || cur?.status === "done" || cur?.status === "error") {
        // done/error 都先保留，避免用户只是滚动或列表重渲染时重新打播放地址接口。
        continue;
      }

      playResolveInFlight.add(key);
      setPlayResolveByBvid((p) => ({...p, [key]: {status: "loading"}}));

      void (async () => {
        try {
          // resolveBilibiliPlayUrl 内部只请求一次 playurl；画质/音质切换复用返回的 dash 字段。
          const data = await resolveBilibiliPlayUrl(item.bvid, {cid: item.cid});
          if (!containsPlayKey(key)) return;
          setPlayResolveByBvid((p) => ({...p, [key]: {status: "done", data}}));
        } catch (e) {
          if (!containsPlayKey(key)) return;
          const message = e instanceof Error ? e.message : String(e);
          const accessInfo = e instanceof BilibiliPlayResolveError ? e.accessInfo : undefined;
          setPlayResolveByBvid((p) => ({...p, [key]: {status: "error", message, accessInfo}}));
        } finally {
          playResolveInFlight.delete(key);
        }
      })();
    }
  });

  // 后端批量下载期间会持续推送事件；这里按 BV 归档，供每张卡片独立渲染进度条。
  const offProgress = EventsOn("bilibili-download-progress", (payload: DownloadProgress) => {
    const key = bilibiliPlayResolveKey({bvid: payload?.bvid, cid: payload?.cid});
    if (!key) return;

    setProgressByBvid((p) => ({
      ...p,
      [key]: {
        ...payload,
        percent: Math.max(0, Math.min(100, Number(payload.percent) || 0)),
      },
    }));
  });
  onCleanup(offProgress);

  // 后续 UI 操作都只通过 BV 读写解析状态，避免同一个视频在不同来源列表里 id 不一致。
  function entryForItem(item: MediaCardItem): PlayResolveEntry | undefined {
    const key = bilibiliPlayResolveKey(item);
    return key ? playResolveByBvid()[key] : undefined;
  }

  function handlePickQn(item: MediaCardItem, qn: number): void {
    const key = bilibiliPlayResolveKey(item);
    if (!key) return;

    const cur = playResolveByBvid()[key];
    if (cur?.status !== "done" || qn === cur.data.selectedQn) return;

    const next = switchResolvedPlayAtQn(cur.data, qn);
    if (!next) {
      // 理论上选择器只会列出 dash 里已有的 qn；这里是防御式提示。
      showToast("当前 DASH 数据中没有这个画质的流地址", "warning");
      return;
    }

    setPlayResolveByBvid((p) => ({...p, [key]: {status: "done", data: next}}));
  }

  function handlePickAudio(item: MediaCardItem, audioId: number): void {
    const key = bilibiliPlayResolveKey(item);
    if (!key) return;

    const cur = playResolveByBvid()[key];
    if (cur?.status !== "done" || cur.data.bestAudio?.id === audioId) return;

    const next = switchResolvedAudio(cur.data, audioId);
    if (!next) {
      showToast("当前 DASH 数据中没有这个音轨的流地址", "warning");
      return;
    }

    setPlayResolveByBvid((p) => ({...p, [key]: {status: "done", data: next}}));
  }

  function buildDownloadTasks(items: MediaCardItem[]): DownloadTask[] {
    const seenKeys = new Set<string>();
    return items
      .map((item) => {
        // 同一解析键（BV 或 BV+cid）不重复提交；后端也会去重。
        const key = bilibiliPlayResolveKey(item);
        if (key) {
          if (seenKeys.has(key)) return null;
          seenKeys.add(key);
        }

        const entry = entryForItem(item);
        if (entry?.status !== "done") return null;

        // 这里已经是用户最终选择的视频/音频流地址；后端不再重新解析画质。
        const videoURL = streamBaseUrl(entry.data.bestVideo);
        if (!videoURL) return null;

        const audioURL = entry.data.bestAudio ? streamBaseUrl(entry.data.bestAudio) : "";
        return {item, videoURL, audioURL};
      })
      .filter((v): v is DownloadTask => v !== null);
  }

  // 后端批量接口只需要稳定的下载参数；目录名仍沿用原来的来源规则。
  function toBackendTask(task: DownloadTask): DashDownloadTask {
    return {
      sourceName: downloadDirName(task.item),
      sourceKind: task.item.sourceListKind ?? "",
      upperName: task.item.upperName ?? "",
      bvid: task.item.bvid,
      cid: task.item.cid ?? 0,
      title: task.item.title,
      cover: task.item.cover ?? "",
      duration: task.item.duration ?? 0,
      play: task.item.play ?? 0,
      danmaku: task.item.danmaku ?? 0,
      pubtime: task.item.pubtime ?? 0,
      videoURL: task.videoURL,
      audioURL: task.audioURL,
    };
  }

  async function runDownloadTasks(tasks: DownloadTask[]): Promise<{ success: number; failed: number }> {
    // 所有待提交任务先置为下载中；真实字节进度由后端事件逐条刷新。
    for (const task of tasks) {
      const key = bilibiliPlayResolveKey(task.item);
      if (key) {
        setDownloadingByBvid((p) => ({...p, [key]: true}));
      }
    }

    try {
      // 真正的并发下载、休眠控制、缓存判断都在后端完成；前端只提交任务列表并等待最终结果。
      const batch = await DownloadVideosByDash(tasks.map(toBackendTask));
      const byKey = new Map(tasks.map((task) => [bilibiliPlayResolveKey(task.item), task.item]));

      // 后端返回每条任务的最终结果，前端只移除成功项，失败项保留给用户重试。
      for (const item of batch.results ?? []) {
        const key = bilibiliPlayResolveKey({bvid: item.bvid, cid: item.cid});
        const media = key ? byKey.get(key) : undefined;
        if (!media) continue;

        if (item.error) {
          showToast(`下载失败：${media.title}，${item.error}`, "error");
        } else {
          removeVideoAfterDownloadSuccess(media.bvid, item.cid);
        }
      }

      return {success: batch.success ?? 0, failed: batch.failed ?? 0};
    } catch (e) {
      showToast(e instanceof Error ? e.message : String(e), "error");
      return {success: 0, failed: tasks.length};
    } finally {
      // 批量调用结束后清理按钮态；失败项仍留在列表，但进度条回到待下载状态。
      for (const task of tasks) {
        const key = bilibiliPlayResolveKey(task.item);
        if (!key) continue;
        setDownloadingByBvid((p) => ({...p, [key]: false}));
        setProgressByBvid((p) => {
          const next = {...p};
          delete next[key];
          return next;
        });
      }
      setDownloading(false);
    }
  }

  async function startDownload(items = videoList()): Promise<void> {
    // 默认下载当前列表里的全部视频；单个卡片下载会传入只含一个 item 的数组。
    if (downloading()) return;
    if (items.length === 0) {
      showToast("暂无可下载视频", "warning");
      return;
    }

    const tasks = buildDownloadTasks(items);
    if (tasks.length === 0) {
      showToast("暂无可用流地址，请稍候重试", "warning");
      return;
    }

    setDownloading(true);
    const {success, failed} = await runDownloadTasks(tasks);

    if (failed === 0) {
      showToast(`下载完成：成功 ${success} 个`, "success");
      return;
    }
    showToast(`下载完成：成功 ${success} 个，失败 ${failed} 个`, "warning");
  }

  async function downloadOne(item: MediaCardItem): Promise<void> {
    const key = bilibiliPlayResolveKey(item);
    if (downloading() || (key && downloadingByBvid()[key])) return;

    const entry = entryForItem(item);
    if (entry?.status === "loading") {
      showToast("视频还在解析中，请稍候", "warning");
      return;
    }
    if (entry?.status === "error") {
      showToast(`解析失败：${entry.message}`, "error");
      return;
    }

    const tasks = buildDownloadTasks([item]);
    if (tasks.length === 0) {
      showToast("暂无可用流地址，请稍候重试", "warning");
      return;
    }

    setDownloading(true);
    const {failed} = await runDownloadTasks(tasks);
    if (failed === 0) {
      showToast(`下载完成：${item.title}`, "success");
    }
  }

  function canDownload(item: MediaCardItem): boolean {
    return entryForItem(item)?.status === "done";
  }

  function isDownloading(item: MediaCardItem): boolean {
    const key = bilibiliPlayResolveKey(item);
    return key ? downloadingByBvid()[key] : false;
  }

  function progressFor(item: MediaCardItem): DownloadProgress | undefined {
    const key = bilibiliPlayResolveKey(item);
    return key ? progressByBvid()[key] : undefined;
  }

  return {
    canDownload,
    downloadOne,
    downloading,
    entryForItem,
    handlePickAudio,
    handlePickQn,
    isDownloading,
    listSourceSummary,
    progressFor,
    startDownload,
  };
}
