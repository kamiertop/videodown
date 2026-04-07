import type {JSXElement} from "solid-js";

/** 通用图标：刷新（24x24，stroke），支持外部传入 class 控制大小/旋转等。 */
export default function IconRefresh(props: {class?: string}): JSXElement {
    return (
        <svg
            xmlns="http://www.w3.org/2000/svg"
            class={props.class ?? "h-4 w-4"}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
        >
            <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
            />
        </svg>
    );
}

