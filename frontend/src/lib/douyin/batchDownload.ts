import {DownloadedAwemeIDs} from "@bindings/github.com/kamiertop/videodown/douyin/download/service.ts";
import {GetSkipDownloaded} from "@bindings/github.com/kamiertop/videodown/utils/settings";
import {createSignal} from "solid-js";
import {awaitDouyinDownloadLock, notifyDouyinDownload, runDouyinDownloadTasks} from "./downloadQueue.ts";
import {addDouyinVideos, type DouyinDownloadItem, douyinVideoList, removeDouyinVideos} from "./store.ts";

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

export type DouyinBatchStatus
    = "queued"
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
  /** 已成功下载的视频数（只统计本会话的内容，顺带下载的手动项除外）。 */
  downloadedCount: number;
  /** 增量下载跳过的数量（历史已下载且文件仍在），同样只统计本会话的内容。 */
  skippedCount: number;
  /** 当前失败（待重试或最终仍失败）的数量，同样只统计本会话的内容。 */
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

/** 一次“一键下载全部”的会话记录：state 供状态卡展示，其余字段供队列调度和续传使用。 */
export interface DouyinBatchSession {
  id: number;
  state: DouyinBatchState;
  /** 来源页面的翻页加载器（内部保存游标）；停止/中断后保留，供“继续批量”从断点恢复。 */
  loader?: DouyinBatchPageLoader;
  /** 会话启动时入队的第一批内容；继续批量时换成列表里仍残留的失败项。 */
  initialItems: readonly DouyinDownloadItem[];
  /** 会话收集的失败项，供“继续批量”筛选出仍需下载的内容。 */
  failedItems: DouyinDownloadItem[];
  /** 已请求停止；排队中的会话被停止等于取消排队。 */
  stopRequested: boolean;
  /** 是否真正执行过：取消排队而从未运行的会话，继续时仍用原始首批。 */
  started: boolean;
}

const [sessions, setSessions] = createSignal<DouyinBatchSession[]>([]);

export {sessions as douyinBatchSessions};

/** 是否有批量会话正在排队或执行；下载页用它决定是否隐藏手动下载摘要行。 */
export function douyinBatchActive(): boolean {
  return sessions().some((session) => session.state.active);
}

let nextSessionID = 0;
// 队列调度标志：同一时刻只有一个 drainSessions 在跑，会话逐个串行执行，
// 下载并发与休眠始终由用户设置控制，多个批量不会叠加请求频率。
let draining = false;

// 全部成功的会话卡片再展示一小会儿就自动关闭；有失败或被停止的保留，交给用户处理。
const SUCCESS_DISMISS_DELAY = 3000;

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

// start/resume 共用的“重新排队”状态片段：是否已有人在前面决定提示文案。
function queuedState(): Pick<DouyinBatchState, "status" | "statusText" | "stopping" | "canResume" | "active"> {
  const queuedBehind = sessions().some((session) => session.state.active);
  return {
    status: "queued",
    statusText: queuedBehind ? "排队等待前一个批量结束" : "正在启动",
    stopping: false,
    canResume: false,
    active: true,
  };
}

function findSession(id: number): DouyinBatchSession | undefined {
  return sessions().find((session) => session.id === id);
}

function patchState(
    id: number,
    patch: Partial<DouyinBatchState> | ((prev: DouyinBatchState) => Partial<DouyinBatchState>),
): void {
  setSessions((prev) => prev.map((session) => {
    if (session.id !== id) return session;
    const resolved = typeof patch === "function" ? patch(session.state) : patch;
    return {...session, state: {...session.state, ...resolved}};
  }));
}

/**
 * 启动批量下载会话：可以同时创建多个，后排队的等前一个结束后自动开始。
 * 返回 false 仅表示来源没有可下载内容，或同一来源已在排队/执行中。
 */
export function startDouyinBatch(source: DouyinBatchSource): boolean {
  const initialItems = source.initialItems.filter((item) => item.awemeId.trim().length > 0);
  if (initialItems.length === 0 && !source.loader) return false;
  // 同一来源重复点击直接拒绝，否则排队两个相同会话会把同一批内容完整下载两遍。
  const duplicated = sessions().some((session) => session.state.active && session.state.title === source.title);
  if (duplicated) return false;

  setSessions((prev) => [...prev, {
    id: ++nextSessionID,
    state: {
      title: source.title,
      totalCount: source.totalCount,
      loadedCount: initialItems.length,
      downloadedCount: 0,
      skippedCount: 0,
      failedCount: 0,
      ...queuedState(),
    },
    loader: source.loader,
    initialItems,
    failedItems: [],
    stopRequested: false,
    started: false,
  }]);
  void drainSessions();
  return true;
}

/** 请求停止批量会话：执行中的等当前批结束后停止，排队中的直接取消排队。 */
export function stopDouyinBatch(id: number): void {
  const session = findSession(id);
  if (!session || !session.state.active || session.state.stopping) return;

  if (session.state.status === "queued") {
    setSessions((prev) => prev.map((s) => s.id === id ? {
      ...s,
      stopRequested: true,
      state: {...s.state, status: "stopped", active: false, stopping: false, canResume: !!s.loader,
        statusText: "已取消排队"},
    } : s));
    return;
  }
  setSessions((prev) => prev.map((s) => s.id === id ? {
    ...s,
    stopRequested: true,
    // 立刻反馈“正在停止”：停止是协作式的，要等当前批（含任务后休眠）跑完才生效。
    state: {...s.state, stopping: true, statusText: "正在停止，等待当前批结束…"},
  } : s));
}

/**
 * 继续已停止/中断的批量会话：重新排队，复用保存的翻页加载器从断点继续（加载器内部
 * 保存着游标）。首批换成失败项中仍留在列表里的——已被其他会话顺带下载或被用户移除
 * 的不再重下；从未执行过（取消排队）的会话仍用原始首批。
 */
export function resumeDouyinBatch(id: number): boolean {
  const session = findSession(id);
  if (!session || session.state.active || !session.state.canResume) return false;

  const initialItems = session.started ? leftoversInList(session) : session.initialItems;
  setSessions((prev) => prev.map((s) => s.id === id ? {
    ...s,
    initialItems,
    stopRequested: false,
    state: {...s.state, ...queuedState()},
  } : s));
  void drainSessions();
  return true;
}

/** 清除已结束会话的状态卡片。 */
export function dismissDouyinBatch(id: number): void {
  setSessions((prev) => {
    const session = prev.find((s) => s.id === id);
    if (!session || session.state.active) return prev;
    return prev.filter((s) => s.id !== id);
  });
}

/** 会话收集的失败项中仍留在下载列表里的部分：只有这些才需要继续下载。 */
function leftoversInList(session: DouyinBatchSession): readonly DouyinDownloadItem[] {
  const inList = new Set(douyinVideoList().map((item) => item.awemeId.trim()));
  return session.failedItems.filter((item) => inList.has(item.awemeId.trim()));
}

/** 队列调度：逐个取出排队中的会话串行执行，全部结束后退出。 */
async function drainSessions(): Promise<void> {
  if (draining) return;
  draining = true;
  try {
    while (true) {
      const next = sessions().find((session) => session.state.status === "queued");
      if (!next) break;
      await runSession(next.id, next.loader);
    }
  } finally {
    draining = false;
  }
}

// 执行单个会话：先等全局下载锁空闲（轮到它时可能正好有手动下载在跑），再运行主循环。
async function runSession(id: number, loader: DouyinBatchPageLoader | undefined): Promise<void> {
  await awaitDouyinDownloadLock(
      () => runSessionLoop(id, loader),
      () => findSession(id)?.stopRequested === true,
      () => patchState(id, {statusText: "有手动下载进行中，等待开始…"}),
  );
}

/**
 * 会话主循环：下载 → 批后清场 → 翻页，全部页完成后对失败项统一重试。
 * 每批取下载列表全量：会话期间用户手动解析加入的内容也会顺带下载；
 * 但它们不属于本会话——不计入下载/失败统计，也不会被批后清场收走，
 * 批量结束时仍留在列表里等下一轮顺带下载或用户手动处理。
 */
async function runSessionLoop(id: number, loader: DouyinBatchPageLoader | undefined): Promise<boolean> {
  const session = findSession(id);
  if (!session) return true;
  // 标记已执行并释放首批引用（继续批量只会用 failedItems，整批内容不必再持有）。
  setSessions((prev) => prev.map((s) => s.id === id ? {...s, started: true, initialItems: []} : s));

  const stopRequested = () => findSession(id)?.stopRequested ?? true;
  // 本会话负责的内容：首批 + 翻页加载的每一页；手动加入的内容不在此列。
  const ownedKeys = new Set(session.initialItems.map((item) => item.awemeId.trim()).filter(Boolean));
  // 本会话当前批的内容游标：首批是 initialItems，之后是每个新加载的页。
  let batchItems: readonly DouyinDownloadItem[] = session.initialItems;
  // 批内失败项先收进这里并移出下载列表，保证列表里只有当前批的内容。
  const failedItems: DouyinDownloadItem[] = [];
  const failedIDs = new Set<string>();
  let pageLoadError = "";
  // 增量下载：开启时每批下载前剔除本会话中“已下载过且文件仍在”的内容；设置读取失败按默认开启处理。
  const incremental = await GetSkipDownloaded().catch(() => true);
  // 查询失败只提示一次（例如后端未随新版本重启导致绑定缺失），不能无声回退成重复下载。
  let incrementalQueryNotified = false;

  addDouyinVideos(session.initialItems);
  patchState(id, {status: "downloading", statusText: "正在下载当前批"});

  while (true) {
    // 增量过滤只作用于本会话的内容；手动加入的项保留用户的重下意图。
    // 不做“整页已下载就停止翻页”的早停推断：列表顺序不保证与下载顺序一致，
    // 保险起见翻完全部页，只靠过滤省掉已下载内容的下载请求。
    if (incremental && batchItems.length > 0) {
      try {
        const downloaded = (await DownloadedAwemeIDs(batchItems.map((item) => item.awemeId.trim()))) ?? [];
        if (downloaded.length > 0) {
          removeDouyinVideos(new Set(downloaded.map((awemeId) => awemeId.trim())));
          patchState(id, (prev) => ({skippedCount: prev.skippedCount + downloaded.length}));
        }
      } catch (error) {
        if (!incrementalQueryNotified) {
          incrementalQueryNotified = true;
          notifyDouyinDownload(`查询已下载记录失败，本次批量不做增量过滤：${errorMessage(error)}`, "warning");
        }
      }
    }

    patchState(id, {status: "downloading", statusText: "正在下载当前批"});

    const run = await runDouyinDownloadTasks(douyinVideoList());
    const ownedSuccess = run.successIDs.filter((awemeId) => ownedKeys.has(awemeId)).length;
    patchState(id, (prev) => ({downloadedCount: prev.downloadedCount + ownedSuccess}));

    // 批后清场：只收走本会话里下载失败和没有可用地址的项并移出队列，
    // 下一批从干净列表开始，几千个视频也不会在下载页堆积卡片。
    const sweptKeys = new Set<string>();
    for (const leftover of douyinVideoList()) {
      const key = leftover.awemeId.trim();
      if (!ownedKeys.has(key) || sweptKeys.has(key)) continue;
      sweptKeys.add(key);
      if (!failedIDs.has(key)) {
        failedIDs.add(key);
        failedItems.push(leftover);
      }
    }
    removeDouyinVideos(sweptKeys);
    patchState(id, {failedCount: failedItems.length});

    if (stopRequested()) break;
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

    for (const item of nextPage) {
      const key = item.awemeId.trim();
      if (key) ownedKeys.add(key);
    }
    addDouyinVideos(nextPage);
    batchItems = nextPage;
    patchState(id, (prev) => ({loadedCount: prev.loadedCount + nextPage.length}));
  }

  // 最终重试：失败任务不触发任务后休眠。
  if (!stopRequested() && failedItems.length > 0) {
    patchState(id, {status: "retrying", statusText: `正在重试 ${failedItems.length} 个失败项`});
    addDouyinVideos(failedItems);
    const retry = await runDouyinDownloadTasks(douyinVideoList());
    const ownedRetry = retry.successIDs.filter((awemeId) => ownedKeys.has(awemeId)).length;
    patchState(id, (prev) => ({downloadedCount: prev.downloadedCount + ownedRetry}));
  }

  // 把收集的失败项放回列表：重试仍失败的留在列表里，被停止的也能手动续传。
  addDouyinVideos(failedItems);
  const inList = new Set(douyinVideoList().map((item) => item.awemeId.trim()));
  const remaining = failedItems.filter((item) => inList.has(item.awemeId.trim())).length;
  const stopped = stopRequested();
  const finalState = findSession(id)?.state;
  const downloadedTotal = finalState?.downloadedCount ?? 0;
  const skippedTotal = finalState?.skippedCount ?? 0;
  // 增量收尾文案只计算一次：卡片与 toast 共用同一判断，“无新增”文案共用同一段。
  const noNewText = `无新增视频，已跳过 ${skippedTotal} 个`;
  const incrementalCardText = skippedTotal > 0
      ? (downloadedTotal > 0 ? `已增量完成（新下载 ${downloadedTotal}，跳过 ${skippedTotal}）` : noNewText)
      : "";
  const incrementalToastText = skippedTotal > 0
      ? (downloadedTotal > 0 ? `增量下载完成：新下载 ${downloadedTotal} 个，跳过 ${skippedTotal} 个` : noNewText)
      : "";

  setSessions((prev) => prev.map((s) => s.id === id ? {...s, failedItems} : s));
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
                : incrementalCardText || "全部完成",
  });
  if (pageLoadError) {
    notifyDouyinDownload(`批量下载中断：${pageLoadError}，可回到来源页面重新发起`, "warning");
  } else if (remaining > 0) {
    notifyDouyinDownload(`批量下载结束：${remaining} 个内容仍失败，已保留在列表中`, "warning");
  } else if (stopped) {
    notifyDouyinDownload("批量下载已停止", "info");
  } else if (incrementalToastText) {
    notifyDouyinDownload(incrementalToastText, "success");
  } else {
    notifyDouyinDownload("批量下载全部完成", "success");
  }
  if (!stopped && !pageLoadError && remaining === 0) {
    setTimeout(() => dismissDouyinBatch(id), SUCCESS_DISMISS_DELAY);
  }
  return true;
}
