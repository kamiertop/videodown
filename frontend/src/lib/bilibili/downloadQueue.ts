import {createEffect, createSignal, onCleanup} from "solid-js";
import {BatchResolvePlayUrl, DownloadVideosByDash} from "../../../wailsjs/go/api/BiliBili";
import {api} from "../../../wailsjs/go/models";
import {EventsOn} from "../../../wailsjs/runtime";
import type {MediaCardItem} from "../model.ts";
import {
  bilibiliPlayResolveKey,
  buildResolvedPlayInfo,
  type ResolvedPlayInfo,
  shouldSkipPlayUrl,
  streamBaseUrl,
  switchResolvedAudio,
  switchResolvedPlayAtQn,
  type VideoAccessInfo,
  videoAccessInfoFromView,
} from "./playResolve.ts";
import {removeVideoAfterDownloadSuccess, videoList} from "./store.ts";

type PlayUrlRequest = api.PlayUrlRequest;
type PlayUrlResult = api.PlayUrlResult;

type ToastType = "error" | "success" | "info" | "warning";
type ShowToast = (message: string, type?: ToastType) => void;

export type PlayResolveEntry =
    | { status: "loading" }
    | { status: "done"; data: ResolvedPlayInfo }
    | { status: "error"; message: string; accessInfo?: VideoAccessInfo };

export type DownloadPhase = "video" | "audio" | "merge" | "sleep" | "done" | "error";

export interface DownloadProgress {
  bvid: string;
  cid?: number;
  title: string;
  phase: DownloadPhase;
  downloaded: number;
  total: number;
  percent: number;
  sleepRemaining?: number;
  sleepTotal?: number;
}

interface DownloadTask {
  item: MediaCardItem;
  cid: number;
  bvid: string;
  uiKey: string | null;
  backendKey: string | null;
  videoURL: string;
  audioURL: string;
}

type DashDownloadTask = api.DashDownloadTask;

// 解析播放地址是异步副作用。这个 Set 防止 Solid effect 重跑时对同一个 BV 重复发起解析请求。
const playResolveInFlight = new Set<string>();
const [playResolveByBvid, setPlayResolveByBvid] = createSignal<Record<string, PlayResolveEntry>>({});
const [downloading, setDownloading] = createSignal<boolean>(false);
const [downloadingByBvid, setDownloadingByBvid] = createSignal<Record<string, boolean>>({});
const [progressByBvid, setProgressByBvid] = createSignal<Record<string, DownloadProgress>>({});

let progressListenerReady = false;
let activeToast: ShowToast | null = null;

function notify(message: string, type?: ToastType): void {
  activeToast?.(message, type);
}

function ensureProgressListener(): void {
  if (progressListenerReady) return;
  progressListenerReady = true;

  EventsOn("bilibili-download-progress", (payload: DownloadProgress) => {
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
}

function containsPlayKey(key: string): boolean {
  return videoList().some((v) => bilibiliPlayResolveKey(v) === key);
}

export function useBilibiliDownloadQueue(showToast: ShowToast) {
  ensureProgressListener();
  activeToast = showToast;

  onCleanup(() => {
    if (activeToast === showToast) {
      activeToast = null;
    }
  });
  // playResolveByBvid：每个 BV 的 DASH 解析状态，卡片会根据它显示 loading/error/画质音质选择器。
  // downloading 是整批下载的全局锁；downloadingByBvid 用来控制单张卡片的按钮和进度条。

  createEffect(() => {
    // 只要全局 videoList 变化，就清理已移除视频的解析状态，并为新增视频批量解析 DASH。
    const list = videoList();
    const bvSet = new Set(
        list.map((i) => bilibiliPlayResolveKey(i)).filter((k): k is string => !!k),
    );

    setPlayResolveByBvid((prev) => {
      const next: Record<string, PlayResolveEntry> = {...prev};
      for (const k of Object.keys(next)) {
        if (!bvSet.has(k)) {
          playResolveInFlight.delete(k);
          delete next[k];
        }
      }
      return next;
    });

    const map = playResolveByBvid();

    // 收集所有待解析的视频
    const pending: Array<{ item: MediaCardItem; key: string }> = [];
    for (const item of list) {
      const key = bilibiliPlayResolveKey(item);
      if (!key || playResolveInFlight.has(key)) continue;

      const cur = map[key];
      if (cur?.status === "loading" || cur?.status === "done" || cur?.status === "error") {
        continue;
      }

      playResolveInFlight.add(key);
      setPlayResolveByBvid((p) => ({...p, [key]: {status: "loading"}}));
      pending.push({item, key});
    }

    if (pending.length === 0) return;

    void (async () => {
      try {
        const requests = pending.map(({item}) => ({
          bvid: item.bvid!,
          cid: item.cid || 0,
          qn: 0,
        } as PlayUrlRequest));
        const results = await BatchResolvePlayUrl(requests);

        for (let i = 0; i < pending.length; i += 1) {
          const {item, key} = pending[i];
          const result: PlayUrlResult | undefined = results?.[i];

          if (!containsPlayKey(key)) continue;

          if (!result) {
            setPlayResolveByBvid((p) => ({...p, [key]: {status: "error", message: "解析接口未返回结果"}}));
            continue;
          }
          if (result.error) {
            setPlayResolveByBvid((p) => ({...p, [key]: {status: "error", message: result.error || "未知错误"}}));
            continue;
          }
          if (!result.detail?.view || !result.play_url) {
            setPlayResolveByBvid((p) => ({...p, [key]: {status: "error", message: "解析接口返回数据不完整"}}));
            continue;
          }

          const view = result.detail.view;
          const accessInfo = videoAccessInfoFromView(view);

          if (shouldSkipPlayUrl(accessInfo)) {
            setPlayResolveByBvid((p) => ({
              ...p,
              [key]: {
                status: "error",
                message: "充电专属视频，当前账号不可直接下载",
                accessInfo,
              },
            }));
            continue;
          }

          const aid = Number(view.aid);
          const cid = item.cid && item.cid > 0 ? item.cid : Number(view.cid);
          const bvid = view.bvid?.trim() || item.bvid!;
          const partCount = Math.max(1, Number(view.videos || view.pages?.length || 1));

          const data = buildResolvedPlayInfo({aid, cid, bvid, partCount}, result.play_url, accessInfo);
          setPlayResolveByBvid((p) => ({...p, [key]: {status: "done", data}}));
        }
      } catch (e) {
        const message = e instanceof Error ? (e.message ?? String(e)) : String(e);
        for (const {key} of pending) {
          if (!containsPlayKey(key)) continue;
          setPlayResolveByBvid((p) => ({...p, [key]: {status: "error", message}}));
        }
      } finally {
        for (const {key} of pending) {
          playResolveInFlight.delete(key);
        }
      }
    })();
  });

  // 后端批量下载期间会持续推送事件；这里按 BV 归档，供每张卡片独立渲染进度条。
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
      notify("当前 DASH 数据中没有这个画质的流地址", "warning");
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
      notify("当前 DASH 数据中没有这个音轨的流地址", "warning");
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
          const backendKey = bilibiliPlayResolveKey({bvid: entry.data.bvid, cid: entry.data.cid});
          return {
            item,
            cid: entry.data.cid,
            bvid: entry.data.bvid,
            uiKey: key,
            backendKey,
            videoURL,
            audioURL,
          };
        })
        .filter((v): v is DownloadTask => v !== null);
  }

  // 后端批量接口只需要稳定的下载参数；目录规则由后端根据 kind/upperName/sourceName 统一判断。
  function toBackendTask(task: DownloadTask): DashDownloadTask {
    return {
      sourceName: task.item.sourceListName ?? "",
      sourceKind: task.item.sourceListKind ?? "",
      upperName: task.item.upperName ?? "",
      bvid: task.bvid,
      cid: task.cid,
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
      const key = task.uiKey;
      if (key) {
        setDownloadingByBvid((p) => ({...p, [key]: true}));
      }
    }

    try {
      // 真正的并发下载、休眠控制、缓存判断都在后端完成；前端只提交任务列表并等待最终结果。
      const batch = await DownloadVideosByDash(tasks.map(toBackendTask));
      const byKey = new Map(tasks.flatMap((task) => {
        const pairs: Array<[string, MediaCardItem]> = [];
        if (task.uiKey) pairs.push([task.uiKey, task.item]);
        if (task.backendKey) pairs.push([task.backendKey, task.item]);
        return pairs;
      }));

      // 后端返回每条任务的最终结果，前端只移除成功项，失败项保留给用户重试。
      for (const item of batch.results ?? []) {
        const key = bilibiliPlayResolveKey({bvid: item.bvid, cid: item.cid});
        const media = key ? byKey.get(key) : undefined;
        if (!media) continue;

        if (item.error) {
          notify(`下载失败：${media.title}，${item.error}`, "error");
        } else {
          removeVideoAfterDownloadSuccess(media.bvid, item.cid);
        }
      }

      return {success: batch.success ?? 0, failed: batch.failed ?? 0};
    } catch (e) {
      notify(e instanceof Error ? e.message : String(e), "error");
      return {success: 0, failed: tasks.length};
    } finally {
      // 批量调用结束后清理按钮态；失败项仍留在列表，但进度条回到待下载状态。
      for (const task of tasks) {
        const key = task.uiKey;
        if (key) {
          setDownloadingByBvid((p) => ({...p, [key]: false}));
        }
        setProgressByBvid((p) => {
          const next = {...p};
          if (key) delete next[key];
          if (task.backendKey) delete next[task.backendKey];
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
      notify("暂无可下载视频", "warning");
      return;
    }

    const tasks = buildDownloadTasks(items);
    if (tasks.length === 0) {
      notify("暂无可用流地址，请稍候重试", "warning");
      return;
    }

    setDownloading(true);
    const {success, failed} = await runDownloadTasks(tasks);

    if (failed === 0) {
      notify(`下载完成：成功 ${success} 个`, "success");
      return;
    }
    notify(`下载完成：成功 ${success} 个，失败 ${failed} 个`, "warning");
  }

  async function downloadOne(item: MediaCardItem): Promise<void> {
    const key = bilibiliPlayResolveKey(item);
    if (downloading() || (key && downloadingByBvid()[key])) return;

    const entry = entryForItem(item);
    if (entry?.status === "loading") {
      notify("视频还在解析中，请稍候", "warning");
      return;
    }
    if (entry?.status === "error") {
      notify(`解析失败：${entry.message}`, "error");
      return;
    }

    const tasks = buildDownloadTasks([item]);
    if (tasks.length === 0) {
      notify("暂无可用流地址，请稍候重试", "warning");
      return;
    }

    setDownloading(true);
    const {failed} = await runDownloadTasks(tasks);
    if (failed === 0) {
      notify(`下载完成：${item.title}`, "success");
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
    if (!key) return undefined;
    const progress = progressByBvid();
    if (progress[key]) return progress[key];
    return Object.entries(progress).find(([k]) => k.startsWith(`${key}:`))?.[1];
  }

  return {
    canDownload,
    downloadOne,
    downloading,
    entryForItem,
    handlePickAudio,
    handlePickQn,
    isDownloading,
    progressFor,
    startDownload,
  };
}
