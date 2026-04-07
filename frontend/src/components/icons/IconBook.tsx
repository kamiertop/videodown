import type {JSXElement} from "solid-js";

/** 通用图标：书本/列表（20x20，fill）。 */
export default function IconBook(props: {class?: string}): JSXElement {
    return (
        <svg
            xmlns="http://www.w3.org/2000/svg"
            class={props.class ?? "h-4 w-4"}
            viewBox="0 0 20 20"
            fill="currentColor"
        >
            <path d="M4 3a2 2 0 012-2h8a2 2 0 012 2v12a2 2 0 01-2 2H6a2 2 0 01-2-2V3z"/>
        </svg>
    );
}

