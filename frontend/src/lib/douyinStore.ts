import {createSignal} from "solid-js";

export interface DouyinDownloadItem {
  awemeId: string;
  title: string;
  cover: string;
  duration: number;
  authorName: string;
  publishTime?: number;
  diggCount?: number;
  collectCount?: number;
  link?: string;
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

export function clearDouyinVideos(): void {
  setDouyinVideoList([]);
}
