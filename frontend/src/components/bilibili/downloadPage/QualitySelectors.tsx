import {createMemo, For, type JSXElement, Show} from "solid-js";
import type {PlayResolveEntry} from "../../../lib/bilibili/downloadQueue.ts";
import {
  audioDetailTitle,
  audioSelectLabel,
  sortedAudioTracks,
  sortedDashVideoQualities,
} from "../../../lib/bilibili/playResolve.ts";

interface QualitySelectorsProps {
  entry: PlayResolveEntry | undefined;
  onPickAudio: (audioId: number) => void;
  onPickQn: (qn: number) => void;
}

// 画质/音质选择器只展示已经解析好的 DASH 数据。
// 选择画质或音质时不会重新请求后端，只把当前卡片选中的 bestVideo/bestAudio 换成 dash 里已有的流。
export default function QualitySelectors(props: QualitySelectorsProps): JSXElement {
  const done = createMemo(() => {
    const e = props.entry;
    return e?.status === "done" ? e.data : null;
  });
  const qualities = createMemo(() => {
    const data = done();
    // 这里只列出 dash.video 中实际有 baseUrl/base_url 的画质，support_formats 只用来补展示文案。
    return data ? sortedDashVideoQualities(data.play) : [];
  });
  const tracks = createMemo(() => {
    const data = done();
    // 音频同理：按 dash.audio 里的真实音轨展示，不额外请求。
    return data ? sortedAudioTracks(data.play.dash?.audio) : [];
  });

  return (
      <div class="flex min-w-0 flex-wrap items-center justify-start gap-2">
        <Show when={done()}>
          {(data) => (
              <>
                <div class="min-w-0">
                  <Show
                      when={qualities().length > 0}
                      fallback={<p class="text-xs text-base-content/70">暂无画质</p>}
                  >
                    <label>
                      <div class="flex items-center gap-1.5">
                        <div class="shrink-0 text-xs text-base-content/60">画质</div>
                        <select
                            class="select select-info select-xs w-28 bg-base-100 font-medium text-base-content"
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

                <div class="min-w-0">
                  <Show when={tracks().length > 0}>
                    <label>
                      <div class="flex items-center gap-1.5">
                        <div class="shrink-0 text-xs text-base-content/60">音质</div>
                        <select
                            class="select select-info select-xs w-30 bg-base-100 font-medium text-base-content"
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
          <span class="loading loading-spinner loading-sm shrink-0 text-primary"/>
        </Show>
        <Show when={props.entry?.status === "error"}>
          <p class="text-xs text-error">{props.entry?.status === "error" ? props.entry.message : ""}</p>
        </Show>
      </div>
  );
}
