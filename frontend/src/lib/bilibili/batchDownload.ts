import {DownloadedVideoKeys} from "@bindings/github.com/kamiertop/videodown/bilibili/download/service.ts";
import {GetConcurrencyNum, GetSkipDownloaded} from "@bindings/github.com/kamiertop/videodown/utils/settings";
import {createSignal} from "solid-js";
import type {MediaCardItem} from "../model.ts";
import {bilibiliPlayResolveKey} from "./playResolve.ts";
import {
  awaitBilibiliDownloadLock,
  notifyBilibiliDownload,
  refreshBilibiliPlayUrls,
  resetBilibiliPlayResolve,
  runBilibiliDownloadTasks,
  setBilibiliAutoResolve,
} from "./downloadQueue.ts";
import {addVideos, removeVideos, videoList} from "./store.ts";

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

export type BilibiliBatchStatus
    = "queued"
    | "resolving"
    | "downloading"
    | "loadingPage"
    | "retrying"
    | "done"
    | "stopped";

export interface BilibiliBatchState {
  title: string;
  totalCount?: number;
  /** 已加载入队的视频数；总数未知时作为分母展示。 */
  loadedCount: number;
  /** 已成功下载的视频数（只统计本会话的内容）。 */
  downloadedCount: number;
  /** 增量下载跳过的数量（历史已下载且文件仍在），同样只统计本会话的内容。 */
  skippedCount: number;
  /** 当前失败（待重试或最终仍失败）的数量，同样只统计本会话的内容。 */
  failedCount: number;
  status: BilibiliBatchStatus;
  /** 状态的动态补充文案。 */
  statusText: string;
  /** 已请求停止，正在等待当前批结束。 */
  stopping: boolean;
  /** 会话结束后是否可以继续（被停止或加载中断，且翻页加载器仍可用）。 */
  canResume: boolean;
  active: boolean;
}

/** 一次“一键下载全部”的会话记录：state 供状态卡展示，其余字段供队列调度和续传使用。 */
export interface BilibiliBatchSession {
  id: number;
  state: BilibiliBatchState;
  /** 来源页面的翻页加载器（内部保存游标）；停止/中断后保留，供“继续批量”从断点恢复。 */
  loader?: BilibiliBatchPageLoader;
  /** 会话启动时入队的第一批内容；继续批量时换成列表里仍残留的失败项。 */
  initialItems: readonly MediaCardItem[];
  /** 会话收集的失败项（下载失败 + 解析失败），供“继续批量”筛选出仍需下载的内容。 */
  failedItems: MediaCardItem[];
  /** 已请求停止；排队中的会话被停止等于取消排队。 */
  stopRequested: boolean;
  /** 是否真正执行过：取消排队而从未运行的会话，继续时仍用原始首批。 */
  started: boolean;
}

const [sessions, setSessions] = createSignal<BilibiliBatchSession[]>([]);

export {sessions as bilibiliBatchSessions};

/** 是否有批量会话正在排队或执行；下载页用它决定是否隐藏手动下载摘要行。 */
export function bilibiliBatchActive(): boolean {
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
function queuedState(): Pick<BilibiliBatchState, "status" | "statusText" | "stopping" | "canResume" | "active"> {
  const queuedBehind = sessions().some((session) => session.state.active);
  return {
    status: "queued",
    statusText: queuedBehind ? "排队等待前一个批量结束" : "正在启动",
    stopping: false,
    canResume: false,
    active: true,
  };
}

// 下载列表当前全部解析键的集合：清场、续传筛选和收尾统计共用。
function keysInList(): Set<string | null> {
  return new Set(videoList().map((item) => bilibiliPlayResolveKey(item)));
}

function findSession(id: number): BilibiliBatchSession | undefined {
  return sessions().find((session) => session.id === id);
}

function patchState(
    id: number,
    patch: Partial<BilibiliBatchState> | ((prev: BilibiliBatchState) => Partial<BilibiliBatchState>),
): void {
  setSessions((prev) => prev.map((session) => {
    if (session.id !== id) return session;
    const resolved = typeof patch === "function" ? patch(session.state) : patch;
    return {...session, state: {...session.state, ...resolved}};
  }));
}

/**
 * 启动批量下载会话：可以同时创建多个，后排队的等前一个结束后自动开始。
 * 会话控制器在后台逐块“解析地址 → 下载”推进并自动翻页，全部页完成后
 * 对失败项（下载失败 + 解析失败）统一重试。整组（页）入队展示——下载页
 * 列表显示当前批全部内容，尚未轮到的只是暂时没有播放地址；解析与下载仍
 * 按下载并发数分块推进，解析请求只比下载领先一小块，不会一次轰炸整页。
 * 返回 false 仅表示来源没有可下载内容，或同一来源已在排队/执行中。
 */
export function startBilibiliBatch(source: BilibiliBatchSource): boolean {
  const initialItems = source.initialItems.filter((item) => item.bvid?.trim().length > 0);
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
export function stopBilibiliBatch(id: number): void {
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
 * 继续已停止/中断的批量会话：重新排队，复用保存的翻页加载器从断点继续
 * （加载器内部保存着游标）。首批换成失败项中仍留在列表里的——已被其他
 * 会话下载或被用户移除的不再重下；从未执行过（取消排队）的会话仍用原始首批。
 */
export function resumeBilibiliBatch(id: number): boolean {
  const session = findSession(id);
  if (!session || session.state.active || !session.state.canResume) return false;

  const initialItems = session.started ? leftoversInList(session) : session.initialItems;
  // 清掉残留失败项的旧解析记录，否则解析失败的旧 entry 会让重试直接跳过这些视频。
  resetBilibiliPlayResolve(initialItems);
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
export function dismissBilibiliBatch(id: number): void {
  setSessions((prev) => {
    const session = prev.find((s) => s.id === id);
    if (!session || session.state.active) return prev;
    return prev.filter((s) => s.id !== id);
  });
}

/** 会话收集的失败项中仍留在下载列表里的部分：只有这些才需要继续下载。 */
function leftoversInList(session: BilibiliBatchSession): readonly MediaCardItem[] {
  const inList = keysInList();
  return session.failedItems.filter((item) => {
    const key = bilibiliPlayResolveKey(item);
    return !!key && inList.has(key);
  });
}

/**
 * 队列调度：逐个取出排队中的会话串行执行，全部结束后退出。
 * 队列运行期间关闭下载页的“到页即全量解析”，解析改由会话按块驱动；
 * 全部会话结束后恢复自动解析，放回列表的失败项重新解析出状态供手动重试。
 */
async function drainSessions(): Promise<void> {
  if (draining) return;
  draining = true;
  setBilibiliAutoResolve(false);
  try {
    while (true) {
      const next = sessions().find((session) => session.state.status === "queued");
      if (!next) break;
      await runSession(next.id, next.loader);
    }
  } finally {
    draining = false;
    setBilibiliAutoResolve(true);
    void refreshBilibiliPlayUrls();
  }
}

// 执行单个会话：先等全局下载锁空闲（轮到它时可能正好有手动下载在跑），再运行主循环。
async function runSession(id: number, loader: BilibiliBatchPageLoader | undefined): Promise<void> {
  await awaitBilibiliDownloadLock(
      () => runSessionLoop(id, loader),
      () => findSession(id)?.stopRequested === true,
      () => patchState(id, {statusText: "有手动下载进行中，等待开始…"}),
  );
}

/**
 * 会话主循环：逐组“解析 → 下载 → 批后清场”推进并自动翻页，全部页完成后
 * 对失败项统一重试。会话只处理和统计自己的内容；用户手动加入的内容不受
 * 批后清场影响，也不计入本会话的失败数，批量结束时仍留在列表里。
 */
async function runSessionLoop(id: number, loader: BilibiliBatchPageLoader | undefined): Promise<boolean> {
  const session = findSession(id);
  if (!session) return true;
  // 标记已执行并释放首批引用（继续批量只会用 failedItems，整批内容不必再持有）。
  setSessions((prev) => prev.map((s) => s.id === id ? {...s, started: true, initialItems: []} : s));

  const stopRequested = () => findSession(id)?.stopRequested ?? true;
  // 批内失败项（下载失败 + 解析失败，如充电专属）先收进这里并移出下载列表，
  // 保证列表里只有当前块的内容。
  const failedItems: MediaCardItem[] = [];
  const failedKeys = new Set<string>();
  let pageLoadError = "";

  // 下载并发数即分块大小：解析只比下载领先一小块（下载远慢于解析），
  // 避免点击后一次性发出整页的解析请求。读取失败时退回单块单下。
  const chunkSize = Math.max(1, Number(await GetConcurrencyNum().catch(() => 1)) || 1);

  // 一组（页）视频整组入队展示：下载页列表显示当前批全部内容，
  // 尚未轮到的项只是暂时没有播放地址；解析与下载仍只按小块推进。
  const processQueue = async (items: readonly MediaCardItem[], retrying: boolean, sweepAfter: boolean): Promise<void> => {
    if (items.length === 0) return;
    addVideos(items);

    for (let start = 0; start < items.length; start += chunkSize) {
      if (stopRequested()) return;
      const chunk = items.slice(start, start + chunkSize);

      // 下载时再解析：只解析当前小块（给 refreshBilibiliPlayUrls 传子集）。
      patchState(id, {status: "resolving", statusText: retrying ? "正在重新解析失败项" : "正在解析视频地址"});
      await refreshBilibiliPlayUrls(chunk);
      if (stopRequested()) return;

      patchState(id, {status: retrying ? "retrying" : "downloading", statusText: retrying ? "正在重试失败项" : "正在下载当前批"});
      const run = await runBilibiliDownloadTasks(chunk);
      patchState(id, (prev) => ({downloadedCount: prev.downloadedCount + run.success}));
    }

    if (!sweepAfter) return;
    // 组后清场：本组里仍留在列表的就是失败项（下载失败 + 解析失败 + 无可用地址），
    // 收进失败列表并移出队列，下一组从干净列表开始，几千个视频也不会堆积卡片。
    // 只收本组的项：手动加入或其他会话残留的内容不属于本会话，保持原样。
    const inList = keysInList();
    const sweptIDs = new Set<number>();
    for (const item of items) {
      const key = bilibiliPlayResolveKey(item);
      if (!key || failedKeys.has(key) || !inList.has(key)) continue;
      failedKeys.add(key);
      failedItems.push(item);
      sweptIDs.add(item.id);
    }
    removeVideos(sweptIDs);
    patchState(id, {failedCount: failedItems.length});
  };

  // 增量下载：开启时每组解析播放地址前剔除“已下载过且文件仍在”的内容，
  // 整页已下载的视频不再浪费解析请求；设置读取失败按默认开启处理。
  // 不做“整组已下载就停止翻页”的早停推断：列表顺序不保证与下载顺序一致，翻完全部页确保无遗漏。
  const incremental = await GetSkipDownloaded().catch(() => true);
  // 查询失败只提示一次（例如后端未随新版本重启导致绑定缺失），不能无声回退成“先解析再判断”。
  let incrementalQueryNotified = false;

  // 返回仍需解析下载的子集；已下载项从展示列表移除（多数情况尚未入列，移除是 no-op）并计入跳过数。
  const filterIncremental = async (items: readonly MediaCardItem[]): Promise<readonly MediaCardItem[]> => {
    if (!incremental || items.length === 0) return items;
    try {
      const keys = items.map((item) => ({bvid: item.bvid ?? "", cid: item.cid ?? 0}));
      const downloaded = (await DownloadedVideoKeys(keys)) ?? [];
      if (downloaded.length === 0) return items;
      const skip = new Set(downloaded);
      removeVideos(new Set(downloaded.map((index) => items[index].id)));
      patchState(id, (prev) => ({skippedCount: prev.skippedCount + skip.size}));
      return items.filter((_, index) => !skip.has(index));
    } catch (error) {
      if (!incrementalQueryNotified) {
        incrementalQueryNotified = true;
        notifyBilibiliDownload(`查询已下载记录失败，本次批量不做增量过滤：${errorMessage(error)}`, "warning");
      }
      return items;
    }
  };

  let queueItems: readonly MediaCardItem[] = session.initialItems;
  patchState(id, {status: "resolving", statusText: "正在解析视频地址"});
  while (true) {
    await processQueue(await filterIncremental(queueItems), false, true);
    if (stopRequested()) break;
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

    queueItems = nextPage;
    patchState(id, (prev) => ({loadedCount: prev.loadedCount + nextPage.length}));
  }

  // 最终重试：清掉失败项的解析记录重新解析（解析失败可能只是临时网络问题）；
  // 失败项同样按块推进。重试不清场：仍失败的留在列表里，会话收尾本来就要放回展示。
  if (!stopRequested() && failedItems.length > 0) {
    resetBilibiliPlayResolve(failedItems);
    await processQueue(failedItems, true, false);
  }

  // 把收集的失败项放回列表：重试仍失败的留在列表里，被停止的也能手动续传。
  addVideos(failedItems);
  // 收尾统计只看本会话的失败项：手动加入或其他会话残留的内容不算本会话失败。
  const inList = keysInList();
  const remaining = failedItems.filter((item) => {
    const key = bilibiliPlayResolveKey(item);
    return !!key && inList.has(key);
  }).length;
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
    notifyBilibiliDownload(`批量下载中断：${pageLoadError}，可回到来源页面重新发起`, "warning");
  } else if (remaining > 0) {
    notifyBilibiliDownload(`批量下载结束：${remaining} 个视频仍失败，已保留在列表中`, "warning");
  } else if (stopped) {
    notifyBilibiliDownload("批量下载已停止", "info");
  } else if (incrementalToastText) {
    notifyBilibiliDownload(incrementalToastText, "success");
  } else {
    notifyBilibiliDownload("批量下载全部完成", "success");
  }
  if (!stopped && !pageLoadError && remaining === 0) {
    setTimeout(() => dismissBilibiliBatch(id), SUCCESS_DISMISS_DELAY);
  }
  return true;
}
