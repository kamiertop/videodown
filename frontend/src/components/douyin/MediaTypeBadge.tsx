import type {JSXElement} from "solid-js";
import type {DouyinMediaBadge} from "../../lib/douyin/media.ts";
import IconLivePhoto from "../icons/IconLivePhoto.tsx";

function ImageBadgeIcon(): JSXElement {
  return (
      <svg
          class="h-3.5 w-3.5"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 12 12"
          aria-hidden="true"
      >
        <path
            d="M1.455 0C.65 0 0 .651 0 1.455V8c0 .803.651 1.455 1.455 1.455H8c.803 0 1.455-.652 1.455-1.455V1.455C9.455.65 8.803 0 8 0H1.455z"
            fill="currentColor"
        />
        <path
            d="M4 12a1.455 1.455 0 0 1-1.455-1.454h5.819a2.182 2.182 0 0 0 2.181-2.182V2.545C11.35 2.545 12 3.197 12 4v5.09A2.909 2.909 0 0 1 9.09 12H4z"
            fill="currentColor"
        />
      </svg>
  );
}

export default function MediaTypeBadge(props: { type: DouyinMediaBadge }): JSXElement {
  const label = () => props.type === "live-photo" ? "动图" : "图文";

  return (
      <span
          class="absolute right-1.5 top-1.5 z-10 flex h-6 w-6 items-center justify-center rounded-md bg-black/55 text-white shadow-sm ring-1 ring-white/20 backdrop-blur-sm"
          aria-label={label()}
          title={label()}
      >
        {props.type === "live-photo" ? <IconLivePhoto class="h-4 w-4"/> : <ImageBadgeIcon/>}
      </span>
  );
}
