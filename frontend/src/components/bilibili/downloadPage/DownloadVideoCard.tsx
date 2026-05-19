import type {JSXElement} from "solid-js";
import {For, Show} from "solid-js";
import type {DownloadProgress, PlayResolveEntry} from "../../../lib/bilibili/downloadQueue.ts";
import {formatCount, formatDate, formatDuration} from "../../../lib/format.ts";
import type {MediaCardItem} from "../../../lib/model.ts";
import IconChat from "../../icons/IconChat.tsx";
import IconEye from "../../icons/IconEye.tsx";
import QualitySelectors from "./QualitySelectors.tsx";

interface DownloadVideoCardProps {
  canDownload: boolean;
  coverDownloading: boolean;
  downloading: boolean;
  item: MediaCardItem;
  entry: PlayResolveEntry | undefined;
  progress: DownloadProgress | undefined;
  onDownloadCover: () => void;
  onDownload: () => void;
  onPickAudio: (audioId: number) => void;
  onPickQn: (qn: number) => void;
  onRemove: () => void;
}

// 下载列表里的单张视频卡片。
// 它只负责展示和把用户操作抛给父层；解析状态、画质切换、下载进度都由 useBilibiliDownloadQueue 管。
export default function DownloadVideoCard(props: DownloadVideoCardProps): JSXElement {
  function accessLabels(): string[] {
    // 充电专属、试看、付费等标签来自视频详情；即使解析失败也尽量展示原因。
    const entry = props.entry;
    if (entry?.status === "done") return entry.data.accessInfo.labels;
    if (entry?.status === "error") return entry.accessInfo?.labels ?? [];
    return [];
  }

  const progressText = () => {
    // 后端把视频、音频、合并、休眠阶段映射成同一条 0-100 的进度。
    const progress = props.progress;
    if (!progress) return "";
    if (progress.phase === "video") return "视频下载";
    if (progress.phase === "audio") return "音频下载";
    if (progress.phase === "merge") return "合并中";
    if (progress.phase === "sleep") return `休眠中 ${Math.max(0, Math.ceil(progress.sleepRemaining ?? 0))}s`;
    if (progress.phase === "done") return "完成";
    return "下载失败";
  };

  return (
      <article
          class="grid gap-3 rounded-lg border border-base-200 bg-base-100 p-2.5 shadow-sm md:grid-cols-[10rem_minmax(0,1fr)]">
        <div class="relative aspect-video w-full shrink-0 overflow-hidden rounded bg-base-200">
          <img
              class="h-full w-full object-cover"
              src={props.item.cover}
              alt={props.item.title}
              referrerPolicy="no-referrer"
          />
          <div
              class="pointer-events-none absolute inset-x-0 bottom-0 flex items-end justify-between bg-linear-to-t from-black/75 to-transparent px-2 pb-1.5 pt-6">
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

        <div class="grid min-w-0 content-between gap-2">
          <div class="grid grid-cols-[minmax(0,1fr)_3.5rem] items-start gap-2">
            <div class="min-w-0">
              <h3 class="line-clamp-2 text-sm font-semibold leading-5 text-base-content">
                {props.item.title}
              </h3>
              <div class="mt-1 flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 text-xs text-base-content/65">
                <span class="inline-flex min-w-0 items-center gap-1">
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    class="h-3.5 w-3.5 shrink-0"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                >
                    <path
                        fill-rule="evenodd"
                        d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"
                        clip-rule="evenodd"
                    />
                  </svg>
                <span class="truncate">{props.item.upperName || "未知 UP"}</span>
              </span>
                <span class="text-base-content/30">|</span>
                <span>{formatDate(props.item.pubtime)}</span>
              </div>
            </div>

            <Show when={props.downloading && props.progress} fallback={<div class="h-14 w-14"/>}>
              {(progress): JSXElement => (
                  // 固定宽度的右侧进度块占用标题区空白，避免下载状态改变时按钮行抖动。
                  <div
                      class="grid w-14 shrink-0 justify-self-end text-center justify-items-center gap-1">
                    <div
                        class="radial-progress text-info"
                        style={{
                          "--value": String(Math.round(progress().percent)),
                          "--size": "2.25rem",
                          "--thickness": "3px",
                        }}
                        role="progressbar"
                        aria-label={progressText()}
                    >
                      <span class="text-[0.6rem] font-medium tabular-nums text-base-content">
                        {Math.round(progress().percent)}
                      </span>
                    </div>
                    <span class="max-w-14 truncate text-[0.65rem] leading-none text-base-content/65">
                      {progressText()}
                    </span>
                  </div>
              )}
            </Show>
          </div>

          <Show when={accessLabels().length > 0}>
            <div class="flex flex-wrap gap-1">
              <For each={accessLabels()}>
                {(label): JSXElement => (
                    <span class="badge badge-warning badge-sm">{label}</span>
                )}
              </For>
            </div>
          </Show>

          <div class="flex flex-wrap items-center justify-end gap-2 border-t border-base-200 pt-2">
            <div class="mr-auto min-w-0">
              <QualitySelectors entry={props.entry} onPickAudio={props.onPickAudio} onPickQn={props.onPickQn}/>
            </div>
            <button
                class="btn btn-ghost btn-xs"
                type="button"
                onClick={props.onDownloadCover}
                disabled={props.coverDownloading || !props.item.cover}
            >
              {props.coverDownloading ? "保存中..." : "封面"}
            </button>
            <button class="btn btn-warning btn-xs" type="button" onClick={props.onRemove}>
              移除
            </button>
            <button
                class="btn btn-info btn-xs"
                type="button"
                onClick={props.onDownload}
                disabled={!props.canDownload || props.downloading}
            >
              {props.downloading ? "下载中..." : "下载"}
            </button>
          </div>
        </div>
      </article>
  );
}
