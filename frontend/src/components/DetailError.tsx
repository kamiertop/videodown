import type {JSXElement} from "solid-js";

export default function DetailError(props: { message: string; onRetry: () => void }): JSXElement {
    return (
        <div class="flex h-full items-center justify-center p-6">
            <div class="rounded-xl border border-error/30 bg-error/10 px-5 py-4 text-center">
                <p class="text-sm font-semibold text-error">加载详情失败</p>
                <p class="mt-1 max-w-md text-xs text-error/80">{props.message}</p>
                <button class="btn btn-error btn-sm mt-3" onClick={props.onRetry}>重试</button>
            </div>
        </div>
    );
}
