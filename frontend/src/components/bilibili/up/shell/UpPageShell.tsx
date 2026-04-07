import type {JSXElement} from "solid-js";

/**
 * Bilibili UP 页面“页面壳”组件（header + 可选 headerBelow + 内容区）。
 * 目前未在路由中使用，保留用于后续拆分更多 UP 子页面。
 */
export default function UpPageShell(props: {
    header: JSXElement;
    headerBelow?: JSXElement;
    children: JSXElement;
}): JSXElement {
    return (
        <section class="flex h-full min-h-0 flex-col gap-3 overflow-hidden bg-base-200/40 p-3">
            <div class="shrink-0">
                {props.header}
            </div>
            {props.headerBelow ? (
                <div class="shrink-0">
                    {props.headerBelow}
                </div>
            ) : null}
            {props.children}
        </section>
    );
}

