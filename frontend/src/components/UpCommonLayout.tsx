import type {JSXElement} from "solid-js";

export default function UpCommonLayout(props: {
    headerLeft: JSXElement;
    headerRight?: JSXElement;
    headerBelow?: JSXElement;
    children: JSXElement;
}): JSXElement {
    return (
        <section class="flex h-full min-h-0 flex-col gap-3 overflow-hidden bg-base-200/40 p-3">
            <header
                class="flex h-12 shrink-0 items-center gap-3 overflow-hidden rounded-xl border border-base-300 bg-base-100 px-4"
            >
                {/* 左侧永远可见；右侧超出则截断 */}
                <div class="flex min-w-0 flex-1 items-center gap-2">
                    {props.headerLeft}
                </div>
                {props.headerRight ? (
                    <div class="flex shrink-0 items-center gap-2">
                        {props.headerRight}
                    </div>
                ) : null}
            </header>

            {props.headerBelow ? (
                <div class="shrink-0">
                    {props.headerBelow}
                </div>
            ) : null}

            <main class="min-h-0 flex-1 overflow-auto rounded-xl border border-base-300 bg-base-100">
                {props.children}
            </main>
        </section>
    );
}

