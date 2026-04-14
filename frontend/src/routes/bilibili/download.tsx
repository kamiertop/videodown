import {createFileRoute} from '@tanstack/solid-router'
import {
    createSignal, For,
    type JSXElement,
} from "solid-js";
import {extractBvid} from "../../lib/format";
import Toast from "../../components/Toast";
import {useToast} from "../../hooks/useToast";
import type {MediaCardItem} from "../../lib/model.ts";


export const Route = createFileRoute('/bilibili/download')({
    component: DownLoad,
})


const [videoList, setVideoList] = createSignal<MediaCardItem[]>([]);

export function addVideo(selectedVideos: MediaCardItem[]): void {
    // 向页面中添加下载链接
    setVideoList([...videoList(), ...selectedVideos]);
}

// @ts-ignore - Will be used when implementing video removal
function removeVideo(index: number): void {
    // 从页面中移除下载链接
    setVideoList(
        prev => prev.filter((_, i) => i !== index)
    );
}

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
        <div class={"flex flex-col pt-4 pl-4 pr-4 pb-4"}>
            {/*解析框组件*/}
            <section class={""}>
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
            <section>
                {/*视频信息展示组件*/}
                <For each={videoList()}>
                    {(item) => (
                        <div>
                            {item.bvid}
                        </div>
                    )}
                </For>
            </section>
            <Toast message={message()} type={type()}/>
        </div>
    )
}


