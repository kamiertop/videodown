import {Show, type JSXElement} from "solid-js";

export interface ResolveProgress {
  total: number;
  resolved: number;
  errors: number;
  loading: number;
}

interface DownloadSummaryBarProps {
  count: number;
  downloading: boolean;
  onDownload: () => void;
  resolveProgress?: ResolveProgress;
  /** 本页累计已下载成功的视频数（下载完成后仍保留展示）。 */
  completedCount?: number;
}

// 下载列表顶部的汇总栏：显示待下载数量、解析进度、下载按钮。
export default function DownloadSummaryBar(props: DownloadSummaryBarProps): JSXElement {
  const done = () => {
    const p = props.resolveProgress;
    if (!p) return 0;
    return p.resolved + p.errors;
  };

  const isResolving = () => {
    const p = props.resolveProgress;
    return p && p.loading > 0;
  };

  return (
    <Show when={props.count > 0 || (props.completedCount ?? 0) > 0}>
      <section class="mt-2 flex flex-row items-center justify-between rounded-lg border border-base-200 bg-base-100 p-3 shadow-sm">
        <div class="flex min-w-0 flex-1 items-center gap-3">
          <Show when={props.count > 0}>
            <div class="flex items-center gap-1.5">
              <div class="badge badge-primary">{props.count}</div>
              <span class="text-xs">个视频待下载</span>
            </div>
          </Show>
          <Show when={(props.completedCount ?? 0) > 0}>
            <div class="flex items-center gap-1.5">
              <div class="badge badge-success">{props.completedCount}</div>
              <span class="text-xs text-success">个视频已完成</span>
            </div>
          </Show>
          <Show when={isResolving()}>
            <div class="flex items-center gap-1.5 text-xs text-info">
              <span class="loading loading-spinner loading-xs"></span>
              <span>
                正在解析 {done()}/{props.resolveProgress!.total}
              </span>
            </div>
          </Show>
        </div>
        <button
          class="btn btn-success btn-xs gap-1.5"
          type="button"
          onClick={props.onDownload}
          disabled={props.downloading}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-3.5 w-3.5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
            />
          </svg>
          {props.downloading ? "下载中..." : "开始下载"}
        </button>
      </section>
    </Show>
  );
}
