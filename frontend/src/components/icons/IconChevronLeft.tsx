import type {JSXElement} from "solid-js";

/** 通用图标：返回/向左箭头（24x24，stroke）。 */
export default function IconChevronLeft(props: {class?: string}): JSXElement {
    return (
        <svg
            xmlns="http://www.w3.org/2000/svg"
            class={props.class ?? "h-4 w-4"}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
        >
            <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7"/>
        </svg>
    );
}

