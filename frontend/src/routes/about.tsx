import { createFileRoute } from '@tanstack/solid-router'
import type {JSXElement} from "solid-js";

export const Route = createFileRoute('/about')({
  component: AboutComponent,
})

function AboutComponent(): JSXElement {
  return (
    <div class="flex flex-col gap-6 p-8">
      <div class="text-center">
        <h1 class="text-4xl font-bold text-primary mb-2">关于 VideoDown</h1>
        <p class="text-base-content/70">一款简洁高效的视频下载工具</p>
      </div>

      <div class="card w-full max-w-3xl bg-base-100 shadow-xl mx-auto">
        <div class="card-body">
          <div class="flex items-center gap-4 mb-6">
            <div class="avatar placeholder">
              <div class="bg-primary text-primary-content rounded-2xl w-20">
                <span class="text-3xl">🎬</span>
              </div>
            </div>
            <div>
              <h2 class="card-title text-2xl">VideoDown</h2>
              <p class="text-sm text-base-content/70">版本 v1.0.0</p>
            </div>
          </div>

          <div class="divider"></div>

          <div class="prose max-w-none">
            <h3 class="text-xl font-semibold mb-3">✨ 功能特性</h3>
            <ul class="space-y-2 text-base-content/80">
              <li class="flex items-start gap-2">
                <span class="text-success mt-1">✓</span>
                <span>支持 B 站、抖音等多个主流视频平台</span>
              </li>
              <li class="flex items-start gap-2">
                <span class="text-success mt-1">✓</span>
                <span>批量下载，自动合并音视频</span>
              </li>
              <li class="flex items-start gap-2">
                <span class="text-success mt-1">✓</span>
                <span>多种画质选择，最高支持 4K</span>
              </li>
              <li class="flex items-start gap-2">
                <span class="text-success mt-1">✓</span>
                <span>智能去水印功能</span>
              </li>
              <li class="flex items-start gap-2">
                <span class="text-success mt-1">✓</span>
                <span>跨平台支持（Windows、macOS、Linux）</span>
              </li>
            </ul>
          </div>

          <div class="divider"></div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mt-6">
            <div class="stat bg-base-200 rounded-box">
              <div class="stat-figure text-primary">
                <span class="text-4xl">🚀</span>
              </div>
              <div class="stat-title">高速下载</div>
              <div class="stat-value text-primary">多线程</div>
              <div class="stat-desc">提升下载效率</div>
            </div>
            
            <div class="stat bg-base-200 rounded-box">
              <div class="stat-figure text-secondary">
                <span class="text-4xl">🔒</span>
              </div>
              <div class="stat-title">安全可靠</div>
              <div class="stat-value text-secondary">无广告</div>
              <div class="stat-desc">纯净体验</div>
            </div>
            
            <div class="stat bg-base-200 rounded-box">
              <div class="stat-figure text-accent">
                <span class="text-4xl">💎</span>
              </div>
              <div class="stat-title">精心打造</div>
              <div class="stat-value text-accent">开源</div>
              <div class="stat-desc">持续更新</div>
            </div>
          </div>

          <div class="divider"></div>

          <div class="text-center mt-6">
            <p class="text-base-content/70 mb-4">
              使用本软件时请遵守相关法律法规，尊重创作者版权
            </p>
            <div class="flex justify-center gap-4">
              <button class="btn btn-outline btn-primary">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                </svg>
                GitHub 仓库
              </button>
              <button class="btn btn-outline btn-secondary">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                </svg>
                使用文档
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
