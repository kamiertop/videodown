import * as model from "@bindings/github.com/kamiertop/videodown/douyin/model/models";
import type {DouyinDownloadAsset, DouyinVideoOption} from "./store.ts";

/**
 * 抖音作品在前端展示和下载流程中使用的统一媒体类型。
 *
 * 这里按抖音手机端的展示语义归类：
 * - 单个动图素材：`live-photo`
 * - 只要存在多个素材，或静态图与动图混合：`image`
 * - 没有图片素材且没有图文兜底信号：`video`
 */
export type DouyinMediaKind = "video" | "image" | "live-photo";

/**
 * 卡片右上角只展示非普通视频的类型标识。
 */
export type DouyinMediaBadge = Exclude<DouyinMediaKind, "video">;

type DouyinMediaItem = {
  images?: model.ImageItem[] | null;
  is_live_photo?: number;
  is_slides?: boolean;
  media_type?: number;
  music?: model.Music;
};

function firstURL(playAddr: model.PlayInfo | undefined): string {
  return playAddr?.url_list?.[0] ?? "";
}

function optionID(prefix: string, url: string, index: number): string {
  return `${prefix}-${index}-${url.slice(0, 32)}`;
}

function codecLabel(item: model.BitRateItem): string {
  if (item.is_h265 === 1) return "H.265";
  if (item.is_bytevc1 === 1) return "ByteVC1";
  return "H.264";
}

function videoOptionLabel(name: string, width: number | undefined, height: number | undefined, fps: number | undefined, dataSize: number, codec: string): string {
  const resolution = width && height ? `${width}×${height}` : "未知分辨率";
  const frameRate = fps ? ` · ${fps}fps` : "";
  return `${name} · ${resolution}${frameRate} · ${formatDataSize(dataSize)} · ${codec}`;
}

// 抖音接口里的 data_size 是字节数；展示给用户时统一压成易读单位。
export function formatDataSize(value: number | undefined): string {
  const size = Number(value) || 0;
  if (size <= 0) return "未知大小";
  const units = ["B", "KB", "MB", "GB"];
  let n = size;
  let index = 0;
  while (n >= 1024 && index < units.length - 1) {
    n /= 1024;
    index += 1;
  }
  return `${n >= 10 || index === 0 ? n.toFixed(0) : n.toFixed(1)} ${units[index]}`;
}

export function douyinVideoOptions(item: model.AwemeItem): DouyinVideoOption[] {
  const seen = new Set<string>();
  const options: DouyinVideoOption[] = [];

  // bit_rate 是抖音最细的清晰度列表。每一项一般都有 gear_name、码率、编码信息和独立 play_addr。
  for (const [index, entry] of (item.video?.bit_rate ?? []).entries()) {
    const url = firstURL(entry.play_addr);
    if (!url || seen.has(url)) continue;
    seen.add(url);
    const gearName = entry.gear_name || `bit_rate_${index + 1}`;
    const codec = codecLabel(entry);
    const dataSize = Number(entry.play_addr?.data_size ?? 0);
    options.push({
      id: optionID("bitrate", url, index),
      label: videoOptionLabel(gearName, entry.play_addr?.width, entry.play_addr?.height, entry.FPS, dataSize, codec),
      gearName,
      dataSize,
      width: entry.play_addr?.width,
      height: entry.play_addr?.height,
      fps: entry.FPS,
      bitRate: entry.bit_rate,
      codec,
      url,
    });
  }

  // 有些响应没有 bit_rate，或者 bit_rate 中缺少可用地址；保留顶层 play_addr 作为兜底下载选项。
  const fallbacks: Array<{ name: string; codec: string; playAddr?: model.PlayInfo }> = [
    {name: "play_addr_h264", codec: "H.264", playAddr: item.video?.play_addr_h264},
    {name: "play_addr_265", codec: "H.265", playAddr: item.video?.play_addr_265},
    {name: "play_addr", codec: "默认", playAddr: item.video?.play_addr},
  ];

  for (const [index, fallback] of fallbacks.entries()) {
    const url = firstURL(fallback.playAddr);
    if (!url || seen.has(url)) continue;
    seen.add(url);
    const dataSize = Number(fallback.playAddr?.data_size ?? 0);
    options.push({
      id: optionID(fallback.name, url, index),
      label: videoOptionLabel(fallback.name, fallback.playAddr?.width, fallback.playAddr?.height, undefined, dataSize, fallback.codec),
      gearName: fallback.name,
      dataSize,
      width: fallback.playAddr?.width,
      height: fallback.playAddr?.height,
      codec: fallback.codec,
      url,
    });
  }

  return options;
}

export function defaultDouyinVideoOption(options: DouyinVideoOption[]): DouyinVideoOption | undefined {
  // 默认在所有编码格式中选择最高画质。
  return [...options]
      .sort((a, b) => ((b.width ?? 0) * (b.height ?? 0) - (a.width ?? 0) * (a.height ?? 0)) || (b.fps ?? 0) - (a.fps ?? 0) || (b.bitRate ?? 0) - (a.bitRate ?? 0) || b.dataSize - a.dataSize)[0];
}

export function douyinVideoURL(item: model.AwemeItem): string {
  return defaultDouyinVideoOption(douyinVideoOptions(item))?.url ?? "";
}

// TODO: 简化逻辑，直接返回一个字符串
export function douyinCoverCandidates(item: model.AwemeItem): model.Cover[] {
  const covers = [
    item.video?.dynamic_cover,
    item.video?.raw_cover,
    item.video?.cover,
    item.video?.origin_cover,
  ].filter((cover): cover is model.Cover => !!cover && (cover.url_list?.length ?? 0) > 0);

  const seen = new Set<string>();
  return covers.filter((cover) => {
    const key = cover.uri || cover.url_list?.[0];
    if (!key || seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}

export function douyinImageURLs(item: DouyinMediaItem): string[] {
  // 图片合集的无水印地址在 url_list
  return (item.images ?? [])
      .map((image) => image.url_list?.[0] ?? "")
      .filter((url) => url.length > 0);
}

function imageVideoURL(image: model.ImageItem): string {
  return image.video?.play_addr_h264?.url_list?.[0]
      ?? image.video?.play_addr?.url_list?.[0]
      ?? image.video?.bit_rate?.[0]?.play_addr?.url_list?.[0]
      ?? "";
}

/** 从图片地址路径识别原始扩展名；查询参数不会影响判断。 */
function imageExtFromURL(rawURL: string): string {
  try {
    const path = new URL(rawURL).pathname;
    const name = path.slice(path.lastIndexOf("/") + 1).toLowerCase();
    const dot = name.lastIndexOf(".");
    const ext = dot >= 0 ? name.slice(dot) : "";
    if ([".jpg", ".jpeg", ".png", ".webp", ".gif"].includes(ext)) return ext === ".jpeg" ? ".jpg" : ext;
  } catch {
    // URL 不可解析时由后端使用默认扩展名兜底。
  }
  return ".jpg";
}

export function douyinMusicURL(item: DouyinMediaItem): string {
  return item.music?.play_url?.url_list?.[0] ?? "";
}

export function douyinDownloadAssets(item: DouyinMediaItem): DouyinDownloadAsset[] {
  return (item.images ?? [])
      .map((image): DouyinDownloadAsset | undefined => {
        if (isLivePhotoImage(image)) {
          const url = imageVideoURL(image);
          return url ? {url, kind: "video", ext: ".mp4"} : undefined;
        }
        const url = image.url_list?.[0] ?? "";
        return url ? {url, kind: "image", ext: imageExtFromURL(url)} : undefined;
      })
      .filter((asset): asset is DouyinDownloadAsset => asset != null);
}

/**
 * 判断单个图片素材是否是动图素材。
 *
 * 这里只使用接口明确给出的动图字段；不再用 `image.video.play_addr` 推断展示类型，
 * 避免普通图片因为带有兜底 video 结构而被误判成动图。
 */
function isLivePhotoImage(image: model.ImageItem): boolean {
  return image.live_photo_type === 1 || image.clip_type === 5;
}

/**
 * 只有一个素材且这个素材本身是动图时，手机端才展示为“动图”。
 */
function isSingleLivePhoto(item: DouyinMediaItem): boolean {
  const images = item.images ?? [];
  return images.length === 1 && isLivePhotoImage(images[0]);
}

/**
 * 返回抖音作品的唯一媒体类型判断结果。
 *
 * 所有页面都应通过这个函数判断普通视频、图文、动图，不要在页面组件里直接组合
 * `media_type`、`images`、`is_live_photo`、`is_slides` 等字段，避免不同入口判断不一致。
 */
export function douyinMediaKind(item: DouyinMediaItem): DouyinMediaKind {
  const images = item.images ?? [];
  if (images.length > 0) return isSingleLivePhoto(item) ? "live-photo" : "image";
  if (item.media_type === 2 || item.is_slides === true) return "image";
  return "video";
}

/**
 * 将媒体类型转换成卡片需要的右上角标识；普通视频不展示类型标识。
 */
export function douyinMediaBadge(item: DouyinMediaItem): DouyinMediaBadge | undefined {
  const kind = douyinMediaKind(item);
  return kind === "video" ? undefined : kind;
}

/**
 * 兼容旧调用名：是否按图文处理。
 */
export function isDouyinImageAlbum(item: DouyinMediaItem): boolean {
  return douyinMediaKind(item) === "image";
}

/**
 * 兼容旧调用名：是否按单个动图处理。
 */
export function isDouyinLivePhoto(item: DouyinMediaItem): boolean {
  return douyinMediaKind(item) === "live-photo";
}
