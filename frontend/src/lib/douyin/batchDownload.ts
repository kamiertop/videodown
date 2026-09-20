import {createSignal} from "solid-js";
import {downloading, notifyDouyinDownload, runDouyinDownloadTasks, withDouyinDownloadLock} from "./downloadQueue.ts";
import {addDouyinVideos, type DouyinDownloadItem, douyinVideoList, removeDouyinVideo} from "./store.ts";

/**
 * 批量下载的翻页加载器：由来源页面在创建会话时构造。
 * 闭包必须自包含（只捕获纯值、直接调后端绑定），不引用组件 signal，
 * 这样跳转到下载页、来源组件卸载后仍能继续翻页。
 */
export interface DouyinBatchPageLoader {
  loadNext(): Promise<readonly DouyinDownloadItem[]>;
  hasMore(): boolean;
}

export interface DouyinBatchSource {
  title: string;
  /** 接口已知的视频总数（用户作品数、收藏夹总数等）；未知则留空。 */
  totalCount?: number;
  /** 点击时来源页面已加载的全部视频。 */
  initialItems: readonly DouyinDownloadItem[];
  /** 无翻页来源（如播放历史）可以不传。 */
  loader?: DouyinBatchPageLoader;
}

export type DouyinBatchStatus =
    | "downloading"
    | "loadingPage"
    | "retrying"
    | "done"
    | "stopped";

export interface DouyinBatchState {
  title: string;
  totalCount?: number;
  /** 已加载入队的视频数；总数未知时作为分母展示。 */
  loadedCount: number;
  /** 已成功下载的视频数。 */
  downloadedCount: number;
  /** 当前失败（待重试或最终仍失败）的数量。 */
  failedCount: number;
  status: DouyinBatchStatus;
  /** 状态的动态补充文案。 */
  statusText: string;
  /** 已请求停止，正在等待当前批结束。 */
  stopping: boolean;
  /** 会话结束后是否可以继续（被停止或加载中断，且翻页加载器仍可用）。 */
  canResume: boolean;
  active: boolean;
}

const [batchState, setBatchState] = createSignal<DouyinBatchState | null>(null);

export {batchState as douyinBatchState};

let stopRequested = false;
// 会话序号：只允许最新会话写状态，防止旧会话的收尾代码覆盖新会话。
let sessionSeq = 0;
// 当前会话的翻页加载器：停止/中断后保留，供“继续批量”从断点恢复。
let sessionLoader: DouyinBatchPageLoader | undefined;

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

function patchState(
    id: number,
    patch: Partial<DouyinBatchState> | ((prev: DouyinBatchState) => Partial<DouyinBatchState>),
): void {
  if (id !== sessionSeq) return;
  setBatchState((prev) => {
    if (!prev) return prev;
    const resolved = typeof patch === "function" ? patch(prev) : patch;
    return {...prev, ...resolved};
  });
}

/**
 * 启动批量下载会话：首批立即入队，控制器在后台逐批“下载 → 翻页 → 下载”，
 * 全部页完成后对失败项统一重试。返回 false 表示当前有下载任务占用（手动或批量）。
 */
export function startDouyinBatch(source: DouyinBatchSource): boolean {
  if (batchState()?.active || downloading()) return false;

  const initialItems = source.initialItems.filter((item) => item.awemeId.trim().length > 0);
  if (initialItems.length === 0 && !source.loader) return false;

  stopRequested = false;
  const id = ++sessionSeq;
  sessionLoader = source.loader;
  addDouyinVideos(initialItems);
  setBatchState({
    title: source.title,
    totalCount: source.totalCount,
    loadedCount: initialItems.length,
    downloadedCount: 0,
    failedCount: 0,
    status: "downloading",
    statusText: "正在下载",
    stopping: false,
    canResume: false,
    active: true,
  });
  void runSession(id, source.loader);
  return true;
}

/** 请求停止批量会话：当前批下载结束后停止，不再翻页和重试。 */
export function stopDouyinBatch(): void {
  const state = batchState();
  if (!state?.active || state.stopping) return;
  stopRequested = true;
  // 立刻反馈“正在停止”：停止是协作式的，要等当前批（含任务后休眠）跑完才生效。
  setBatchState((prev) => prev ? {...prev, stopping: true, statusText: "正在停止，等待当前批结束…"} : prev);
}

/**
 * 继续已停止/中断的批量会话：复用保存的翻页加载器从断点继续（加载器内部
 * 保存着游标），列表中残留的失败项会作为首批重新下载。统计数字接着累计。
 */
export function resumeDouyinBatch(): boolean {
  const state = batchState();
  if (!state || state.active || !state.canResume || !sessionLoader) return false;
  if (downloading()) {
    notifyDouyinDownload("已有下载任务进行中，请稍后再试", "warning");
    return false;
  }

  stopRequested = false;
  const id = ++sessionSeq;
  setBatchState((prev) => prev ? {
    ...prev,
    status: "downloading",
    statusText: "正在下载当前批",
    stopping: false,
    canResume: false,
    active: true,
  } : prev);
  void runSession(id, sessionLoader);
  return true;
}

/** 会话结束后清除状态卡片。 */
export function dismissDouyinBatch(): void {
  if (!batchState()?.active) {
    sessionLoader = undefined;
    setBatchState(null);
  }
}

async function runSession(id: number, loader: DouyinBatchPageLoader | undefined): Promise<void> {
  // 批内失败项先收进这里并移出下载列表，保证列表里只有当前批的内容；
  // 全部页下载完后再统一重试，仍失败的存在列表里交给用户手动处理。
  const failedItems: DouyinDownloadItem[] = [];
  const failedIDs = new Set<string>();
  let pageLoadError = "";

  // withDouyinDownloadLock 在锁被占用时返回 undefined；startDouyinBatch 已挡掉绝大多数
  // 场景，这里兜底处理启动瞬间的手动下载竞态。
  const ran = await withDouyinDownloadLock(async () => {
    while (true) {
      patchState(id, {status: "downloading", statusText: "正在下载当前批"});

      // 每批取下载列表全量：会话期间用户手动解析加入的内容也会顺带下载。
      const run = await runDouyinDownloadTasks(douyinVideoList());
      patchState(id, (prev) => ({downloadedCount: prev.downloadedCount + run.success}));

      // 批后清场：下载失败和没有可用地址的项都收进失败列表并移出队列，
      // 下一批从干净列表开始，几千个视频也不会在下载页堆积卡片。
      for (const leftover of douyinVideoList()) {
        const key = leftover.awemeId.trim();
        if (!key || failedIDs.has(key)) {
          removeDouyinVideo(key);
          continue;
        }
        failedIDs.add(key);
        failedItems.push(leftover);
        removeDouyinVideo(key);
      }
      patchState(id, {failedCount: failedItems.length});

      if (stopRequested) break;
      if (!loader || !loader.hasMore()) break;

      patchState(id, {status: "loadingPage", statusText: "正在加载下一页"});
      let nextPage: readonly DouyinDownloadItem[];
      try {
        nextPage = await loader.loadNext();
      } catch (error) {
        pageLoadError = errorMessage(error);
        break;
      }
      // 接口还有下一页却返回空列表时按结束处理，防止死循环。
      if (nextPage.length === 0) break;

      addDouyinVideos(nextPage);
      patchState(id, (prev) => ({loadedCount: prev.loadedCount + nextPage.length}));
    }

    // 最终重试：失败任务不触发任务后休眠。
    if (!stopRequested && failedItems.length > 0) {
      patchState(id, {status: "retrying", statusText: `正在重试 ${failedItems.length} 个失败项`});
      addDouyinVideos(failedItems);
      const retry = await runDouyinDownloadTasks(douyinVideoList());
      patchState(id, (prev) => ({downloadedCount: prev.downloadedCount + retry.success}));
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
  addDouyinVideos(failedItems);
  const remaining = douyinVideoList().length;
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
    notifyDouyinDownload(`批量下载中断：${pageLoadError}，可回到来源页面重新发起`, "warning");
  } else if (remaining > 0) {
    notifyDouyinDownload(`批量下载结束：${remaining} 个内容仍失败，已保留在列表中`, "warning");
  } else if (stopped) {
    notifyDouyinDownload("批量下载已停止", "info");
  } else {
    notifyDouyinDownload("批量下载全部完成", "success");
  }
}
