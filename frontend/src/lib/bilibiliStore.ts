import type {MediaCardItem} from "./model.ts";
import {createSignal} from "solid-js";

const [videoList, setVideoList] = createSignal<MediaCardItem[]>([])

export {videoList};

export function addVideos(selectedVideos: MediaCardItem[]): void {
    // 去重并合并
    setVideoList(prev => {
        const existingIds = new Set(prev.map(item => item.id));
        const newVideos = selectedVideos.filter(item => !existingIds.has(item.id));
        return [...prev, ...newVideos];
    });
}

export function removeVideo(id: number): void {
    setVideoList(prev => prev.filter(item => item.id !== id));
}

export function clearVideos(): void {
    setVideoList([]);
}
