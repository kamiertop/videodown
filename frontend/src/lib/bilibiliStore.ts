import type {MediaCardItem} from "./model.ts";
import {createSignal} from "solid-js";

const [videoList, setVideoList] = createSignal<MediaCardItem[]>([])

export {videoList};

function mediaIdentity(item: MediaCardItem): string {
    const bvid = item.bvid?.trim().toUpperCase();
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

export function clearVideos(): void {
    setVideoList([]);
}
