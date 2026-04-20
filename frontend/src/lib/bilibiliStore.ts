import {createSignal} from "solid-js";
import type {MediaCardItem} from "./model.ts";

const [videoList, setVideoList] = createSignal<MediaCardItem[]>([])

export {videoList};

function mediaIdentity(item: MediaCardItem): string {
  const bvid = item.bvid?.trim().toUpperCase();
  const cid = item.cid;
  if (bvid && cid != null && cid > 0) return `bvid:${bvid}:cid:${cid}`;
  if (bvid) return `bvid:${bvid}`;
  return `id:${item.id}`;
}

export function addVideos(selectedVideos: MediaCardItem[]): void {
  // 同一个视频可能从 UP 投稿、合集、手动解析等不同入口进入列表；下载时必须按 BV 去重。
  setVideoList(prev => {
    const existing = new Set(prev.map(mediaIdentity));
    const newVideos = selectedVideos.filter(item => {
      const key = mediaIdentity(item);
      if (existing.has(key)) return false;
      existing.add(key);
      return true;
    });
    return [...prev, ...newVideos];
  });
}

export function removeVideo(id: number): void {
  setVideoList(prev => prev.filter(item => item.id !== id));
}

export function removeVideosByBvid(bvid: string): void {
  const key = bvid.trim().toUpperCase();
  if (!key) return;
  setVideoList(prev => prev.filter(item => item.bvid?.trim().toUpperCase() !== key));
}

/** 下载成功回调：多 P 只移除对应 cid 的一行，单 P 仍按 BV 移除无 cid 的条目。 */
export function removeVideoAfterDownloadSuccess(bvid: string, cid?: number): void {
  const bv = bvid.trim().toUpperCase();
  if (!bv) return;
  const part = cid != null && cid > 0;
  setVideoList((prev) =>
    prev.filter((item) => {
      if (item.bvid?.trim().toUpperCase() !== bv) return true;
      if (part) return Number(item.cid) !== cid;
      const ic = item.cid;

      return ic != null && ic > 0;
    }),
  );
}

export function clearVideos(): void {
  setVideoList([]);
}
