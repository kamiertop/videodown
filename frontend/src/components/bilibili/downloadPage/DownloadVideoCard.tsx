import type {JSXElement} from "solid-js";
import {For, Show} from "solid-js";
import IconChat from "../../icons/IconChat.tsx";
import IconEye from "../../icons/IconEye.tsx";
import type {DownloadProgress, PlayResolveEntry} from "../../../lib/bilibiliDownloadQueue.ts";
import {formatListSource} from "../../../lib/bilibiliDownloadQueue.ts";
import {formatCount, formatDate, formatDuration} from "../../../lib/format.ts";
import type {MediaCardItem} from "../../../lib/model.ts";
import QualitySelectors from "./QualitySelectors.tsx";

interface DownloadVideoCardProps {
    canDownload: boolean;
    downloading: boolean;
    item: MediaCardItem;
    entry: PlayResolveEntry | undefined;
    progress: DownloadProgress | undefined;
    onDownload: () => void;
    onPickAudio: (audioId: number) => void;
    onPickQn: (qn: number) => void;
    onRemove: () => void;
}

// 下载列表里的单张视频卡片。
// 它只负责展示和把用户操作抛给父层；解析状态、画质切换、下载进度都由 useBilibiliDownloadQueue 管。
export default function DownloadVideoCard(props: DownloadVideoCardProps): JSXElement {
    const listSrc = () => formatListSource(props.item);
    const accessLabels = () => {
        // 充电专属、试看、付费等标签来自视频详情；即使解析失败也尽量展示原因。
        const entry = props.entry;
        if (entry?.status === "done") return entry.data.accessInfo.labels;
        if (entry?.status === "error") return entry.accessInfo?.labels ?? [];
        return [];
    };
    const progressText = () => {
        // 后端把视频、音频、合并阶段映射成同一条 0-100 的进度。
        const progress = props.progress;
        if (!progress) return "";
        if (progress.phase === "video") return `视频下载 ${Math.round(progress.percent)}%`;
        if (progress.phase === "audio") return `音频下载 ${Math.round(progress.percent)}%`;
        if (progress.phase === "merge") return "合并中";
        if (progress.phase === "done") return "完成";
        return "下载失败";
    };

    return (
        <article class="flex flex-row gap-3">
            <div class="relative w-40 aspect-video object-cover">
                <img
                    class="h-full w-full object-cover"
                    src={props.item.cover}
                    alt={props.item.title}
                    referrerPolicy="no-referrer"
                />
                <div class="pointer-events-none absolute inset-x-0 bottom-0 flex items-end justify-between bg-linear-to-t from-black/75 to-transparent px-2 pb-1.5 pt-6">
                    <div class="flex items-center gap-2 text-white/90">
                        <span class="flex items-center gap-0.5 text-[11px]">
                            <IconEye class="h-3 w-3"/>{formatCount(props.item.play)}
                        </span>
                        <span class="flex items-center gap-0.5 text-[11px]">
                            <IconChat class="h-3 w-3"/> {props.item.danmaku}
                        </span>
                    </div>
                    <span class="rounded bg-black/65 px-1 py-0.5 text-[11px] tabular-nums text-white/95">
                        {formatDuration(props.item.duration)}
                    </span>
                </div>
            </div>
            <div class="flex min-w-0 flex-1 flex-col gap-2">
                <h3 class="text-base font-semibold leading-snug text-base-content line-clamp-2">
                    {props.item.title}
                </h3>
                <Show when={accessLabels().length > 0}>
                    <div class="flex flex-wrap gap-1">
                        <For each={accessLabels()}>
                            {(label) => (
                                <span class="badge badge-warning badge-sm">
                                    {label}
                                </span>
                            )}
                        </For>
                    </div>
                </Show>
                <div class="flex flex-row gap-3 justify-between">
                    <div class="mt-4 w-64">
                        <Show when={listSrc()}>
                            <p class="text-sm font-medium text-primary">{listSrc()}</p>
                        </Show>
                        <div class="flex mt-6 items-center gap-x-2 gap-y-0.5 text-sm text-base-content/75">
                            <span class="inline-flex items-center gap-1">
                                <svg
                                    xmlns="http://www.w3.org/2000/svg"
                                    class="h-3 w-3 shrink-0"
                                    viewBox="0 0 20 20"
                                    fill="currentColor"
                                >
                                    <path
                                        fill-rule="evenodd"
                                        d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"
                                        clip-rule="evenodd"
                                    />
                                </svg>
                                {props.item.upperName}
                            </span>
                            <span class="text-base-content/35">|</span>
                            <span>{formatDate(props.item.pubtime)}</span>
                        </div>
                    </div>
                    <QualitySelectors
                        entry={props.entry}
                        onPickAudio={props.onPickAudio}
                        onPickQn={props.onPickQn}
                    />
                    <div class="flex flex-col gap-2">
                        <button class="btn btn-warning btn-sm" type="button" onClick={props.onRemove}>
                            移除
                        </button>
                        <button
                            class="btn btn-info btn-sm"
                            type="button"
                            onClick={props.onDownload}
                            disabled={!props.canDownload || props.downloading}
                        >
                            {props.downloading ? "下载中..." : "下载"}
                        </button>
                    </div>
                </div>
                <Show when={props.downloading && props.progress}>
                    {(progress) => (
                        <div class="flex items-center gap-2">
                            <progress
                                class="progress progress-info h-2 flex-1"
                                value={progress().percent}
                                max="100"
                            />
                            <span class="w-24 text-right text-xs text-base-content/70">
                                {progressText()}
                            </span>
                        </div>
                    )}
                </Show>
            </div>
        </article>
    );
}
