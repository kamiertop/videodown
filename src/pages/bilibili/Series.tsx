import {For, JSXElement} from "solid-js";

const tips = ["支持合集、系列入口链接", "展示分集列表", "支持整组下载或单集选择"];

export default function BilibiliSeries(): JSXElement {
    return (
        <div class="rounded-3xl border border-neutral-200 bg-neutral-50 p-6 space-y-4">
            <h2 class="text-xl font-semibold">合集 / 系列下载</h2>
            <p class="text-sm text-neutral-500">这里可以接入合集链接、系列详情解析和整组下载。</p>
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

