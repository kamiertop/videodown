import {createFileRoute} from '@tanstack/solid-router'
import {createSignal, For, type JSXElement, Show} from "solid-js";
import {ParseVideo, VideoDetail} from "../../../wailsjs/go/api/Douyin";
import {model} from "../../../wailsjs/go/models";
import EmptyState from "../../components/EmptyState.tsx";
import Toast from "../../components/Toast.tsx";
import {useToast} from "../../hooks/useToast.ts";
import {
  type DouyinDownloadProgress,
  formatDouyinSource,
  useDouyinDownloadQueue
} from "../../lib/douyinDownloadQueue.ts";
import {defaultDouyinVideoOption, douyinImageURLs, douyinVideoOptions, formatDataSize,} from "../../lib/douyinMedia.ts";
import {
  addDouyinVideos,
  type DouyinDownloadItem,
  douyinVideoList,
  removeDouyinVideo,
  updateDouyinVideoOption,
} from "../../lib/douyinStore.ts";
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

  return {
    awemeId,
    sourceKind: "解析结果",
    sourceName: "手动解析",
    title,
    cover,
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
  };
}

function progressText(progress: DouyinDownloadProgress | undefined): string {
  // 后端把视频和图片合集都归一成同一条 0-100 进度，前端只区分阶段文案。
  if (!progress) return "";
  if (progress.phase === "video") return `视频下载 ${Math.round(progress.percent)}%`;
  if (progress.phase === "image") return `图片下载 ${Math.round(progress.percent)}%`;
  if (progress.phase === "done") return "完成";
  return "下载失败";
}

function DouyinDownloadCard(props: {
  item: DouyinDownloadItem;
  canDownload: boolean;
  downloading: boolean;
  progress: DouyinDownloadProgress | undefined;
  onDownload: () => void;
}): JSXElement {
  const isImageAlbum = () => (props.item.imageURLs?.length ?? 0) > 0;
  const selectedOption = () => props.item.videoOptions?.find((option) => option.id === props.item.selectedVideoOptionId);

  return (
    <article class="flex flex-row gap-3 rounded-lg border border-base-300 bg-base-100 p-1">
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
        <Show
          when={isImageAlbum()}
          fallback={
            <span
              class="absolute bottom-1 right-1 rounded-md bg-black/65 px-1.5 py-1 text-[0.65rem] font-medium tabular-nums leading-none text-white">
              {formatDuration(props.item.duration)}
            </span>
          }
        >
          <span
            class="absolute bottom-1 right-1 rounded-md bg-black/65 px-1.5 py-1 text-[0.65rem] font-medium leading-none text-white">
            图集 {props.item.imageURLs?.length ?? 0}
          </span>
        </Show>
      </div>
      {/*中间是视频信息*/}
      <div class="flex min-w-0 flex-1 flex-col gap-2">
        <div class="min-w-0">
          <h3 class="line-clamp-2 text-sm font-semibold leading-5 text-base-content" title={props.item.title}>
            {props.item.title}
          </h3>
          <p class="mt-1 line-clamp-1 text-xs text-base-content/50">@{props.item.authorName}</p>
        </div>

        <div class="flex flex-wrap items-center gap-2 text-xs text-base-content/55">
          <span
            class="rounded-full bg-primary/10 px-2 py-0.5 font-medium text-primary">{formatDouyinSource(props.item)}</span>
          <span>发布 {props.item.publishTime ? formatDate(props.item.publishTime) : "-"}</span>
          <span class="rounded-full bg-base-200 px-2 py-0.5 tabular-nums">赞 {formatCount(props.item.diggCount)}</span>
          <span
            class="rounded-full bg-base-200 px-2 py-0.5 tabular-nums">藏 {formatCount(props.item.collectCount)}</span>
        </div>

        <Show when={!props.canDownload}>
          <p class="text-xs text-warning">没有可用下载地址，可能需要重新进入来源页面刷新数据。</p>
        </Show>

        <Show when={!isImageAlbum() && (props.item.videoOptions?.length ?? 0) > 0}>
          {/* 清晰度选项来自 bit_rate/play_addr；切换后会更新队列里的 videoURL，下载时使用当前选择。 */}
          <div class="flex items-center gap-2 text-xs">
            <span class="text-base-content/55">清晰度</span>
            <select
              class="select select-bordered select-xs min-w-0 w-56"
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

          </div>
        </Show>
        <div class="flex flex-row gap-2 text-xs">
          <span class="text-base-content/55">编码</span>
          <span class="truncate text-base-content/70">
              {selectedOption()?.codec ?? "-"}
            <Show when={selectedOption()?.bitRate}>
                {(bitRate) => ` · ${Math.round(bitRate() / 1000)} kbps`}
              </Show>
            </span>
        </div>

        <Show when={props.downloading && props.progress}>
          {(progress) => (
            <div class="flex items-center gap-2">
              <progress class="progress progress-info h-2 flex-1" value={progress().percent} max="100"/>
              <span class="w-24 text-right text-xs text-base-content/70">{progressText(progress())}</span>
            </div>
          )}
        </Show>


      </div>
      <div class="flex items-center gap-2">
        <div class="flex items-center gap-2 flex-col">
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
  const {message, type, showToast} = useToast();
  // 下载状态集中在 hook 中，页面只负责渲染列表和把用户操作转发给队列。
  const queue = useDouyinDownloadQueue(showToast);

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
            <p class="truncate text-sm text-base-content/80">{queue.listSourceSummary()}</p>
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
