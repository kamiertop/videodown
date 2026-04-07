import type {JSXElement} from "solid-js";

/**
 * Bilibili UP 页面“头部容器”壳组件（通用 header 样式）。
 * 目前未在路由中使用，保留用于后续页面统一布局。
 */
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

