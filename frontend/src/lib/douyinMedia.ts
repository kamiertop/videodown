import {model} from "../../wailsjs/go/models";
import type {DouyinVideoOption} from "./douyinStore.ts";

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
  const fallbacks: Array<{name: string; codec: string; playAddr?: model.PlayInfo}> = [
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

export function douyinImageURLs(item: model.AwemeItem): string[] {
  // 图片合集的无水印地址在 url_list，download_url_list 通常是带水印兜底。
  return (item.images ?? [])
    .map((image) => image.url_list?.[0] ?? image.download_url_list?.[0] ?? "")
    .filter((url) => url.length > 0);
}

export function isDouyinImageAlbum(item: model.AwemeItem): boolean {
  // 抖音图集通常表现为 media_type=2 且 aweme_type=68，视频字段里可能只是配乐。
  return item.media_type === 2 && douyinImageURLs(item).length > 0 && item.aweme_type === 68;
}
