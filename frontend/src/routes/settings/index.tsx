import {createFileRoute} from '@tanstack/solid-router'
import {type JSXElement} from "solid-js";
import ThemeChange from "../../components/ThemeChange.tsx";

// @ts-ignore
export const Route = createFileRoute('/settings/')({
    component: SettingsComponent,
})

function SettingsComponent(): JSXElement {

    return (
        <div class="flex flex-col gap-6 p-8 grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* 下载路径设置 */}
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
                        <label class="label">
                            <span class="label-text">保存位置</span>
                        </label>
                        <div class="join">
                            <input
                                type="text"
                                value="/home/user/Videos"
                                class="input input-bordered join-item flex-1"
                                readonly
                            />
                            <button class="btn btn-primary join-item">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none"
                                     viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                          d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"/>
                                </svg>
                                浏览
                            </button>
                        </div>
                    </div>
                </div>
            </div>

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

            {/* 通知设置 */}
            <div class="card bg-base-100 shadow-xl">
                <div class="card-body">
                    <h2 class="card-title mb-4">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-info" fill="none"
                             viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                  d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"/>
                        </svg>
                        通知提醒
                    </h2>
                    <div class="space-y-3">
                        <label class="label cursor-pointer justify-between">
                            <span class="label-text">下载完成通知</span>
                            <input type="checkbox" class="toggle toggle-info" checked/>
                        </label>
                        <label class="label cursor-pointer justify-between">
                            <span class="label-text">错误提示音</span>
                            <input type="checkbox" class="toggle toggle-info"/>
                        </label>
                    </div>
                </div>
            </div>

            <ThemeChange/>
        </div>
    )
}
