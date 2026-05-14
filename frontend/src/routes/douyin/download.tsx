import {createFileRoute} from '@tanstack/solid-router'
import {createSignal, For, type JSXElement, Show} from "solid-js";
import {DownloadCover, ParseVideo, VideoDetail} from "../../../wailsjs/go/api/Douyin";
import {model} from "../../../wailsjs/go/models";
import EmptyState from "../../components/EmptyState.tsx";
import Toast from "../../components/Toast.tsx";
import {useToast} from "../../hooks/useToast.ts";
import {type DouyinDownloadProgress, useDouyinDownloadQueue} from "../../lib/douyin/downloadQueue.ts";
import {
  defaultDouyinVideoOption,
  douyinCoverCandidates,
  douyinDownloadAssets,
  douyinImageURLs,
  douyinMusicURL,
  douyinVideoOptions,
  formatDataSize,
  isDouyinImageAlbum,
  isDouyinLivePhoto,
} from "../../lib/douyin/media.ts";
import {
  addDouyinVideos,
  type DouyinDownloadItem,
  douyinVideoList,
  removeDouyinVideo,
  updateDouyinVideoOption,
} from "../../lib/douyin/store.ts";
import {formatCount, formatDate, formatDuration} from "../../lib/format.ts";

export const Route = createFileRoute('/douyin/download')({
  component: DouyinDownloadPage,
})

function normalizeDouyinDuration(value?: number): number {
  if (!value || value <= 0) return 0;
  return value >= 1000 ? Math.floor(value / 1000) : value;
}

function awemeCover(item: model.AwemeItem): string {
  return item.video?.cover?.url_list?.[0]
      ?? item.video?.origin_cover?.url_list?.[0]
      ?? "";
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
  const mediaBadge = isDouyinLivePhoto(item)
      ? "live-photo"
      : isDouyinImageAlbum(item) ? "image" : undefined;

  return {
    awemeId,
    sourceKind: "解析结果",
    sourceName: "手动解析",
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
  if (progress.phase === "video") return `视频下载 ${Math.round(progress.percent)}%`;
  if (progress.phase === "image") return `图片下载 ${Math.round(progress.percent)}%`;
  if (progress.phase === "music") return `音乐下载 ${Math.round(progress.percent)}%`;
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
      <article class="flex flex-col gap-3 rounded-lg border border-base-300 bg-base-100 p-2 md:flex-row">
        <div class="relative h-36 w-24 shrink-0 overflow-hidden rounded-lg bg-base-200">
          <Show
              when={props.item.cover}
              fallback={<div
                  class="absolute inset-0 flex items-center justify-center text-xs text-base-content/35">无封面</div>}
          >
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
            <svg
                width="12"
                height="12"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
                class="absolute right-1.5 top-1.5 z-10 h-3.5 w-3.5 drop-shadow-[0_1px_2px_rgba(0,0,0,0.65)]"
                viewBox="0 0 12 12"
                aria-label="图文"
            >
              <path
                  d="M1.455 0C.65 0 0 .651 0 1.455V8c0 .803.651 1.455 1.455 1.455H8c.803 0 1.455-.652 1.455-1.455V1.455C9.455.65 8.803 0 8 0H1.455z"
                  fill="#fff"></path>
              <path
                  d="M4 12a1.455 1.455 0 0 1-1.455-1.454h5.819a2.182 2.182 0 0 0 2.181-2.182V2.545C11.35 2.545 12 3.197 12 4v5.09A2.909 2.909 0 0 1 9.09 12H4z"
                  fill="#fff"></path>
            </svg>
          </Show>
          <Show when={mediaBadge() === "live-photo"}>
            <svg
                class="absolute right-1.5 top-1.5 z-10 h-4 w-4 text-white drop-shadow-[0_1px_2px_rgba(0,0,0,0.65)]"
                viewBox="0 0 1024 1024"
                xmlns="http://www.w3.org/2000/svg"
                aria-label="动态照片"
            >
              <path
                  d="M209.933227 195.114667c9.813333 0 19.626667-3.413333 27.306666-9.813334 18.346667-15.36 20.48-42.24 5.12-60.16a41.770667 41.770667 0 0 0-59.733333-5.12c-18.346667 14.933333-20.48 41.813333-5.546667 59.733334 8.533333 10.24 20.906667 15.36 32.853334 15.36zM84.06656 314.112c6.4 3.84 14.08 5.973333 21.333333 5.973333 14.506667 0 29.013333-7.68 36.693334-21.333333a42.581333 42.581333 0 1 0-73.813334-42.666667 42.112 42.112 0 0 0 15.786667 58.026667z m0 395.52c-20.48 11.52-27.306667 37.973333-15.36 58.026667 7.68 13.653333 22.186667 21.333333 36.693333 21.333333a42.752 42.752 0 0 0 37.12-64 42.666667 42.666667 0 0 0-58.453333-15.36zM35.255893 472.448c2.56 0.426667 5.12 0.853333 7.68 0.853333 20.053333 0 37.973333-14.933333 41.813334-35.413333 3.84-23.466667-11.52-45.226667-34.56-49.493333-23.466667-3.84-45.653333 11.52-49.493334 34.56a42.453333 42.453333 0 0 0 34.56 49.493333z m49.493334 113.493333a42.666667 42.666667 0 1 0-84.053334 14.933334c3.84 20.48 21.76 34.986667 42.24 34.986666 2.133333 0 4.693333 0 7.253334-0.426666 23.04-3.84 38.869333-26.026667 34.56-49.493334zM657.933227 110.634667c5.12 1.706667 9.813333 2.56 14.506666 2.56 17.493333 0 34.133333-10.666667 40.106667-28.16a42.410667 42.410667 0 1 0-79.786667-29.013334 42.368 42.368 0 0 0 25.173334 54.613334z m155.733333 84.053333c12.373333 0 24.32-5.12 32.853333-14.933333a42.666667 42.666667 0 0 0-5.546666-60.16 42.666667 42.666667 0 0 0-60.16 5.12c-14.933333 17.92-12.373333 44.8 5.546666 60.16 7.68 6.826667 17.493333 9.813333 27.306667 9.813333zM512.013227 85.290667c11.093333 0 22.186667-4.736 30.293333-12.458667 7.68-8.149333 12.373333-19.328 12.373333-30.464 0-11.605333-4.693333-22.314667-12.373333-30.464a44.501333 44.501333 0 0 0-60.586667 0c-7.68 8.149333-12.373333 18.901333-12.373333 30.464 0 11.136 4.693333 22.314667 12.373333 30.464 8.106667 7.722667 19.2 12.458667 30.293334 12.458667z m-161.28 28.330666c5.12 0 10.24-0.853333 14.933333-2.986666 22.186667-8.106667 33.28-32.426667 25.173333-54.613334a42.368 42.368 0 0 0-54.613333-25.173333c-22.186667 7.68-33.28 32.426667-25.173333 54.613333a42.666667 42.666667 0 0 0 39.68 28.16z m588.544 323.84c3.413333 20.48 21.76 34.986667 41.813333 34.986667 2.56 0 5.12 0 7.68-0.426667a42.666667 42.666667 0 1 0-14.933333-84.053333c-23.04 3.84-38.826667 26.026667-34.56 49.493333zM512.013227 170.368a341.333333 341.333333 0 1 0 0 682.666667 341.333333 341.333333 0 0 0 0-682.666667z m0 597.333333c-141.141333 0-256-114.858667-256-256s114.858667-256 256-256 256 114.858667 256 256-114.858667 256-256 256z m406.613333-448c7.253333 0 14.506667-2.133333 21.333333-5.973333a42.24 42.24 0 0 0 15.36-58.026667 42.24 42.24 0 0 0-58.026666-15.786666 42.752 42.752 0 0 0 21.333333 79.786666z m-131.84 518.4c-18.346667 15.36-20.48 42.24-5.12 60.16 8.106667 10.24 20.48 15.36 32.426667 15.36 9.813333 0 19.626667-3.413333 27.306666-10.24 18.346667-14.933333 20.48-41.813333 5.546667-59.733333-15.36-18.346667-42.24-20.48-60.16-5.546667z m-609.28 5.546667a42.666667 42.666667 0 0 0 5.546667 60.16c7.68 6.826667 17.493333 10.24 27.306666 10.24 11.946667 0 24.32-5.12 32.853334-15.36 14.933333-17.92 12.373333-44.8-5.546667-60.16a43.093333 43.093333 0 0 0-60.16 5.12z m762.453333-134.4a42.026667 42.026667 0 0 0-58.026666 15.36 42.581333 42.581333 0 0 0 15.36 58.453333c6.826667 3.84 14.08 5.546667 21.333333 5.546667a42.496 42.496 0 0 0 21.333333-79.36z m48.810667-158.293333a42.666667 42.666667 0 0 0-49.493333 34.56c-3.84 23.04 11.52 45.226667 34.56 49.493333 2.56 0.426667 5.12 0.426667 7.68 0.426667 20.053333 0 38.4-14.506667 41.813333-34.986667a42.410667 42.410667 0 0 0-34.56-49.493333zM542.30656 950.528c-15.786667-15.872-44.8-15.872-60.586667 0-3.84 4.266667-6.826667 8.576-8.96 14.165333a43.093333 43.093333 0 0 0-3.413333 16.341334c0 5.589333 1.28 11.136 3.413333 16.298666s5.12 9.856 8.96 14.165334c8.106667 7.722667 19.2 12.458667 30.293334 12.458666 11.093333 0 22.186667-4.736 30.293333-12.458666 3.84-4.309333 6.826667-9.002667 8.96-14.165334 2.133333-5.162667 3.413333-10.752 3.413333-16.298666 0-5.589333-1.28-11.178667-3.413333-16.341334a44.416 44.416 0 0 0-8.96-14.165333z m-176.213333-37.76a42.410667 42.410667 0 0 0-54.613334 25.6 42.368 42.368 0 0 0 39.68 57.173333c17.493333 0 34.133333-10.666667 40.533334-28.16 7.68-22.229333-3.413333-46.506667-25.6-54.613333zM512.013227 341.034667a170.666667 170.666667 0 1 0 0 341.333333 170.666667 170.666667 0 0 0 0-341.333333z m0 213.333333a42.666667 42.666667 0 1 1 0-85.333333 42.666667 42.666667 0 0 1 0 85.333333z m146.346666 358.4c-22.186667 8.106667-33.28 32.426667-25.173333 54.613333 5.973333 17.066667 22.613333 28.16 40.106667 28.16 4.693333 0 9.813333-0.853333 14.506666-2.986666 22.186667-8.106667 33.28-32.426667 25.6-54.613334-8.106667-22.186667-32.853333-33.28-55.04-25.173333z"
                  fill="currentColor"
              />
            </svg>
          </Show>
          <Show when={!mediaBadge()}>
            <span
                class="absolute bottom-1 right-1 rounded-md bg-black/65 px-1.5 py-1 text-[0.65rem] font-medium tabular-nums leading-none text-white">
              {formatDuration(props.item.duration)}
            </span>
          </Show>
        </div>
        {/*中间是视频信息*/}
        <div class="flex min-w-0 flex-1 flex-col gap-3">
          <div class="min-w-0">
            <h3 class="line-clamp-2 text-sm font-semibold leading-5 text-base-content" title={props.item.title}>
              {props.item.title}
            </h3>
            <p class="mt-1 line-clamp-1 text-xs text-base-content/50">@{props.item.authorName}</p>
          </div>

          <div class="flex flex-wrap items-center gap-2 text-xs text-base-content/55">
            <span>发布 {props.item.publishTime ? formatDate(props.item.publishTime) : "-"}</span>
            <span
                class="rounded-full bg-base-200 px-2 py-0.5 tabular-nums">赞 {formatCount(props.item.diggCount)}</span>
            <span
                class="rounded-full bg-base-200 px-2 py-0.5 tabular-nums">藏 {formatCount(props.item.collectCount)}</span>
          </div>

          <Show when={!props.canDownload}>
            <p class="text-xs text-warning">没有可用下载地址，可能需要重新进入详情页刷新数据。</p>
          </Show>

          <Show when={isStandardVideo()}>
            <div class="flex items-center gap-3 text-xs">
              {/* 清晰度选项来自 bit_rate/play_addr；切换后会更新队列里的 videoURL，下载时使用当前选择。 */}
              <Show when={(props.item.videoOptions?.length ?? 0) > 0}>
                <span class="text-base-content/55 shrink-0">清晰度</span>
                <select
                    class="select select-bordered select-xs min-w-0 w-44"
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
              <span class="text-base-content/55 shrink-0">编码</span>
              <span class="truncate text-base-content/70">
              {selectedOption()?.codec ?? "-"}
                <Show when={selectedOption()?.bitRate}>
                {(bitRate) => ` · ${Math.round(bitRate() / 1000)} kbps`}
              </Show>
            </span>
            </div>
          </Show>

          <Show when={props.downloading && props.progress}>
            {(progress) => (
                <div class="flex items-center gap-2">
                  <progress class="progress progress-info h-2 flex-1" value={progress().percent} max="100"/>
                  <span class="w-24 text-right text-xs text-base-content/70">{progressText(progress())}</span>
                </div>
            )}
          </Show>
          <div class="flex flex-wrap items-center justify-end gap-2 border-t border-base-200 pt-2">
            <button
                class="btn btn-ghost btn-sm"
                type="button"
                onClick={props.onDownloadCover}
                disabled={props.coverDownloading || (props.item.coverCandidates?.length ?? 0) === 0}
            >
              {props.coverDownloading ? "保存中..." : "封面"}
            </button>
            <button class="btn btn-warning btn-sm" type="button"
                    onClick={() => removeDouyinVideo(props.item.awemeId)}
                    disabled={props.downloading}>
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
      const path = await DownloadCover(covers, item.title || item.awemeId || "cover");
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

      const detail = await VideoDetail(awemeId);
      const item = detailToDownloadItem(detail.aweme_detail);
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
              placeholder="请输入抖音视频分享链接、分享文案或视频 ID，可按回车直接解析"
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
