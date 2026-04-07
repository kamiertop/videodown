import type {JSXElement} from "solid-js";

/** 通用图标：用户/群组（24x24，stroke）。 */
export default function IconUsers(props: {class?: string}): JSXElement {
    return (
        <svg
            xmlns="http://www.w3.org/2000/svg"
            class={props.class ?? "h-4 w-4"}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="1"
        >
            <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"
            />
        </svg>
    );
}

