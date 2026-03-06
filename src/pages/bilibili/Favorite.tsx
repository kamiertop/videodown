import {For, JSXElement} from "solid-js";

const tips = ["输入收藏夹链接", "解析收藏夹视频列表", "支持分批下载和失败重试"];

export default function BilibiliFavorite(): JSXElement {
    return (
        <div class="rounded-3xl border border-neutral-200 bg-neutral-50 p-6 space-y-4">
            <h2 class="text-xl font-semibold">收藏夹视频下载</h2>
            <p class="text-sm text-neutral-500">这里可以接入收藏夹链接解析和批量下载。</p>
            <div class="grid gap-3 md:grid-cols-3">
                <For each={tips}>
                    {(tip) => (
                        <div class="rounded-2xl border border-neutral-200 bg-white p-4 text-sm text-neutral-600">
                            {tip}
                        </div>
                    )}
                </For>
            </div>
        </div>
    );
}

