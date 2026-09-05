import {ParseVideo, VideoDetail} from "@bindings/github.com/kamiertop/videodown/douyin/api/douyin";
import {DownloadCover} from "@bindings/github.com/kamiertop/videodown/douyin/download/service.ts";
import * as model from "@bindings/github.com/kamiertop/videodown/douyin/model/models";
import {createFileRoute} from '@tanstack/solid-router'
import {createSignal, For, type JSXElement, Show} from "solid-js";
import MediaTypeBadge from "../../components/douyin/MediaTypeBadge.tsx";
import EmptyState from "../../components/EmptyState.tsx";
import NoCover from "../../components/NoCover.tsx";
import Toast from "../../components/Toast.tsx";
import {useToast} from "../../hooks/useToast.ts";
import {type DouyinDownloadProgress, useDouyinDownloadQueue} from "../../lib/douyin/downloadQueue.ts";
import {
  defaultDouyinVideoOption,
  douyinCoverCandidates,
  douyinDownloadAssets,
  douyinImageURLs,
  douyinMediaBadge,
  douyinMusicURL,
  douyinVideoOptions,
  formatDataSize,
} from "../../lib/douyin/media.ts";
import {
  addDouyinVideos,
  type DouyinDownloadItem,
  douyinVideoList,
  removeDouyinVideo,
  updateDouyinVideoOption,
} from "../../lib/douyin/store.ts";
import {formatCount, formatDate, formatDuration} from "../../lib/format.ts";

type AwemeItem = model.AwemeItem;

export const Route = createFileRoute('/douyin/download')({
  component: DouyinDownloadPage,
})

function normalizeDouyinDuration(value?: number): number {
  if (!value || value <= 0) return 0;
  return value >= 1000 ? Math.floor(value / 1000) : value;
}

function awemeCover(item: model.AwemeItem): string {
  return [
    ...(item.video?.raw_cover?.url_list ?? []),
    ...(item.video?.cover?.url_list ?? []),
    ...(item.video?.origin_cover?.url_list ?? []),
  ][0] ?? "";
}

function awemeTitle(item: model.AwemeItem): string {
  return item.item_title || item.desc || item.caption || `作品 ${item.aweme_id || ""}`.trim();
}

function detailToDownloadItem(item: model.AwemeItem): DouyinDownloadItem {
  const awemeId = item.aweme_id || item.group_id || item.sec_item_id;
  const title = awemeTitle(item);
  const cover = awemeCover(item);
  const duration = normalizeDouyinDuration(item.video?.duration ?? item.duration ?? 0);
  const authorName = item.author?.nickname || item.author?.uid || "未知作者";
  const videoOptions = douyinVideoOptions(item);
  const selectedVideoOption = defaultDouyinVideoOption(videoOptions);
  const mediaBadge = douyinMediaBadge(item);

  return {
    awemeId,
    sourceName: "",
    title,
    cover,
    coverCandidates: douyinCoverCandidates(item),
    duration,
    authorName,
    publishTime: item.create_time ?? 0,
    diggCount: item.statistics?.digg_count ?? 0,
    collectCount: item.statistics?.collect_count ?? 0,
    link: awemeId ? `https://www.douyin.com/video/${awemeId}` : undefined,
    videoURL: selectedVideoOption?.url,
    videoOptions,
    selectedVideoOptionId: selectedVideoOption?.id,
    imageURLs: douyinImageURLs(item),
    assets: mediaBadge ? douyinDownloadAssets(item) : undefined,
    musicURL: mediaBadge ? douyinMusicURL(item) : undefined,
    mediaBadge,
  };
}

function progressText(progress: DouyinDownloadProgress | undefined): string {
  // 后端把视频和图片合集都归一成同一条 0-100 进度，前端只区分阶段文案。
  if (!progress) return "";
  if (progress.phase === "video") return "视频下载";
  if (progress.phase === "image") return "图片下载";
  if (progress.phase === "music") return "音乐下载";
  if (progress.phase === "sleep") return `休眠中 ${Math.max(0, Math.ceil(progress.sleepRemaining ?? 0))}s`;
  if (progress.phase === "done") return "完成";
  return "下载失败";
}

function DouyinDownloadCard(props: {
  item: DouyinDownloadItem;
  canDownload: boolean;
  downloading: boolean;
  progress: DouyinDownloadProgress | undefined;
  coverDownloading: boolean;
  onDownloadCover: () => void;
  onDownload: () => void;
}): JSXElement {
  const mediaBadge = () => props.item.mediaBadge;
  const isStandardVideo = () => !mediaBadge();
  const selectedOption = () => props.item.videoOptions?.find((option) => option.id === props.item.selectedVideoOptionId);

  return (
      <article
          class="grid gap-3 rounded-lg border border-base-200 bg-base-100 p-2.5 shadow-sm md:grid-cols-[6rem_minmax(0,1fr)]">
        <div class="relative h-36 w-24 shrink-0 overflow-hidden rounded-lg bg-base-200">
          <Show when={props.item.cover} fallback={<NoCover/>}>
            <img
                src={props.item.cover}
                alt={props.item.title}
                class="h-full w-full object-cover"
                loading="lazy"
                decoding="async"
                referrerPolicy="no-referrer"
            />
          </Show>
          <Show when={mediaBadge() === "image"}>
            <MediaTypeBadge type="image"/>
          </Show>
          <Show when={mediaBadge() === "live-photo"}>
            <MediaTypeBadge type="live-photo"/>
          </Show>
          <Show when={!mediaBadge()}>
            <span
                class="absolute bottom-1 right-1 rounded-md bg-black/65 px-1.5 py-1 text-[0.65rem] font-medium tabular-nums leading-none text-white">
              {formatDuration(props.item.duration)}
            </span>
          </Show>
        </div>
        <div class="grid min-w-0 content-between gap-2">
          <div class="grid grid-cols-[minmax(0,1fr)_3.5rem] items-start gap-2">
            <div class="min-w-0">
              <h3 class="line-clamp-2 text-sm font-semibold leading-5 text-base-content" title={props.item.title}>
                {props.item.title}
              </h3>
              <p class="mt-1 line-clamp-1 text-xs text-base-content/50">@{props.item.authorName}</p>
            </div>

            <Show when={props.downloading && props.progress} fallback={<div class="h-14 w-14"/>}>
              {(progress) => (
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
                        aria-label={progressText(progress())}
                    >
                      <span class="text-[0.6rem] font-medium tabular-nums text-base-content">
                        {Math.round(progress().percent)}
                      </span>
                    </div>
                    <span class="max-w-14 truncate text-[0.65rem] leading-none text-base-content/65">
                      {progressText(progress())}
                    </span>
                  </div>
              )}
            </Show>
          </div>

          <div class="flex flex-wrap items-center gap-2 text-xs text-base-content/55">
            <span>发布 {props.item.publishTime ? formatDate(props.item.publishTime) : "-"}</span>
            <span class="rounded-full bg-base-200 px-2 py-0.5 tabular-nums">
              赞 {formatCount(props.item.diggCount)}
            </span>
            <span class="rounded-full bg-base-200 px-2 py-0.5 tabular-nums">
              藏 {formatCount(props.item.collectCount)}
            </span>
          </div>

          <Show when={!props.canDownload}>
            <p class="text-xs text-warning">没有可用下载地址，可能需要重新进入详情页刷新数据。</p>
          </Show>

          <div class="flex flex-wrap items-center justify-end gap-2 border-t border-base-200 pt-2">
            <Show when={isStandardVideo()}>
              <div class="mr-auto flex min-w-0 flex-wrap items-center gap-3 text-xs">
                <Show when={(props.item.videoOptions?.length ?? 0) > 0}>
                  <span class="shrink-0 text-base-content/55">清晰度</span>
                  <select
                      class="select select-bordered select-xs w-44 min-w-0"
                      value={props.item.selectedVideoOptionId ?? ""}
                      disabled={props.downloading}
                      onChange={(event) => updateDouyinVideoOption(props.item.awemeId, event.currentTarget.value)}
                  >
                    <For each={props.item.videoOptions ?? []}>
                      {(option) => (
                          <option value={option.id}>
                            {option.gearName} · {formatDataSize(option.dataSize)}
                          </option>
                      )}
                    </For>
                  </select>
                </Show>
                <span class="shrink-0 text-base-content/55">编码</span>
                <span class="truncate text-base-content/70">
                  {selectedOption()?.codec ?? "-"}
                  <Show when={selectedOption()?.bitRate}>
                    {(bitRate) => ` · ${Math.round(bitRate() / 1000)} kbps`}
                  </Show>
                </span>
              </div>
            </Show>
            <button
                class="btn btn-ghost btn-xs"
                type="button"
                onClick={props.onDownloadCover}
                disabled={props.coverDownloading || (props.item.coverCandidates?.length ?? 0) === 0}
            >
              {props.coverDownloading ? "保存中..." : "封面"}
            </button>
            <button class="btn btn-warning btn-xs" type="button"
                    onClick={() => removeDouyinVideo(props.item.awemeId)}
                    disabled={props.downloading}>
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

function DouyinDownloadPage(): JSXElement {
  const [videoURL, setVideoURL] = createSignal("");
  const [parsing, setParsing] = createSignal(false);
  const [coverDownloadingIDs, setCoverDownloadingIDs] = createSignal<string[]>([]);
  const {message, type, showToast} = useToast();
  // 下载状态集中在 hook 中，页面只负责渲染列表和把用户操作转发给队列。
  const queue = useDouyinDownloadQueue(showToast);

  function setCoverDownloading(awemeId: string, downloading: boolean): void {
    setCoverDownloadingIDs((prev) => {
      const next = new Set(prev);
      if (downloading) next.add(awemeId);
      else next.delete(awemeId);
      return [...next];
    });
  }

  async function downloadCover(item: DouyinDownloadItem): Promise<void> {
    const covers = item.coverCandidates ?? [];
    if (covers.length === 0) {
      showToast("当前内容没有可用封面", "warning");
      return;
    }
    if (coverDownloadingIDs().includes(item.awemeId)) return;

    setCoverDownloading(item.awemeId, true);
    try {
      const path = await DownloadCover(covers, ({
        awemeId: item.awemeId,
        sourceName: item.sourceName ?? "",
        title: item.title || item.awemeId || "cover",
        cover: item.cover,
        duration: item.duration,
        authorName: item.authorName,
        publishTime: item.publishTime ?? 0,
        diggCount: item.diggCount ?? 0,
        collectCount: item.collectCount ?? 0,
        videoURL: "",
        imageURLs: [],
        assets: [],
        musicURL: "",
      }));
      showToast(`封面已保存：${path}`, "success");
    } catch (error) {
      showToast(error instanceof Error ? error.message : String(error), "error");
    } finally {
      setCoverDownloading(item.awemeId, false);
    }
  }

  async function parseVideo(): Promise<void> {
    if (parsing()) return;

    const input = videoURL().trim();
    if (!input) {
      showToast("请输入抖音视频链接或视频 ID", "error");
      return;
    }

    setParsing(true);
    try {
      // 复制分享文案时通常包含短链；如果用户直接输入 awemeId，就跳过重定向解析。
      const awemeId = input.includes("http") ? await ParseVideo(input) : input;
      if (!awemeId) {
        showToast("未能解析出视频 ID", "error");
        return;
      }

      const detail: AwemeItem = await VideoDetail(awemeId);
      const item = detailToDownloadItem(detail);
      if (!item.awemeId) {
        showToast("解析成功，但详情中没有视频 ID", "error");
        return;
      }

      addDouyinVideos([item]);
      setVideoURL("");
      showToast(`已添加：${item.title || item.awemeId}`, "success");
    } catch (error) {
      showToast(error instanceof Error ? error.message : String(error), "error");
    } finally {
      setParsing(false);
    }
  }

  return (
      <section class="flex h-full min-h-0 flex-col p-3">
        <section class="flex flex-row join gap-2">
          <input
              type="text"
              placeholder="请输入抖音视频分享链接、视频 ID、精选链接，可按回车直接解析"
              value={videoURL()}
              onInput={(event) => setVideoURL(event.currentTarget.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter") void parseVideo();
              }}
              class="input input-success w-full"
              disabled={parsing()}
          />
          <button
              class="btn btn-outline btn-secondary"
              type="button"
              onClick={() => void parseVideo()}
              disabled={parsing()}
          >
            {parsing() ? "解析中..." : "解析"}
          </button>
          <button
              class="btn btn-outline btn-info"
              type="button"
              onClick={() => setVideoURL("")}
              disabled={parsing()}
          >
            清空
          </button>

        </section>

        <Show when={douyinVideoList().length > 0}>
          <section class="mt-2 flex flex-row items-center justify-between rounded-lg p-3 shadow-sm">
            <div class="flex min-w-0 flex-1 flex-col gap-1">
              <div class="flex items-center gap-2">
                <div class="badge badge-primary">{douyinVideoList().length}</div>
                <span class="text-xs">个内容待下载</span>
              </div>
            </div>
            <button
                class="btn btn-success btn-xs gap-1.5"
                type="button"
                onClick={() => void queue.startDownload()}
                disabled={queue.downloading()}
            >
              {queue.downloading() ? "下载中..." : "开始下载"}
            </button>
          </section>
        </Show>

        <div class="mt-3 min-h-0 flex-1 overflow-hidden rounded-lg border border-base-300 bg-base-100">
          <Show
              when={douyinVideoList().length > 0}
              fallback={<EmptyState title="下载列表为空"
                                    description="可以解析视频链接，或从收藏、合集、用户页勾选后加入下载列表。"/>}
          >
            <div class="flex h-full flex-col gap-1.5 overflow-auto p-2">
              <For each={douyinVideoList()}>
                {(item) => (
                    <DouyinDownloadCard
                        item={item}
                        canDownload={queue.canDownload(item)}
                        downloading={queue.isDownloading(item)}
                        progress={queue.progressFor(item)}
                        coverDownloading={coverDownloadingIDs().includes(item.awemeId)}
                        onDownloadCover={() => void downloadCover(item)}
                        onDownload={() => void queue.downloadOne(item)}
                    />
                )}
              </For>
            </div>
          </Show>
        </div>
        <Toast message={message()} type={type()}/>
      </section>
  )
}
