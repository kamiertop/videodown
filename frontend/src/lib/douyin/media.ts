import {model} from "../../../wailsjs/go/models";
import type {DouyinDownloadAsset, DouyinVideoOption} from "./store.ts";

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
      label: `${gearName} · ${formatDataSize(dataSize)} · ${codec}`,
      gearName,
      dataSize,
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
      label: `${fallback.name} · ${formatDataSize(dataSize)} · ${fallback.codec}`,
      gearName: fallback.name,
      dataSize,
      codec: fallback.codec,
      url,
    });
  }

  return options;
}

export function defaultDouyinVideoOption(options: DouyinVideoOption[]): DouyinVideoOption | undefined {
  // 默认优先选兼容性最好的 H.264 高码率项；如果没有 H.264，再退到接口返回的第一个可用项。
  const h264 = options
      .filter((option) => option.codec === "H.264")
      .sort((a, b) => (b.bitRate ?? 0) - (a.bitRate ?? 0) || b.dataSize - a.dataSize)[0];
  return h264 ?? options[0];
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

export function douyinImageURLs(item: model.AwemeItem): string[] {
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

export function douyinMusicURL(item: model.AwemeItem): string {
  return item.music?.play_url?.url_list?.[0] ?? "";
}

export function douyinDownloadAssets(item: model.AwemeItem): DouyinDownloadAsset[] {
  return (item.images ?? [])
      .map((image): DouyinDownloadAsset | undefined => {
        if (isLivePhotoImage(image)) {
          const url = imageVideoURL(image);
          return url ? {url, kind: "video", ext: ".mp4"} : undefined;
        }
        const url = image.url_list?.[0] ?? "";
        return url ? {url, kind: "image", ext: ".jpg"} : undefined;
      })
      .filter((asset): asset is DouyinDownloadAsset => asset != null);
}

function hasDouyinImages(item: model.AwemeItem): boolean {
  return item.images != null;
}

function isLivePhotoImage(image: model.ImageItem): boolean {
  // image.video 是 ImageVideo 值类型（Go 非指针），即使静态图片也会序列化为非 null 对象，
  // 因此必须检查 video 是否有实质内容（play_addr 有 URL）而非判断 null。
  return image.live_photo_type === 1
      || image.clip_type === 5
      || (image.video?.play_addr?.url_list?.length ?? 0) > 0;
}

function hasStaticImage(item: model.AwemeItem): boolean {
  return (item.images ?? []).some((image) => !isLivePhotoImage(image));
}

function hasOnlyLivePhotoImages(item: model.AwemeItem): boolean {
  const images = item.images ?? [];
  return images.length > 0 && images.every(isLivePhotoImage);
}

// 是否是图片合集；只要存在静态照片，即使混有动图，也按图文处理。
export function isDouyinImageAlbum(item: model.AwemeItem): boolean {
  return hasDouyinImages(item) && hasStaticImage(item);
}

// 是否是动图
export function isDouyinLivePhoto(item: model.AwemeItem): boolean {
  return item.is_live_photo === 1
      || hasOnlyLivePhotoImages(item)
      || (!hasStaticImage(item) && (item.media_type === 42 || item.is_slides));
}
