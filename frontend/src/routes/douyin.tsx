import { createFileRoute } from '@tanstack/solid-router'
import type {JSXElement} from "solid-js";

export const Route = createFileRoute('/douyin')({
  component: DouyinComponent,
})

function DouyinComponent(): JSXElement {
  return (
    <div class="flex flex-col gap-6 p-8">
      <div>
        <h1 class="text-4xl font-bold text-error">抖音视频下载</h1>
        <p class="mt-2 text-base-content/70">
          支持下载抖音短视频、合集、直播回放等内容
        </p>
      </div>

      <div class="card w-full max-w-2xl bg-base-100 shadow-xl">
        <div class="card-body">
          <div class="form-control">
            <label class="label">
              <span class="label-text font-semibold">视频链接</span>
            </label>
            <input 
              type="text" 
              placeholder="请输入抖音视频链接（如：https://v.douyin.com/xxxxx）" 
              class="input input-bordered w-full focus:input-error"
            />
            <label class="label">
              <span class="label-text-alt text-base-content/50">
                支持分享链接、作品链接等多种格式
              </span>
            </label>
          </div>

          <div class="form-control mt-4">
            <label class="label cursor-pointer justify-start gap-4">
              <span class="label-text font-semibold">下载画质：</span>
              <select class="select select-bordered w-full max-w-xs">
                <option disabled selected>选择画质</option>
                <option>原画（无水印）</option>
                <option>高清 1080P</option>
                <option>标清 720P</option>
              </select>
            </label>
          </div>

          <div class="form-control mt-4">
            <label class="label cursor-pointer justify-start gap-4">
              <span class="label-text font-semibold">是否去水印：</span>
              <input 
                type="checkbox" 
                class="toggle toggle-error toggle-lg" 
                checked
              />
              <span class="label-text text-success font-semibold">✓ 已开启</span>
            </label>
          </div>

          <div class="card-actions justify-end mt-6">
            <button class="btn btn-error btn-wide">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              解析并下载
            </button>
          </div>
        </div>
      </div>

      <div class="divider"></div>

      <div class="alert alert-info shadow-lg">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
        </svg>
        <div>
          <h3 class="font-bold">温馨提示</h3>
          <div class="text-xs">
            <ul class="list-disc list-inside mt-2 space-y-1">
              <li>自动去除视频水印，保持原画质量</li>
              <li>支持批量下载功能</li>
              <li>请尊重创作者版权，合理使用</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}
