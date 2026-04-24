import {createFileRoute} from '@tanstack/solid-router'
import {For, type JSXElement, Show} from "solid-js";
import EmptyState from "../../components/EmptyState.tsx";
import {clearDouyinVideos, douyinVideoList, removeDouyinVideo} from "../../lib/douyinStore.ts";
import {formatCount, formatDate, formatDuration} from "../../lib/format.ts";

export const Route = createFileRoute('/douyin/download')({
  component: DouyinDownloadPage,
})

function DouyinDownloadPage(): JSXElement {
  return (
    <section class="flex h-full min-h-0 flex-col p-4">
      <header class="mb-3 rounded-lg border border-base-300 bg-base-100 px-4 py-3">
        <div class="flex items-center gap-3">
          <div class="min-w-0 flex-1">
            <h2 class="text-base font-bold">下载中心</h2>
            <p class="text-sm text-base-content/60">从收藏视频页加入的内容会先出现在这里。</p>
          </div>
          <Show when={douyinVideoList().length > 0}>
            <button class="btn btn-ghost btn-sm text-error" onClick={clearDouyinVideos}>
              清空列表
            </button>
          </Show>
        </div>
      </header>

      <div class="min-h-0 flex-1 overflow-hidden rounded-lg border border-base-300 bg-base-100">
        <Show
          when={douyinVideoList().length > 0}
          fallback={<EmptyState title="下载列表为空" description="可以从收藏视频页勾选后加入下载列表。"/>}
        >
          <div
            class="grid h-full grid-cols-2 gap-3 overflow-auto p-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
            <For each={douyinVideoList()}>
              {(item) => (
                <article class="overflow-hidden rounded-lg border border-base-300 bg-base-100">
                  <div class="relative w-full overflow-hidden bg-base-200 aspect-3/5">
                    <Show
                      when={item.cover}
                      fallback={<div
                        class="absolute inset-0 flex items-center justify-center text-xs text-base-content/35">无封面</div>}
                    >
                      <img
                        src={item.cover}
                        alt={item.title}
                        class="h-full w-full object-cover"
                        loading="lazy"
                        decoding="async"
                        referrerPolicy="no-referrer"
                      />
                    </Show>
                    <span
                      class="absolute bottom-1 right-1 rounded-md bg-black/65 px-1.5 py-1 text-[0.65rem] font-medium tabular-nums leading-none text-white">
                      {formatDuration(item.duration)}
                    </span>
                  </div>
                  <div class="p-2.5">
                    <h3 class="line-clamp-2 text-[13px] font-semibold leading-5 text-base-content" title={item.title}>
                      {item.title}
                    </h3>
                    <p class="mt-1 line-clamp-1 text-[11px] text-base-content/50">@{item.authorName}</p>
                    <p class="mt-1.5 text-[11px] text-base-content/45">
                      发布 {item.publishTime ? formatDate(item.publishTime) : "-"}
                    </p>
                    <div class="mt-2 flex items-center gap-1.5 text-[11px] text-base-content/55">
                      <span
                        class="rounded-full bg-base-200 px-2 py-0.5 tabular-nums">赞 {formatCount(item.diggCount)}</span>
                      <span
                        class="rounded-full bg-base-200 px-2 py-0.5 tabular-nums">藏 {formatCount(item.collectCount)}</span>
                    </div>
                    <div class="mt-3 flex items-center justify-between gap-2">
                      <a
                        class="link link-hover text-[11px] text-primary"
                        href={item.link}
                        target="_blank"
                        rel="noreferrer"
                      >
                        打开
                      </a>
                      <button class="btn btn-ghost btn-xs text-error" onClick={() => removeDouyinVideo(item.awemeId)}>
                        移除
                      </button>
                    </div>
                  </div>
                </article>
              )}
            </For>
          </div>
        </Show>
      </div>
    </section>
  )
}
