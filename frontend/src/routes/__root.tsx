import {createRootRoute, Outlet} from '@tanstack/solid-router'
import {createSignal, type JSXElement, onMount, Show} from "solid-js";
import {GetTheme, SetCloseToTray} from "@bindings/github.com/kamiertop/videodown/utils/settings";
import {ForceQuit, HideWindow} from "@bindings/github.com/kamiertop/videodown/internal/app/controller";
import {Application, Events} from "@wailsio/runtime";
import {DownloadUpdate, InstallUpdate, LatestResult} from "@bindings/github.com/kamiertop/videodown/internal/updater/updater";
import type {Result} from "@bindings/github.com/kamiertop/videodown/internal/updater";
import HomeHeader from "../components/Header.tsx";

export const Route = createRootRoute({
  component: RootComponent,
})

function RootComponent(): JSXElement {
  const [showChoice, setShowChoice] = createSignal(false);
  const [rememberChoice, setRememberChoice] = createSignal(false);
  const [update, setUpdate] = createSignal<Result | null>(null);
  const [installing, setInstalling] = createSignal(false);

  onMount(async () => {
    const theme: string = await GetTheme().catch(() => 'light');
    document.documentElement.setAttribute('data-theme', theme);

    Events.On("before-close-prompt", () => {
      setRememberChoice(false);
      setShowChoice(true);
    });

    Events.On("update-available", ({data}) => {
      setUpdate(data as Result);
    });
    // The startup check can finish before the frontend listener is mounted.
    const cached = await LatestResult().catch(() => null);
    if (cached?.available) setUpdate(cached);
  });

  async function handleChoice(minimize: boolean) {
    setShowChoice(false);
    if (rememberChoice()) {
      await SetCloseToTray(minimize);
    }
    if (!minimize) {
      if (!rememberChoice()) {
        await ForceQuit();
      }
      Application.Quit();
      return;
    }
    HideWindow();
  }

  async function installUpdate() {
    const result = update();
    if (!result || installing()) return;
    setInstalling(true);
    try {
      const packagePath = await DownloadUpdate(result.downloadURL);
      await InstallUpdate(packagePath);
    } catch (error) {
      setInstalling(false);
      window.alert("启动更新失败：" + (error instanceof Error ? error.message : String(error)));
    }
  }

  return (
      <div class="h-screen bg-base-200 flex flex-col">
        <HomeHeader/>
        <div class="flex-1 min-h-0 ">
          <Outlet/>
        </div>

        <Show when={update()}>
          {(result) => (
            <div class="fixed bottom-4 left-4 z-40 w-[calc(100%-2rem)] max-w-md rounded-lg border border-info/30 bg-base-100 p-4 shadow-xl">
              <button class="btn btn-ghost btn-sm btn-circle absolute right-2 top-2" aria-label="关闭更新提示" onClick={() => setUpdate(null)}>×</button>
              <h3 class="font-semibold">发现新版本 {result().latestVersion}</h3>
              <p class="mt-1 text-sm text-base-content/70">当前版本 {result().currentVersion}</p>
              <Show when={result().releaseNotes}>
                <p class="mt-2 max-h-20 overflow-y-auto whitespace-pre-wrap text-sm text-base-content/70">{result().releaseNotes}</p>
              </Show>
              <Show when={result().downloadURL}>
                <button class="btn btn-info btn-sm mt-3" disabled={installing()} onClick={installUpdate}>{installing() ? "正在启动更新…" : "下载并安装更新"}</button>
              </Show>
            </div>
          )}
        </Show>

        <Show when={showChoice()}>
          <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
            <div class="w-[calc(100%-2rem)] max-w-sm rounded-lg bg-base-100 p-6 text-base-content shadow-2xl">
              <h3 class="text-lg font-bold mb-2">关闭窗口</h3>
              <p class="text-sm text-base-content/70 mb-4">
                你希望现在如何关闭窗口？
              </p>
              <label class="label cursor-pointer justify-start gap-3 px-0 mb-5">
                <input
                    type="checkbox"
                    class="checkbox checkbox-info checkbox-sm"
                    checked={rememberChoice()}
                    onChange={(event) => setRememberChoice(event.currentTarget.checked)}
                />
                <span class="label-text">记住选择，下次不再询问</span>
              </label>
              <div class="flex flex-col gap-3">
                <button
                    class="btn btn-outline btn-info"
                    onClick={() => handleChoice(true)}
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
                  </svg>
                  最小化到托盘
                </button>
                <button
                    class="btn btn-ghost"
                    onClick={() => handleChoice(false)}
                >
                  退出程序
                </button>
                <button
                    class="btn btn-ghost btn-sm"
                    onClick={() => {
                      setRememberChoice(false);
                      setShowChoice(false);
                    }}
                >
                  取消
                </button>
              </div>
            </div>
          </div>
        </Show>
      </div>
  )
}
