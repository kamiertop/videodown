import {createMemo, For, Show, type JSXElement} from "solid-js";
import {
    audioDetailTitle,
    audioSelectLabel,
    sortedAudioTracks,
    sortedDashVideoQualities,
} from "../../../lib/bilibiliPlayResolve.ts";
import type {PlayResolveEntry} from "../../../lib/bilibiliDownloadQueue.ts";

interface QualitySelectorsProps {
    entry: PlayResolveEntry | undefined;
    onPickAudio: (audioId: number) => void;
    onPickQn: (qn: number) => void;
}

export default function QualitySelectors(props: QualitySelectorsProps): JSXElement {
    const done = createMemo(() => {
        const e = props.entry;
        return e?.status === "done" ? e.data : null;
    });
    const qualities = createMemo(() => {
        const data = done();
        return data ? sortedDashVideoQualities(data.play) : [];
    });
    const tracks = createMemo(() => {
        const data = done();
        return data ? sortedAudioTracks(data.play.dash?.audio) : [];
    });

    return (
        <div class="flex flex-row gap-1 mt-4 flex-1">
            <Show when={done()}>
                {(data) => (
                    <>
                        <div class="flex min-w-0 flex-1 flex-col gap-1.5">
                            <Show
                                when={qualities().length > 0}
                                fallback={<p class="text-sm text-base-content/70">暂无可用画质数据</p>}
                            >
                                <label class="flex flex-col gap-1">
                                    <div class="flex max-w-full items-center gap-2 sm:max-w-md">
                                        <div>画质：</div>
                                        <select
                                            class="select select-info max-w-30 flex-1 bg-base-100 font-medium text-base-content"
                                            value={String(data().selectedQn)}
                                            onChange={(ev) => {
                                                const qn = Number(ev.currentTarget.value);
                                                if (Number.isFinite(qn)) props.onPickQn(qn);
                                            }}
                                        >
                                            <For each={qualities()}>
                                                {(quality) => (
                                                    <option value={String(quality.qn)} title={quality.title}>
                                                        {quality.label}
                                                    </option>
                                                )}
                                            </For>
                                        </select>
                                    </div>
                                </label>
                            </Show>
                        </div>
                        <div class="flex min-w-0 flex-1 flex-col gap-1.5">
                            <Show when={tracks().length > 0}>
                                <label class="flex flex-col gap-1">
                                    <div class="flex max-w-full items-center gap-2 sm:max-w-md">
                                        <div>音质：</div>
                                        <select
                                            class="select select-info max-w-35 flex-1 bg-base-100 font-medium text-base-content"
                                            value={String(data().bestAudio?.id ?? "")}
                                            onChange={(ev) => {
                                                const id = Number(ev.currentTarget.value);
                                                if (Number.isFinite(id)) props.onPickAudio(id);
                                            }}
                                        >
                                            <For each={tracks()}>
                                                {(audio) => (
                                                    <option value={String(audio.id)} title={audioDetailTitle(audio)}>
                                                        {audioSelectLabel(audio)}
                                                    </option>
                                                )}
                                            </For>
                                        </select>
                                    </div>
                                </label>
                            </Show>
                        </div>
                    </>
                )}
            </Show>
            <Show when={props.entry?.status === "loading"}>
                <span class="loading loading-spinner loading-md shrink-0 text-primary"/>
            </Show>
            <Show when={props.entry?.status === "error"}>
                <p class="text-sm text-error">{props.entry?.status === "error" ? props.entry.message : ""}</p>
            </Show>
        </div>
    );
}
