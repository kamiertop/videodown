import type {JSXElement} from "solid-js";

export default function UpHeaderShell(props: {
    children: JSXElement;
}): JSXElement {
    return (
        <header
            class="flex h-12 shrink-0 items-center gap-3 overflow-hidden rounded-xl border border-base-300 bg-base-100 px-4"
        >
            {props.children}
        </header>
    );
}

