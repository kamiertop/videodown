import {GetConcurrencyNum} from "@bindings/github.com/kamiertop/videodown/utils/settings";
import {createSignal} from "solid-js";
import {waitBulkDownloadPage} from "../bulkDownloadThrottle.ts";
import type {MediaCardItem} from "../model.ts";
import {bilibiliPlayResolveKey} from "./playResolve.ts";
import {
  downloading,
  notifyBilibiliDownload,
  refreshBilibiliPlayUrls,
  resetBilibiliPlayResolve,
  runBilibiliDownloadTasks,
  setBilibiliAutoResolve,
  withBilibiliDownloadLock,
} from "./downloadQueue.ts";
import {addVideos, removeVideo, videoList} from "./store.ts";

/**
 * 批量下载的翻页加载器：由来源页面在创建会话时构造。
 * 闭包必须自包含（只捕获纯值、直接调后端绑定），不引用组件 signal，
 * 这样跳转到下载页、来源组件卸载后仍能继续翻页。
 */
export interface BilibiliBatchPageLoader {
  loadNext(): Promise<readonly MediaCardItem[]>;
  hasMore(): boolean;
}

export interface BilibiliBatchSource {
  title: string;
  /** 接口已知的视频总数（UP 投稿数、收藏夹总数等）；未知则留空。 */
  totalCount?: number;
  /** 点击时来源页面已加载的全部视频。 */
  initialItems: readonly MediaCardItem[];
  /** 无翻页来源（如动态页已加载内容）可以不传。 */
  loader?: BilibiliBatchPageLoader;
}

export type BilibiliBatchStatus =
    | "resolving"
    | "downloading"
    | "loadingPage"
    | "retrySleep"
    | "retrying"
    | "done"
    | "stopped";

export interface BilibiliBatchState {
  title: string;
  totalCount?: number;
  /** 已加载入队的视频数；总数未知时作为分母展示。 */
  loadedCount: number;
  /** 已成功下载的视频数。 */
  downloadedCount: number;
  /** 当前失败（待重试或最终仍失败）的数量。 */
  failedCount: number;
  status: BilibiliBatchStatus;
  /** 状态的动态补充文案，例如休眠剩余秒数。 */
  statusText: string;
  /** 已请求停止，正在等待当前批结束。 */
  stopping: boolean;
  /** 会话结束后是否可以继续（被停止或加载中断，且翻页加载器仍可用）。 */
  canResume: boolean;
  active: boolean;
}

const [batchState, setBatchState] = createSignal<BilibiliBatchState | null>(null);

export {batchState as bilibiliBatchState};

let stopRequested = false;
// 会话序号：只允许最新会话写状态，防止旧会话的收尾代码覆盖新会话。
let sessionSeq = 0;
// 当前会话的翻页加载器：停止/中断后保留，供“继续批量”从断点恢复。
let sessionLoader: BilibiliBatchPageLoader | undefined;

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

function patchState(
    id: number,
    patch: Partial<BilibiliBatchState> | ((prev: BilibiliBatchState) => Partial<BilibiliBatchState>),
): void {
  if (id !== sessionSeq) return;
  setBatchState((prev) => {
    if (!prev) return prev;
    const resolved = typeof patch === "function" ? patch(prev) : patch;
    return {...prev, ...resolved};
  });
}

/**
 * 启动批量下载会话：控制器在后台逐块“解析地址 → 下载”推进并自动翻页，
 * 全部页完成后对失败项（下载失败 + 解析失败）统一重试。
 * 整组（页）入队展示——下载页列表和抖音一样显示当前批全部内容，
 * 尚未轮到的只是暂时没有播放地址；解析与下载仍按下载并发数分块推进，
 * 解析请求只比下载领先一小块，不会一次轰炸整页。
 * 返回 false 表示当前有下载任务占用（手动或批量）。
 */
export function startBilibiliBatch(source: BilibiliBatchSource): boolean {
  if (batchState()?.active || downloading()) return false;

  const initialItems = source.initialItems.filter((item) => item.bvid?.trim().length > 0);
  if (initialItems.length === 0 && !source.loader) return false;

  stopRequested = false;
  const id = ++sessionSeq;
  sessionLoader = source.loader;
  // 批量期间关闭下载页的自动全量解析，解析改由会话按块驱动。
  setBilibiliAutoResolve(false);
  // 首批立即整组入队展示；处理仍按块推进。
  addVideos(initialItems);
  setBatchState({
    title: source.title,
    totalCount: source.totalCount,
    loadedCount: initialItems.length,
    downloadedCount: 0,
    failedCount: 0,
    status: "resolving",
    statusText: "正在解析视频地址",
    stopping: false,
    canResume: false,
    active: true,
  });
  void runSession(id, source.loader, initialItems);
  return true;
}

/** 请求停止批量会话：当前批下载结束后停止，不再翻页和重试。 */
export function stopBilibiliBatch(): void {
  const state = batchState();
  if (!state?.active || state.stopping) return;
  stopRequested = true;
  // 立刻反馈“正在停止”：停止是协作式的，要等当前批（含任务后休眠）跑完才生效。
  setBatchState((prev) => prev ? {...prev, stopping: true, statusText: "正在停止，等待当前批结束…"} : prev);
}

/**
 * 继续已停止/中断的批量会话：复用保存的翻页加载器从断点继续（加载器内部
 * 保存着游标），列表中残留的失败项会作为首批重新解析下载。统计数字接着累计。
 */
export function resumeBilibiliBatch(): boolean {
  const state = batchState();
  if (!state || state.active || !state.canResume || !sessionLoader) return false;
  if (downloading()) {
    notifyBilibiliDownload("已有下载任务进行中，请稍后再试", "warning");
    return false;
  }

  stopRequested = false;
  const id = ++sessionSeq;
  setBilibiliAutoResolve(false);
  setBatchState((prev) => prev ? {
    ...prev,
    status: "resolving",
    statusText: "正在解析视频地址",
    stopping: false,
    canResume: false,
    active: true,
  } : prev);
  // 继续时列表里残留的是上次会话放回的失败项：清掉旧解析记录后作为首批重新分块处理，
  // 否则解析失败的旧 entry 会让重试直接跳过这些视频。
  const resumeQueue = videoList();
  resetBilibiliPlayResolve(resumeQueue);
  void runSession(id, sessionLoader, resumeQueue);
  return true;
}

/** 会话结束后清除状态卡片。 */
export function dismissBilibiliBatch(): void {
  if (!batchState()?.active) {
    sessionLoader = undefined;
    setBatchState(null);
  }
}

async function runSession(id: number, loader: BilibiliBatchPageLoader | undefined, initialItems: readonly MediaCardItem[]): Promise<void> {
  // 批内失败项（下载失败 + 解析失败，如充电专属）先收进这里并移出下载列表，
  // 保证列表里只有当前块的内容；全部页完成后再统一重试。
  const failedItems: MediaCardItem[] = [];
  const failedKeys = new Set<string>();
  let pageLoadError = "";

  try {

  // withBilibiliDownloadLock 在锁被占用时返回 undefined；startBilibiliBatch 已挡掉
  // 绝大多数场景，这里兜底处理启动瞬间的手动下载竞态。
  const ran = await withBilibiliDownloadLock(async () => {
    // 下载并发数即分块大小：解析只比下载领先一小块（下载远慢于解析），
    // 避免点击后一次性发出整页的解析请求。读取失败时退回单块单下。
    const chunkSize = Math.max(1, Number(await GetConcurrencyNum().catch(() => 1)) || 1);

    // 一组（页）视频整组入队展示：下载页列表与抖音一致显示当前批全部内容，
    // 尚未轮到的项只是暂时没有播放地址；解析与下载仍只按小块推进。
    const processQueue = async (items: readonly MediaCardItem[], retrying: boolean, sweepAfter: boolean): Promise<void> => {
      if (items.length === 0) return;
      addVideos(items);

      for (let start = 0; start < items.length; start += chunkSize) {
        if (stopRequested) return;
        const chunk = items.slice(start, start + chunkSize);

        // 下载时再解析：只解析当前小块（给 refreshBilibiliPlayUrls 传子集）。
        patchState(id, {status: "resolving", statusText: retrying ? "正在重新解析失败项" : "正在解析视频地址"});
        await refreshBilibiliPlayUrls(chunk);
        if (stopRequested) return;

        patchState(id, {status: retrying ? "retrying" : "downloading", statusText: retrying ? "正在重试失败项" : "正在下载当前批"});
        const run = await runBilibiliDownloadTasks(chunk);
        patchState(id, (prev) => ({downloadedCount: prev.downloadedCount + run.success}));
      }

      if (!sweepAfter) return;
      // 组后清场：本组里仍留在列表的就是失败项（下载失败 + 解析失败 + 无可用地址），
      // 收进失败列表并移出队列，下一组从干净列表开始，几千个视频也不会堆积卡片。
      const inList = new Set(videoList().map((item) => bilibiliPlayResolveKey(item)));
      for (const item of items) {
        const key = bilibiliPlayResolveKey(item);
        if (!key || failedKeys.has(key) || !inList.has(key)) continue;
        failedKeys.add(key);
        failedItems.push(item);
        removeVideo(item.id);
      }
      patchState(id, {failedCount: failedItems.length});
    };

    let queue: readonly MediaCardItem[] = initialItems;
    while (true) {
      await processQueue(queue, false, true);
      if (stopRequested) break;
      if (!loader || !loader.hasMore()) break;

      patchState(id, {status: "loadingPage", statusText: "正在加载下一页"});
      let nextPage: readonly MediaCardItem[];
      try {
        nextPage = await loader.loadNext();
      } catch (error) {
        pageLoadError = errorMessage(error);
        break;
      }
      // 接口还有下一页却返回空列表时按结束处理，防止死循环。
      if (nextPage.length === 0) break;

      queue = nextPage;
      patchState(id, (prev) => ({loadedCount: prev.loadedCount + nextPage.length}));
    }

    // 最终重试：清掉失败项的解析记录重新解析（解析失败可能只是临时网络问题），
    // 重试前按分页休眠设置缓一口气（设为 0 则立即重试）；失败项同样按块推进。
    // 重试不清场：仍失败的留在列表里，会话收尾本来就要放回展示。
    if (!stopRequested && failedItems.length > 0) {
      patchState(id, {status: "retrySleep", statusText: "准备重试失败项"});
      await waitBulkDownloadPage((message) => patchState(id, {statusText: message}));
      if (!stopRequested) {
        resetBilibiliPlayResolve(failedItems);
        await processQueue(failedItems, true, false);
      }
    }

    return true;
  });

  if (!ran && id === sessionSeq) {
    patchState(id, {status: "stopped", active: false, stopping: false, canResume: !!loader,
      statusText: "已有下载任务进行中，批量任务未启动"});
    return;
  }

  if (id !== sessionSeq) return;

  // 把收集的失败项放回列表：重试仍失败的留在列表里，被停止的也能手动续传。
  addVideos(failedItems);
  const remaining = videoList().length;
  const stopped = stopRequested;
  patchState(id, {
    status: stopped ? "stopped" : "done",
    active: false,
    stopping: false,
    // 被停止或加载下一页中断的会话保留“继续”入口；正常跑完的没有意义。
    canResume: !!loader && (stopped || !!pageLoadError),
    failedCount: remaining,
    statusText: pageLoadError
        ? `加载下一页失败：${pageLoadError}`
        : stopped
            ? "已停止"
            : remaining > 0
                ? `完成，${remaining} 个仍失败`
                : "全部完成",
  });
  if (pageLoadError) {
    notifyBilibiliDownload(`批量下载中断：${pageLoadError}，可回到来源页面重新发起`, "warning");
  } else if (remaining > 0) {
    notifyBilibiliDownload(`批量下载结束：${remaining} 个视频仍失败，已保留在列表中`, "warning");
  } else if (stopped) {
    notifyBilibiliDownload("批量下载已停止", "info");
  } else {
    notifyBilibiliDownload("批量下载全部完成", "success");
  }
  } finally {
    // 恢复下载页的自动解析（放回列表的失败项需要重新解析出状态供手动重试）。
    // 只有最新会话能恢复：若期间已启动新会话，开关由新会话负责。
    if (id === sessionSeq) {
      setBilibiliAutoResolve(true);
      void refreshBilibiliPlayUrls();
    }
  }
}
