import {type JSXElement, Show} from "solid-js";
import type {DouyinDownloadItem} from "../../lib/douyinStore.ts";
import NoCover from "../NoCover.tsx";

export interface DouyinVideoCardItem {
  id: string;
  cover: string;
  title: string;
  author: string;
  publishText: string;
  durationText: string;
  downloadItem: DouyinDownloadItem;
  showImgLabel: boolean;
}

export default function DouyinVideoCard(props: {
  item: DouyinVideoCardItem;
  onClick: () => void;
  selectedClass: string;
}): JSXElement {

  return (
    <article
      class={`flex cursor-pointer flex-col overflow-hidden rounded-lg transition-colors ${props.selectedClass}`}
      style={{"content-visibility": "auto", "contain-intrinsic-size": "220px"}}
      onClick={props.onClick}
    >
      <div class="relative aspect-3/5 w-full overflow-hidden bg-base-200">
        <Show when={props.item.cover} fallback={<NoCover/>}>
          <img
            src={props.item.cover}
            class="h-full w-full object-cover"
            alt={props.item.title}
            loading="lazy"
            decoding="async"
            referrerPolicy="no-referrer"
          />
        </Show>
        {/*图片合集时，显示右上角的标识*/}
        <Show when={props.item.showImgLabel}>
          <svg
            width="12"
            height="12"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
            class="absolute right-1.5 top-1.5 z-10 h-3.5 w-3.5 drop-shadow-[0_1px_2px_rgba(0,0,0,0.65)]"
            viewBox="0 0 12 12"
            aria-label="图文"
          >
            <path
              d="M1.455 0C.65 0 0 .651 0 1.455V8c0 .803.651 1.455 1.455 1.455H8c.803 0 1.455-.652 1.455-1.455V1.455C9.455.65 8.803 0 8 0H1.455z"
              fill="#fff"></path>
            <path
              d="M4 12a1.455 1.455 0 0 1-1.455-1.454h5.819a2.182 2.182 0 0 0 2.181-2.182V2.545C11.35 2.545 12 3.197 12 4v5.09A2.909 2.909 0 0 1 9.09 12H4z"
              fill="#fff"></path>
          </svg>
        </Show>
        {/*图片合集时，不显示视频时长*/}
        <Show when={!props.item.showImgLabel}>
          <span
            class="absolute bottom-1 right-1 rounded-md bg-black/65 px-1.5 py-1 text-[0.65rem] font-medium leading-none text-white">
            {props.item.durationText}
          </span>
        </Show>
      </div>

      <div class="flex flex-col p-2">
        <h3 class="line-clamp-2 text-[12px] font-semibold leading-4.5 text-base-content">
          {props.item.title}
        </h3>
        <p class="mt-1 line-clamp-1 text-[10px] text-base-content/50">
          @{props.item.author}
        </p>
        <p class="mt-1.5 text-[10px] text-base-content/45">
          发布 {props.item.publishText}
        </p>
      </div>
    </article>
  );
}
