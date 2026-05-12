import {For, type JSXElement} from "solid-js";
import type {MediaCardItem} from "../../lib/model.ts";
import VideoCard from "./VideoCard";

/** 静态视频卡片网格 — 少量列表用，大量列表请用 VirtualVideoGrid */
export default function VideoCardGrid(props: {
  medias: MediaCardItem[];
  selectedSet: () => Set<number>;
  onToggleSelect: (id: number) => void;
  onDownloadOne: (media: MediaCardItem) => void;
}): JSXElement {
  return (
    <div class="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6">
      <For each={props.medias}>
        {(media) => (
          <VideoCard
            media={media}
            isSelected={props.selectedSet().has(media.id)}
            onToggleSelect={props.onToggleSelect}
            onDownloadOne={props.onDownloadOne}
          />
        )}
      </For>
    </div>
  );
}
