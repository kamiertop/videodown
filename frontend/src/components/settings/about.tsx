import {createSignal, type JSXElement, onMount} from "solid-js";
import {GetVersion} from "../../../wailsjs/go/utils/Settings";
import {BrowserOpenURL} from "../../../wailsjs/runtime";

import app from "../../assets/app-128.png";

const GITHUB_REPO_URL = 'https://github.com/kamiertop/videodown';

export function AboutSection(): JSXElement {
  const [version, setVersion] = createSignal('');

  onMount(async () => {
    const v = await GetVersion();
    setVersion(v || 'dev');
  });

  return (
      <div class="max-w-lg space-y-6 mx-auto">
        <div class="text-center space-y-3">
          <img src={app} alt="VideoDown" class="mx-auto w-20 h-20 object-contain"/>
          <h1 class="text-2xl font-bold tracking-tight">VideoDown</h1>
          <span class="badge badge-primary badge-lg">{version()}</span>
          <p class="text-sm text-base-content/60">哔哩哔哩 · 抖音 视频下载工具</p>
        </div>

        <div class="divider"/>

        <div class="bg-base-100 rounded-box border border-base-300/60 p-5 space-y-2">
          <h3 class="flex items-center gap-2 text-base font-semibold">
            <span class="text-lg">✨</span>功能特性
          </h3>
          <ul class="space-y-1.5 text-sm text-base-content/70">
            <li class="flex items-start gap-2">
              <span class="text-primary shrink-0 mt-0.5">▸</span>
              <strong>哔哩哔哩</strong> — 视频、番剧、收藏夹、UP 主作品
            </li>
            <li class="flex items-start gap-2">
              <span class="text-primary shrink-0 mt-0.5">▸</span>
              <strong>抖音</strong> — 视频、合集、喜欢列表、用户作品
            </li>
            <li class="flex items-start gap-2">
              <span class="text-primary shrink-0 mt-0.5">▸</span>
              多线程并发下载，下载后自动随机休眠防止被封
            </li>
            <li class="flex items-start gap-2">
              <span class="text-primary shrink-0 mt-0.5">▸</span>
              多主题切换、历史管理、自动分目录整理下载内容
            </li>
          </ul>
        </div>

        <div class="bg-base-100 rounded-box border border-base-300/60 p-5 space-y-2">
          <h3 class="flex items-center gap-2 text-base font-semibold">
            <span class="text-lg">📖</span>开源说明
          </h3>
          <p class="text-sm text-base-content/70">
            本软件完全<span class="text-neutral text-xl font-extrabold">开源免费</span>
            ，源代码托管于 GitHub，欢迎 Star、Fork 与 Pull Request
          </p>
          <div class="flex items-start gap-2 text-sm text-error">
            <span class="shrink-0 font-bold">⚠</span>
            <span>禁止用于商业用途，禁止转售、倒卖、二次打包收费。</span>
          </div>
        </div>

        <div class="bg-base-100 rounded-box border border-base-300/60 p-5 space-y-2">
          <h3 class="flex items-center gap-2 text-base font-semibold">
            <span class="text-lg">⚖️</span>免责声明
          </h3>
          <p class="text-sm text-base-content/60 leading-relaxed">
            本软件仅提供视频下载技术实现，不收集任何用户信息。用户下载的视频版权归属原平台及创作者所有。请遵守相关平台的使用条款，在合法授权范围内使用。因使用本软件产生的任何法律纠纷与开发者无关
          </p>
        </div>

        <div class="text-center space-y-3 pt-2">
          <p class="text-xs text-base-content/40">Go · SolidJS · TailwindCSS · DaisyUI · Wails</p>
          <button class="btn btn-ghost btn-sm gap-2" onClick={() => {
            try {
              BrowserOpenURL(GITHUB_REPO_URL);
            } catch {
              window.open(GITHUB_REPO_URL, '_blank', 'noopener,noreferrer');
            }
          }}>
            <svg class="h-5 w-5 fill-current" viewBox="0 0 24 24">
              <path
                  d="M12 0C5.374 0 0 5.373 0 12c0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0112 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.3 24 12c0-6.627-5.373-12-12-12z"/>
            </svg>
            GitHub
          </button>
        </div>
      </div>
  )
}
