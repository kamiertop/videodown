import {createFileRoute} from '@tanstack/solid-router'
import {createSignal, For, type JSXElement, Show,} from "solid-js";
import IconChat from "../../components/icons/IconChat.tsx";
import IconEye from "../../components/icons/IconEye.tsx";
import Toast from "../../components/Toast";
import {useToast} from "../../hooks/useToast";
import {videoList} from "../../lib/bilibiliStore.ts";
import {extractBvid, formatCount, formatDate, formatDuration} from "../../lib/format";


export const Route = createFileRoute('/bilibili/download')({
    component: DownLoad,
})


function DownLoad(): JSXElement {
    const [videoURL, setVideoURL] = createSignal<string>("");
    const {message, type, showToast} = useToast();

    function parseVideo() {
        const bvid = extractBvid(videoURL());
        if (bvid === null) {
            showToast('请输入有效的 B 站视频链接或 BV 号', 'error');
            return;
        }

        showToast(`成功解析 BV 号: ${bvid}`, 'success');
        // TODO: 发送请求到后端，获取视频信息
    }

    return (
        <div class={"flex flex-col pt-4 pl-4 pr-4 pb-4 h-full"}>
            {/*解析框组件*/}
            <section>
                <div class={"flex flex-row join gap-2"}>
                    <input type="text" placeholder="请输入视频链接, 支持BV号、AV号、视频URL等格式, 可按回车直接解析"
                           value={videoURL()}
                           onInput={(e) => setVideoURL(e.currentTarget.value)}
                           onkeydown={(e) => {
                               if (e.key === "Enter") {
                                   parseVideo();
                               }
                           }}
                           class="input input-success w-full"/>
                    <div class={"btn btn-outline btn-secondary"}
                         onClick={parseVideo}
                    >
                        解析
                    </div>
                    <div class={"btn btn-outline btn-info"}
                         onClick={() => setVideoURL("")}
                    >
                        清空
                    </div>
                </div>
            </section>
            {/*下载操作栏*/}
            <Show when={videoList().length > 0}>
                <section
                    class="mt-2 rounded-lg p-3 flex flex-row justify-between items-center shadow-sm">
                    <div class="flex items-center gap-2">
                        <div class="badge badge-primary">{videoList().length}</div>
                        <span class="text-xs">个视频待下载</span>
                    </div>
                    <button class="btn btn-success btn-xs gap-1.5">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24"
                             stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                  d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/>
                        </svg>
                        开始下载
                    </button>
                </section>
            </Show>
            <section class={"mt-3 flex-1 overflow-y-auto mb-4"}>
                <For each={videoList()}>
                    {(item) => (
                        <div class={"flex flex-row gap-4 pt-3"}>
                            <div class={"relative w-40 overflow-hidden rounded-lg shrink-0"}>
                                <img class={"w-full h-full object-cover"} src={item.cover} alt={item.title}
                                     referrerPolicy="no-referrer"/>
                                <div
                                    class={"pointer-events-none absolute inset-x-0 bottom-0 flex items-end justify-between bg-linear-to-t from-black/70 to-transparent px-2 pb-1.5 pt-5"}>
                                    <div class="flex items-center gap-2.5 text-white/90">
                                        <span class={"flex flex-items-center gap-0.5 text-xs"}>
                                            <IconEye class="h-3 w-3"/>
                                            {formatCount(item.play)}
                                        </span>
                                        <span class={"flex flex-items-center gap-0.5 text-xs"}>
                                            <IconChat class="h-3 w-3"/> {item.danmaku}
                                        </span>
                                    </div>
                                    <span class="rounded bg-black/60 px-1 py-0.5 text-xs tabular-nums text-white/90">
                                        {formatDuration(item.duration)}
                                    </span>
                                </div>
                            </div>
                            <div class="flex flex-col">
                                <div>{item.title}</div>
                                <div class="flex flex-row items-center gap-1 text-xs text-base-content/55 pt-2">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 shrink-0"
                                         viewBox={"0 0 20 20"} fill="currentColor">
                                        <path fill-rule="evenodd"
                                              d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"
                                              clip-rule="evenodd"
                                        />
                                    </svg>
                                    <span>{item.upperName}</span> | <span>{formatDate(item.pubtime)}</span>
                                </div>
                            </div>
                        </div>
                    )}
                </For>
            </section>
            <Toast message={message()} type={type()}/>
        </div>
    )
}


