import type {JSXElement} from "solid-js";

export default function IconPlayCircle(props: { class?: string }): JSXElement {
  return (
    <svg
      class={props.class ?? "h-4 w-4"}
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <path
        d="M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"
        stroke="currentColor"
        stroke-width="1.8"
      />
      <path
        d="M10 8.75v6.5L15.25 12 10 8.75Z"
        fill="currentColor"
      />
    </svg>
  );
}
