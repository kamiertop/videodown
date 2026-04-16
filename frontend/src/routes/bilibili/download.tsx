import {createFileRoute} from '@tanstack/solid-router'
import {createEffect, createMemo, createSignal, For, type JSXElement, Show,} from "solid-js";
import {DownloadVideoByDash} from "../../../wailsjs/go/api/BiliBili";
import {model} from "../../../wailsjs/go/models";
import IconChat from "../../components/icons/IconChat.tsx";
import IconEye from "../../components/icons/IconEye.tsx";
import Toast from "../../components/Toast";
import {useToast} from "../../hooks/useToast";
import {
  audioDetailTitle,
  audioSelectLabel,
  bvidCacheKey,
  refetchBilibiliPlayAtQn,
  resolveBilibiliPlayUrl,
  type ResolvedPlayInfo,
  sortedAudioTracks,
  sortedSupportFormats,
  supportFormatDetailTitle,
  supportFormatSelectLabel,
  switchResolvedAudio,
  switchResolvedPlayAtQn,
} from "../../lib/bilibiliPlayResolve.ts";
import {removeVideo, videoList} from "../../lib/bilibiliStore.ts";
import {extractBvid, formatCount, formatDate, formatDuration} from "../../lib/format";
import type {MediaCardItem} from "../../lib/model.ts";

/** 同一 BV 仅一条解析任务在飞，避免 createEffect 重复触发 */
const playResolveInFlight = new Set<string>();

type PlayResolveEntry =
  | { status: "loading" }
  | { status: "done"; data: ResolvedPlayInfo }
  | { status: "error"; message: string };

function streamBaseUrl(v: { baseUrl?: string; base_url?: string }): string {
  return (v.baseUrl || v.base_url || "").trim();
}

function SupportFormatsForItem(props: {
  entry: () => PlayResolveEntry | undefined;
  busy: () => boolean;
  onPickQn: (qn: number) => void;
}): JSXElement {
  const done = createMemo(() => {
    const e = props.entry();
    return e?.status === "done" ? e.data : null;
  });
  const rows = createMemo<model.SupportFormat[]>(() => {
    const data = done();
    if (!data) return [];
    return sortedSupportFormats(data.play.support_formats ?? []);
  });

  return (
    <div class="flex min-w-0 flex-1 flex-col gap-1.5">
      <Show when={done()}>
        {(data) => (
          <label class="flex flex-col gap-1">
            <div class="flex max-w-full items-center gap-2 sm:max-w-md">
              <Show
                when={rows().length > 0}
                fallback={
                  <p class="text-sm text-base-content/70">暂无可用档位数据</p>
                }
              >
                <div>
                  画质：
                </div>
                <select
                  class="select select-info max-w-30 flex-1 bg-base-100 font-medium text-base-content"
                  value={String(data().selectedQn)}
                  disabled={props.busy()}
                  onChange={(ev) => {
                    const qn = Number(ev.currentTarget.value);
                    if (Number.isFinite(qn)) props.onPickQn(qn);
                  }}
                >
                  <For each={rows()}>
                    {(fmt) => (
                      <option value={String(fmt.quality)} title={supportFormatDetailTitle(fmt)}>
                        {supportFormatSelectLabel(fmt)}
                      </option>
                    )}
                  </For>
                </select>
              </Show>
              <Show when={props.busy()}>
                <span class="loading loading-spinner loading-md shrink-0 text-primary"/>
              </Show>
            </div>
          </label>
        )}
      </Show>
    </div>
  );
}

function AudioTracksForItem(props: {
  entry: () => PlayResolveEntry | undefined;
  onPickAudio: (audioId: number) => void;
}): JSXElement {
  const done = createMemo(() => {
    const e = props.entry();
    return e?.status === "done" ? e.data : null;
  });
  const tracks = createMemo(() => {
    const data = done();
    if (!data) return [];
    return sortedAudioTracks(data.play.dash?.audio);
  });

  return (
    <div class="flex min-w-0 flex-1 flex-col gap-1.5">
      <Show when={done()}>
        {(data) => (
          <Show when={tracks().length > 0}>
            <label class="flex flex-col gap-1">
              <div class="flex max-w-full items-center gap-2 sm:max-w-md">
                <div>
                  音质：
                </div>
                <select
                  class="select select-info max-w-35 flex-1 bg-base-100 font-medium text-base-content"
                  value={String(data().bestAudio?.id ?? "")}
                  onChange={(ev) => {
                    const id = Number(ev.currentTarget.value);
                    if (Number.isFinite(id)) props.onPickAudio(id);
                  }}
                >
                  <For each={tracks()}>
                    {(a) => (
                      <option value={String(a.id)} title={audioDetailTitle(a)}>
                        {audioSelectLabel(a)}
                      </option>
                    )}
                  </For>
                </select>
              </div>
            </label>
          </Show>
        )}
      </Show>
    </div>
  );
}

function formatListSource(item: MediaCardItem): string {
  const kind = item.sourceListKind;
  let name = item.sourceListName;
  if (kind === "全部投稿") {
    return kind;
  }

  return `${kind}「${name}」`;
}

export const Route = createFileRoute('/bilibili/download')({
  component: DownLoad,
})


function DownLoad(): JSXElement {
  const [videoURL, setVideoURL] = createSignal<string>("");
  const [playResolveByBvid, setPlayResolveByBvid] = createSignal<Record<string, PlayResolveEntry>>({});
  const [qnLoadingByBvid, setQnLoadingByBvid] = createSignal<Record<string, boolean>>({});
  const [downloading, setDownloading] = createSignal<boolean>(false);
  const {message, type, showToast} = useToast();

  const listSourceSummary = createMemo(() => {
    const list = videoList();
    const labels = [...new Set(list.map(i => formatListSource(i)).filter((s): s is string => !!s))];
    if (labels.length === 0) return "";
    if (labels.length === 1) return `来源：${labels[0]}`;
    return `来源：${labels.join("、")}`;
  });

  createEffect(() => {
    const list = videoList();
    const bvSet = new Set(
      list.map((i) => bvidCacheKey(i.bvid)).filter((k): k is string => !!k),
    );

    setPlayResolveByBvid((prev) => {
      const next: Record<string, PlayResolveEntry> = {...prev};
      for (const k of Object.keys(next)) {
        if (!bvSet.has(k)) {
          playResolveInFlight.delete(k);
          delete next[k];
        }
      }
      return next;
    });

    setQnLoadingByBvid((prev) => {
      const next: Record<string, boolean> = {...prev};
      for (const k of Object.keys(next)) {
        if (!bvSet.has(k)) delete next[k];
      }
      return next;
    });

    const map = playResolveByBvid();

    for (const item of list) {
      const key = bvidCacheKey(item.bvid);
      if (!key) continue;
      if (playResolveInFlight.has(key)) continue;

      const cur = map[key];
      if (cur?.status === "loading" || cur?.status === "done" || cur?.status === "error") {
        continue;
      }

      playResolveInFlight.add(key);
      setPlayResolveByBvid((p) => ({...p, [key]: {status: "loading"}}));

      void (async () => {
        try {
          const data = await resolveBilibiliPlayUrl(item.bvid);
          if (!videoList().some((v) => bvidCacheKey(v.bvid) === key)) return;
          setPlayResolveByBvid((p) => ({...p, [key]: {status: "done", data}}));
        } catch (e) {
          if (!videoList().some((v) => bvidCacheKey(v.bvid) === key)) return;
          const message = e instanceof Error ? e.message : String(e);
          setPlayResolveByBvid((p) => ({...p, [key]: {status: "error", message}}));
        } finally {
          playResolveInFlight.delete(key);
        }
      })();
    }
  });

  async function handlePickQn(key: string, qn: number): Promise<void> {
    const cur = playResolveByBvid()[key];
    if (cur?.status !== "done" || qn === cur.data.selectedQn) return;

    const local = switchResolvedPlayAtQn(cur.data, qn);
    if (local) {
      setPlayResolveByBvid((p) => ({...p, [key]: {status: "done", data: local}}));
      return;
    }

    setQnLoadingByBvid((p) => ({...p, [key]: true}));
    try {
      const next = await refetchBilibiliPlayAtQn({
        aid: cur.data.aid,
        cid: cur.data.cid,
        bvid: cur.data.bvid,
        partCount: cur.data.partCount,
        qn,
        preferredAudioId: cur.data.bestAudio?.id,
      });
      if (!videoList().some((v) => bvidCacheKey(v.bvid) === key)) return;
      setPlayResolveByBvid((p) => ({...p, [key]: {status: "done", data: next}}));
    } catch (e) {
      showToast(e instanceof Error ? e.message : String(e), "error");
    } finally {
      setQnLoadingByBvid((p) => ({...p, [key]: false}));
    }
  }

  function handlePickAudio(key: string, audioId: number): void {
    const cur = playResolveByBvid()[key];
    if (cur?.status !== "done") return;
    if (cur.data.bestAudio?.id === audioId) return;
    const next = switchResolvedAudio(cur.data, audioId);
    if (next) {
      setPlayResolveByBvid((p) => ({...p, [key]: {status: "done", data: next}}));
    }
  }

  function parseVideo() {
    const bvid = extractBvid(videoURL());
    if (bvid === null) {
      showToast('请输入有效的 B 站视频链接或 BV 号', 'error');
      return;
    }

    showToast(`成功解析 BV 号: ${bvid}`, 'success');
  }

  async function startDownload(): Promise<void> {
    if (downloading()) return;
    const items = videoList();
    if (items.length === 0) {
      showToast("暂无可下载视频", "warning");
      return;
    }

    const tasks = items
      .map((item) => {
        const key = bvidCacheKey(item.bvid);
        if (!key) return null;
        const entry = playResolveByBvid()[key];
        if (entry?.status !== "done") return null;

        const videoURL = streamBaseUrl(entry.data.bestVideo);
        if (!videoURL) return null;

        const audioURL = entry.data.bestAudio ? streamBaseUrl(entry.data.bestAudio) : "";
        return {item, videoURL, audioURL};
      })
      .filter((v): v is { item: MediaCardItem; videoURL: string; audioURL: string } => v !== null);

    if (tasks.length === 0) {
      showToast("暂无可用流地址，请稍候重试", "warning");
      return;
    }

    setDownloading(true);
    let success = 0;
    let failed = 0;
    try {
      for (const task of tasks) {
        try {
          let dirName = "";
          if (task.item.sourceListKind === "全部投稿") {
            dirName = task.item.upperName;
          } else {
            dirName = task.item.sourceListName ?? "";
          }
          await DownloadVideoByDash(dirName, task.item.bvid, task.item.title, task.videoURL, task.audioURL);
          // 成功之后删除
          removeVideo(task.item.id);
          success += 1;
        } catch (e) {
          failed += 1;
          const msg = e instanceof Error ? e.message : String(e);
          showToast(`下载失败：${task.item.title}，${msg}`, "error");
        }
      }
    } finally {
      setDownloading(false);
    }

    if (failed === 0) {
      showToast(`下载完成：成功 ${success} 个`, "success");
      return;
    }
    showToast(`下载完成：成功 ${success} 个，失败 ${failed} 个`, "warning");
  }

  return (
    <div class={"flex flex-col pt-4 pl-4 pr-4 pb-4 h-full"}>
      <section>
        <div class={"flex flex-row join gap-2"}>
          <input type="text" placeholder="请输入视频链接, 支持BV号、AV号、视频URL等格式, 可按回车直接解析"
                 value={videoURL()}
                 onInput={(e) => setVideoURL(e.currentTarget.value)}
                 onkeydown={(e) => {
                   if (e.key === "Enter") {
                     parseVideo();
                   }
                 }}
                 class="input input-success w-full"/>
          <div class={"btn btn-outline btn-secondary"}
               onClick={parseVideo}
          >
            解析
          </div>
          <div class={"btn btn-outline btn-info"}
               onClick={() => setVideoURL("")}
          >
            清空
          </div>
        </div>
      </section>
      <Show when={videoList().length > 0}>
        <section
          class="mt-2 rounded-lg p-3 flex flex-row justify-between items-center shadow-sm">
          <div class="flex min-w-0 flex-1 flex-col gap-1">
            <div class="flex items-center gap-2">
              <div class="badge badge-primary">{videoList().length}</div>
              <span class="text-xs">个视频待下载</span>
            </div>
            <Show when={listSourceSummary()}>
              <p class="truncate text-sm text-base-content/80">
                {listSourceSummary()}
              </p>
            </Show>
          </div>
          <button class="btn btn-success btn-xs gap-1.5" onClick={() => void startDownload()}
                  disabled={downloading()}>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24"
                 stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/>
            </svg>
            {downloading() ? "下载中..." : "开始下载"}
          </button>
        </section>
      </Show>
      <section class="mt-3 flex flex-1 flex-col gap-3 overflow-y-auto pr-4">
        <For each={videoList()}>
          {(item): JSXElement => {
            const listSrc = formatListSource(item);
            return (
              // 水平布局，左侧图片+弹幕量+播放量+时长
              <article class="flex flex-row gap-3">
                <div class="relative w-40 aspect-video object-cover"
                >
                  <img
                    class="h-full w-full object-cover"
                    src={item.cover}
                    alt={item.title}
                    referrerPolicy="no-referrer"
                  />
                  <div
                    class="pointer-events-none absolute inset-x-0 bottom-0 flex items-end justify-between bg-linear-to-t from-black/75 to-transparent px-2 pb-1.5 pt-6"
                  >
                    <div class="flex items-center gap-2 text-white/90">
                        <span class="flex items-center gap-0.5 text-[11px]">
                            <IconEye class="h-3 w-3"/>{formatCount(item.play)}
                        </span>
                      <span class="flex items-center gap-0.5 text-[11px]">
                          <IconChat class="h-3 w-3"/> {item.danmaku}
                      </span>
                    </div>
                    <span class="rounded bg-black/65 px-1 py-0.5 text-[11px] tabular-nums text-white/95">
                        {formatDuration(item.duration)}
                    </span>
                  </div>
                </div>
                {/*  右侧上下布局，上边是title*/}
                <div class="flex min-w-0 flex-1 flex-col gap-2">
                  <h3 class="text-base font-semibold leading-snug text-base-content line-clamp-2">
                    {item.title}
                  </h3>
                  {/*  下边水平布局*/}
                  <div class="flex flex-row gap-3 justify-between">
                    <div class="mt-4 w-64">
                      <Show when={listSrc}>
                        <p class="text-sm font-medium text-primary">{listSrc}</p>
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
                          {item.upperName}
                    </span>
                        <span class="text-base-content/35">|</span>
                        <span>{formatDate(item.pubtime)}</span>
                      </div>
                    </div>
                    <div class="flex flex-row gap-1 mt-4 flex-1">
                      <SupportFormatsForItem
                        entry={() => {
                          const k = bvidCacheKey(item.bvid);
                          return k ? playResolveByBvid()[k] : undefined;
                        }}
                        busy={() => {
                          const k = bvidCacheKey(item.bvid);
                          return k ? qnLoadingByBvid()[k] : false;
                        }}
                        onPickQn={(qn) => {
                          const k = bvidCacheKey(item.bvid);
                          if (k) void handlePickQn(k, qn);
                        }}
                      />
                      <AudioTracksForItem
                        entry={() => {
                          const k = bvidCacheKey(item.bvid);
                          return k ? playResolveByBvid()[k] : undefined;
                        }}
                        onPickAudio={(audioId) => {
                          const k = bvidCacheKey(item.bvid);
                          if (k) handlePickAudio(k, audioId);
                        }}
                      />
                    </div>
                    <div class="flex flex-col gap-2">
                      <div class="btn btn-warning btn-sm" onClick={() => removeVideo(item.id)}>
                        移除
                      </div>
                      <div class="btn btn-info btn-sm">
                        下载
                      </div>
                    </div>
                  </div>
                </div>
              </article>
            );
          }}
        </For>
      </section>
      <Toast message={message()} type={type()}/>
    </div>
  )
}
