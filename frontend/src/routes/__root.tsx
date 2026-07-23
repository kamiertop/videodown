import {createRootRoute, Outlet} from '@tanstack/solid-router'
import {createSignal, type JSXElement, onMount, Show} from "solid-js";
import {ForceQuit, GetTheme, HideWindow, SetCloseToTray} from "../../wailsjs/go/utils/Settings";
import {EventsOn, Quit} from "../../wailsjs/runtime";
import HomeHeader from "../components/Header.tsx";

export const Route = createRootRoute({
  component: RootComponent,
})

function RootComponent(): JSXElement {
  const [showChoice, setShowChoice] = createSignal(false);
  const [rememberChoice, setRememberChoice] = createSignal(false);

  onMount(async () => {
    const theme: string = await GetTheme().catch(() => 'light');
    document.documentElement.setAttribute('data-theme', theme);

    EventsOn("before-close-prompt", () => {
      setRememberChoice(false);
      setShowChoice(true);
    });
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
      Quit();
      return;
    }
    HideWindow();
  }

  return (
      <div class="h-screen bg-base-200 flex flex-col">
        <HomeHeader/>
        <div class="flex-1 min-h-0 ">
          <Outlet/>
        </div>

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
