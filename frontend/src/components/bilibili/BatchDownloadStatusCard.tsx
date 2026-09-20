import {type JSXElement, Show} from "solid-js";
import {
  dismissBilibiliBatch,
  bilibiliBatchState,
  resumeBilibiliBatch,
  stopBilibiliBatch,
  type BilibiliBatchState,
} from "../../lib/bilibili/batchDownload.ts";

function statusBadge(state: BilibiliBatchState): {text: string; badgeClass: string; spinning: boolean} {
  // 停止是协作式的：置位后要等当前批（含任务后休眠）跑完，期间显示“停止中”。
  if (state.stopping) return {text: "停止中", badgeClass: "badge-warning", spinning: true};
  switch (state.status) {
    case "resolving":
      return {text: "解析地址", badgeClass: "badge-info", spinning: true};
    case "downloading":
      return {text: "下载中", badgeClass: "badge-primary", spinning: true};
    case "loadingPage":
      return {text: "加载下一页", badgeClass: "badge-info", spinning: true};
    case "retrySleep":
      return {text: "重试前休眠", badgeClass: "badge-warning", spinning: true};
    case "retrying":
      return {text: "重试失败项", badgeClass: "badge-warning", spinning: true};
    case "done":
      return state.failedCount > 0
          ? {text: "部分失败", badgeClass: "badge-warning", spinning: false}
          : {text: "已完成", badgeClass: "badge-success", spinning: false};
    case "stopped":
      return {text: "已停止", badgeClass: "badge-neutral", spinning: false};
  }
}

/**
 * 下载页顶部的批量任务卡片：展示“已下载 / 总数”进度和会话状态。
 * B 站每批先解析播放地址再下载；已完成的条目会从下载列表移除，
 * 几千个视频的收藏夹也只保留当前批的内容。
 */
export default function BilibiliBatchStatusCard(): JSXElement {
  const state = bilibiliBatchState;

  return (
      <Show when={state()}>
        {(batch) => {
          const badge = () => statusBadge(batch());
          // 接口没给总数时用已加载数作分母，翻页过程中分母会持续增长。
          const total = () => batch().totalCount ?? batch().loadedCount;
          const percent = () => total() > 0 ? Math.min(100, Math.round((batch().downloadedCount / total()) * 100)) : 0;

          return (
              <section class="mt-2 flex flex-col gap-2 rounded-lg border border-base-300 bg-base-100 p-3 shadow-sm">
                <div class="flex min-w-0 flex-wrap items-center gap-2">
                  <Show when={badge().spinning}>
                    <span class="loading loading-spinner loading-xs text-primary"></span>
                  </Show>
                  <span class={`badge badge-sm ${badge().badgeClass}`}>{badge().text}</span>
                  <h3 class="min-w-0 truncate text-sm font-bold text-base-content" title={batch().title}>
                    {batch().title}
                  </h3>
                  <span class="text-xs tabular-nums text-base-content/60">{percent()}%</span>
                  <div class="ml-auto flex shrink-0 items-center gap-2">
                    <Show when={batch().failedCount > 0}>
                      <span class="badge badge-outline badge-error badge-sm">失败 {batch().failedCount}</span>
                    </Show>
                    <Show
                        when={batch().active}
                        fallback={
                          <div class="flex items-center gap-2">
                            <Show when={batch().canResume}>
                              <button
                                  class="btn btn-primary btn-xs"
                                  type="button"
                                  onClick={() => resumeBilibiliBatch()}
                              >
                                继续批量
                              </button>
                            </Show>
                            <button class="btn btn-ghost btn-xs" type="button" onClick={dismissBilibiliBatch}>关闭</button>
                          </div>
                        }
                    >
                      <div class="group relative shrink-0">
                        <button
                            class="btn btn-warning btn-xs"
                            type="button"
                            onClick={stopBilibiliBatch}
                            disabled={batch().stopping}
                        >
                          {batch().stopping ? "正在停止..." : "停止批量"}
                        </button>
                        <span
                            role="tooltip"
                            class="pointer-events-none absolute right-0 top-full z-50 mt-1.5 w-max max-w-[min(240px,calc(100vw-24px))] whitespace-normal rounded-md bg-neutral px-2.5 py-1.5 text-center text-[11px] leading-4 text-neutral-content opacity-0 shadow-lg transition-opacity group-hover:opacity-100 group-focus-within:opacity-100"
                        >
                          当前批次下载完成（含任务间休眠）后才会停止，期间请耐心等待
                        </span>
                      </div>
                    </Show>
                  </div>
                </div>

                <div class="flex items-center gap-3">
                  <progress
                      class="progress progress-primary h-2 min-w-0 flex-1"
                      value={batch().downloadedCount}
                      max={Math.max(1, total())}
                  />
                  <span class="shrink-0 text-xs tabular-nums text-base-content/70">
                    已下载 {batch().downloadedCount} / {total()}
                  </span>
                </div>

                <p class="text-xs text-base-content/55">
                  {batch().statusText}
                  <Show when={batch().totalCount === undefined}>
                    {` · 已加载 ${batch().loadedCount} 个`}
                  </Show>
                </p>
              </section>
          );
        }}
      </Show>
  );
}
