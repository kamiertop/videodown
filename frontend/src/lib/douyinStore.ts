import {createSignal} from "solid-js";
import type {model} from "../../wailsjs/go/models";

export interface DouyinVideoOption {
  id: string;
  label: string;
  gearName: string;
  dataSize: number;
  bitRate?: number;
  codec: string;
  url: string;
}

export interface DouyinDownloadAsset {
  url: string;
  kind: "image" | "video";
  ext: string;
}

// 下载页只保存轻量任务数据；完整 AwemeItem 太大，也不适合跨页面长期持有。
export interface DouyinDownloadItem {
  awemeId: string;
  sourceKind: string;
  sourceName?: string;
  title: string;
  cover: string;
  coverCandidates?: model.Cover[];
  duration: number;
  authorName: string;
  publishTime?: number;
  diggCount?: number;
  collectCount?: number;
  link?: string;
  videoURL?: string;
  videoOptions?: DouyinVideoOption[];
  selectedVideoOptionId?: string;
  imageURLs?: string[];
  assets?: DouyinDownloadAsset[];
  musicURL?: string;
  mediaBadge?: "image" | "live-photo";
}

const [douyinVideoList, setDouyinVideoList] = createSignal<DouyinDownloadItem[]>([]);

export {douyinVideoList};

function itemIdentity(item: DouyinDownloadItem): string {
  return item.awemeId.trim();
}

export function addDouyinVideos(items: DouyinDownloadItem[]): void {
  setDouyinVideoList((prev) => {
    const existing = new Set(prev.map(itemIdentity));
    const next = items.filter((item) => {
      const key = itemIdentity(item);
      if (!key || existing.has(key)) return false;
      existing.add(key);
      return true;
    });
    return [...prev, ...next];
  });
}

export function removeDouyinVideo(awemeId: string): void {
  const key = awemeId.trim();
  if (!key) return;
  setDouyinVideoList((prev) => prev.filter((item) => item.awemeId.trim() !== key));
}

export function updateDouyinVideoOption(awemeId: string, optionId: string): void {
  const key = awemeId.trim();
  if (!key) return;
  setDouyinVideoList((prev) => prev.map((item) => {
    if (item.awemeId.trim() !== key) return item;
    const option = item.videoOptions?.find((entry) => entry.id === optionId);
    if (!option) return item;
    // 清晰度选择本质上就是替换本任务最终提交给后端的 videoURL。
    return {
      ...item,
      selectedVideoOptionId: option.id,
      videoURL: option.url,
    };
  }));
}

export function clearDouyinVideos(): void {
  setDouyinVideoList([]);
}
