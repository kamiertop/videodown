import {createFileRoute} from '@tanstack/solid-router'
import {createSignal, type JSXElement, onMount} from "solid-js";
import {SetStorage} from "../../../wailsjs/go/main/App";
import {GetStorage} from "../../../wailsjs/go/utils/Settings";
import ThemeChange from "../../components/ThemeChange.tsx";
import Toast from "../../components/Toast";
import {useToast} from "../../hooks/useToast";

// @ts-ignore
export const Route = createFileRoute('/settings/')({
    component: SettingsComponent,
})


function StorageDirectory(): JSXElement {
    const [storagePath, setStoragePath] = createSignal<string>("");
    const {message, type, showToast} = useToast();

    onMount(async () => {
        try {
            const path = await GetStorage();
            setStoragePath(path);
        } catch (error) {
            const errorMsg = error instanceof Error ? error.message : '获取存储路径失败';
            showToast(errorMsg, 'error');
        }
    })

    async function selectDirectory() {
        try {
            const path = await SetStorage();
            if (path) {
                setStoragePath(path);
                showToast('下载目录设置成功', 'success');
            }
        } catch (error) {
            const errorMsg = error instanceof Error ? error.message : '选择目录失败';
            showToast(errorMsg, 'error');
        }
    }

    return (
        <>
            <div class="card bg-base-100 shadow-xl">
                <div class="card-body">
                    <h2 class="card-title mb-4">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-primary" fill="none"
                             viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                  d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
                        </svg>
                        下载路径
                    </h2>
                    <div class="form-control">
                        <div class="flex gap-2">
                            <input
                                type="text"
                                value={storagePath()}
                                readonly
                                class="input input-bordered flex-1"
                                placeholder="选择下载路径"
                            />
                            <button class="btn btn-primary" onClick={selectDirectory}>
                                选择目录
                            </button>
                        </div>
                    </div>
                </div>
            </div>
            <Toast message={message()} type={type()}/>
        </>
    )
}


// 保存偏好，每个UP主下的视频单独存储，合集|系列视频也单独存储
function SavePreferences(): JSXElement {
    return (
        <></>
    )
}

function SettingsComponent(): JSXElement {
    return (
        <div class=" p-8 grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* 下载路径设置 */}

            <StorageDirectory/>
            {/* 画质设置 */}
            <div class="card bg-base-100 shadow-xl">
                <div class="card-body">
                    <h2 class="card-title mb-4">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-secondary" fill="none"
                             viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                  d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                  d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
                        </svg>
                        默认画质
                    </h2>
                    <div class="form-control">
                        <label class="label cursor-pointer justify-between">
                            <span class="label-text">优先选择画质</span>
                            <select class="select select-bordered select-sm w-full max-w-xs">
                                <option>4K 超清</option>
                                <option selected>1080P 高码率</option>
                                <option>720P</option>
                                <option>自动</option>
                            </select>
                        </label>
                    </div>
                </div>
            </div>

            {/* 并发控制 */}
            <div class="card bg-base-100 shadow-xl">
                <div class="card-body">
                    <h2 class="card-title mb-4">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-accent" fill="none"
                             viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                  d="M13 10V3L4 14h7v7l9-11h-7z"/>
                        </svg>
                        下载控制
                    </h2>

                    <div class="w-full max-w-xs">
                        <input type="range" min={1} max="5" value="3" class="range range-secondary" step="1"/>
                        <div class="flex justify-between px-2.5 mt-2 text-xs">
                            <span>|</span>
                            <span>|</span>
                            <span>|</span>
                            <span>|</span>
                            <span>|</span>
                        </div>
                        <div class="flex justify-between px-2.5 mt-2 text-xs">
                            <span>1</span>
                            <span>2</span>
                            <span>3</span>
                            <span>4</span>
                            <span>5</span>
                        </div>
                    </div>
                </div>
            </div>
            
            <SavePreferences/>
            <ThemeChange/>
        </div>
    )
}
