import {For, type JSXElement, Show} from "solid-js";
import {formatCount, formatDate, formatDuration} from "../lib/format";
import IconEye from "./icons/IconEye";
import IconChat from "./icons/IconChat";
import type {MediaCardItem} from "../lib/model.ts";

/**
 * 视频卡片网格组件
 */

export default function VideoCardGrid(props: {
    medias: MediaCardItem[];
    selectedSet: () => Set<number>;
    onToggleSelect: (id: number) => void;
    onDownloadOne: (media: MediaCardItem) => void;
}): JSXElement {
    return (
        <div class="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-4">
            <For each={props.medias}>
                {(media) => (
                    <div
                        class={`group relative flex flex-col rounded-lg transition-all duration-150 hover:shadow-md ${
                            props.selectedSet().has(media.id)
                                ? 'border-2 border-success bg-success/5 shadow-sm shadow-success/20'
                                : 'border border-base-300 bg-base-100 hover:border-base-content/20'
                        }`}
                    >
                        <div
                            class="relative aspect-video w-full cursor-pointer overflow-hidden rounded-t-lg bg-base-200"
                            onClick={() => props.onToggleSelect(media.id)}>
                            <Show
                                when={media.cover !== ""}
                                fallback={
                                    <div
                                        class="flex h-full w-full items-center justify-center text-xs text-base-content/40">
                                        无封面
                                    </div>
                                }
                            >
                                <img src={media.cover} alt={media.title}
                                     class="h-full w-full object-cover transition-transform duration-200 group-hover:scale-105"
                                     loading="lazy"
                                     referrerPolicy="no-referrer"
                                     decoding="async"/>
                            </Show>

                            <div
                                class="pointer-events-none absolute inset-x-0 bottom-0 flex items-end justify-between bg-linear-to-t from-black/70 to-transparent px-2 pb-1.5 pt-5">
                                <div class="flex items-center gap-2.5 text-white/90">
                                    <span class="flex items-center gap-0.5 text-xs">
                                        <IconEye class="h-3 w-3"/>
                                        {formatCount(media.play)}
                                    </span>
                                    <span class="flex items-center gap-0.5 text-xs">
                                        <IconChat class="h-3 w-3"/>
                                        {formatCount(media.danmaku)}
                                    </span>
                                </div>
                                <span
                                    class="rounded bg-black/60 px-1 py-0.5 text-xs tabular-nums text-white/90">
                                    {formatDuration(media.duration)}
                                </span>
                            </div>
                        </div>

                        <div class="flex min-h-0 flex-1 cursor-pointer flex-col px-2.5 pb-2.5 pt-2"
                             onClick={() => props.onToggleSelect(media.id)}>
                            <p
                                class="line-clamp-2 min-h-10 text-sm font-semibold leading-5 text-base-content"
                                title={media.title}
                            >
                                {media.title}
                            </p>
                            <div class="mt-1.5 flex items-center justify-between">
                                <p class="flex min-w-0 flex-1 items-center gap-1 truncate text-xs text-base-content/55">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 shrink-0"
                                         viewBox="0 0 20 20" fill="currentColor">
                                        <path fill-rule="evenodd"
                                              d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"
                                              clip-rule="evenodd"/>
                                    </svg>
                                    <span class="truncate">{media.upperName}</span>
                                </p>
                                <div class="flex shrink-0 items-center gap-1.5">
                                    <Show when={media.pubtime}>
                                        <span
                                            class="text-xs tabular-nums text-base-content/45">{formatDate(media.pubtime)}</span>
                                    </Show>
                                    <button
                                        class="btn btn-ghost btn-xs opacity-0 transition-opacity group-hover:opacity-100"
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            props.onDownloadOne(media);
                                        }}
                                    >
                                        下载
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                )}
            </For>
        </div>
    );
}
