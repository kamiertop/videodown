import {createFileRoute, Link} from '@tanstack/solid-router'
import type {JSXElement} from "solid-js";

export const Route = createFileRoute('/bilibili/up/$mid')({
    component: UpDetail,
})

function UpDetail(): JSXElement {
    const params = Route.useParams();

    return (
        <section class="flex h-full min-h-0 flex-col overflow-hidden bg-base-200/40 p-3">
            {/* 顶栏 */}
            <div class="mb-3 flex shrink-0 items-center gap-3 rounded-xl border border-base-300 bg-base-100 px-4 py-3">
                <Link
                    to="/bilibili/up"
                    class="btn btn-ghost btn-sm gap-1"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24"
                         stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7"/>
                    </svg>
                    返回
                </Link>
                <div class="h-5 w-px bg-base-300"></div>
                <h2 class="text-sm font-bold text-base-content">UP主详情</h2>
                <span class="rounded-full bg-base-200 px-2 py-0.5 text-xs tabular-nums text-base-content/60">
                    mid: {params().mid}
                </span>
            </div>

            {/* 内容区 - 待接入后端投稿视频 API */}
            <div class="flex min-h-0 flex-1 items-center justify-center overflow-auto rounded-xl border border-base-300 bg-base-100">
                <div class="text-center">
                    <svg xmlns="http://www.w3.org/2000/svg" class="mx-auto h-16 w-16 text-base-content/20" fill="none"
                         viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
                        <path stroke-linecap="round" stroke-linejoin="round"
                              d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                    </svg>
                    <p class="mt-3 text-sm font-semibold text-base-content/60">UP主投稿视频</p>
                    <p class="mt-1 max-w-xs text-xs text-base-content/40">
                        待接入后端投稿视频 API 后，这里将展示该 UP主 的所有投稿视频，支持批量选择下载
                    </p>
                </div>
            </div>
        </section>
    );
}
