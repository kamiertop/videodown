import {Show, type JSXElement} from "solid-js";

interface DownloadSummaryBarProps {
    count: number;
    sourceSummary: string;
    downloading: boolean;
    onDownload: () => void;
}

// 下载列表顶部的汇总栏：显示待下载数量、来源摘要，并触发“下载全部”。
export default function DownloadSummaryBar(props: DownloadSummaryBarProps): JSXElement {
    return (
        <Show when={props.count > 0}>
            <section class="mt-2 rounded-lg p-3 flex flex-row justify-between items-center shadow-sm">
                <div class="flex min-w-0 flex-1 flex-col gap-1">
                    <div class="flex items-center gap-2">
                        <div class="badge badge-primary">{props.count}</div>
                        <span class="text-xs">个视频待下载</span>
                    </div>
                    <Show when={props.sourceSummary}>
                        <p class="truncate text-sm text-base-content/80">
                            {props.sourceSummary}
                        </p>
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
