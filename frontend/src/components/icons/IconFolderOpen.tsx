import type {JSXElement} from "solid-js";

export default function IconFolderOpen(props: { class?: string }): JSXElement {
  return (
    <svg
      class={props.class ?? "h-4 w-4"}
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <path
        d="M3.5 6.5A2.5 2.5 0 0 1 6 4h3.2c.7 0 1.36.33 1.78.89l.83 1.11H18a2.5 2.5 0 0 1 2.5 2.5v1"
        stroke="currentColor"
        stroke-width="1.8"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
      <path
        d="M3.44 10.9A2 2 0 0 1 5.4 9.3h14.1a1.5 1.5 0 0 1 1.45 1.88l-1.65 6.3A2 2 0 0 1 17.36 19H5.75a2 2 0 0 1-1.95-2.45l1.15-4.98"
        stroke="currentColor"
        stroke-width="1.8"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
    </svg>
  );
}
