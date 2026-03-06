import {For, JSXElement} from "solid-js";

const tips = ["输入 UP 主主页链接或 mid", "支持批量列出视频", "后续可加全选下载"];

export default function BilibiliUploader(): JSXElement {
    return (
        <div class="rounded-3xl border border-neutral-200 bg-neutral-50 p-6 space-y-4">
            <h2 class="text-xl font-semibold">UP主视频下载</h2>
            <p class="text-sm text-neutral-500">这里可以接入 UP 主主页链接、分页抓取和批量任务列表。</p>
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

