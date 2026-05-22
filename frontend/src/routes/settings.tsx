import {createFileRoute} from '@tanstack/solid-router'
import {createSignal, For, type JSXElement} from "solid-js";
import {AboutSection} from "../components/settings/about";
import {BilibiliSection} from "../components/settings/bilibili";
import {DownloadSection} from "../components/settings/download";
import {FFmpegSection} from "../components/settings/ffmpeg";
import {GeneralSection} from "../components/settings/general";

export const Route = createFileRoute('/settings')({
  component: SettingsComponent,
})

interface NavSection {
  id: string;
  label: string;
  icon: JSXElement;
}

const NAV_SECTIONS: NavSection[] = [
  {
    id: 'general',
    label: '通用',
    icon: (
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
        </svg>
    ),
  },
  {
    id: 'bilibili',
    label: '哔哩哔哩',
    icon: (
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/>
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
    ),
  },
  {
    id: 'download',
    label: '下载',
    icon: (
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
          />
        </svg>
    ),
  },
  {
    id: 'FFmpeg',
    label: 'FFmpeg',
    icon: (
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/>
        </svg>
    )
  },
  {
    id: 'about',
    label: '关于',
    icon: (
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
    ),
  },
];

function SettingsComponent(): JSXElement {
  const [active, setActive] = createSignal('general');

  return (
      <div class="flex flex-col md:flex-row h-full">
        <nav class="w-full md:w-56 shrink-0 border-b md:border-b-0 md:border-r border-base-300 p-4">
          <ul class="flex md:flex-col gap-1 overflow-x-auto">
            <For each={NAV_SECTIONS}>
              {(section) => (
                  <button
                      class={`flex items-center gap-3 px-3 py-2.5 rounded-lg w-full whitespace-nowrap text-sm transition-colors
                          ${active() === section.id
                          ? 'bg-base-200 font-semibold text-base-content'
                          : 'text-base-content/70 hover:bg-base-200/50 hover:text-base-content'
                      }`}
                      onClick={() => setActive(section.id)}
                  >
                    {section.icon}
                    {section.label}
                  </button>
              )}
            </For>
          </ul>
        </nav>

        <div class="flex-1 overflow-y-auto p-6">
          {active() === 'general' && <GeneralSection/>}
          {active() === 'bilibili' && <BilibiliSection/>}
          {active() === 'download' && <DownloadSection/>}
          {active() === 'FFmpeg' && <FFmpegSection/>}
          {active() === 'about' && <AboutSection/>}
        </div>
      </div>
  )
}
