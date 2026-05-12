import {createFileRoute} from '@tanstack/solid-router'
import {createSignal, For, type JSXElement, onCleanup, onMount} from "solid-js";
import {
  GetConcurrencyNum,
  GetSavePreference,
  GetSleepTime,
  GetStorage,
  GetTheme,
  SetConcurrencyNum,
  SetSleepTime,
  SetStorage,
  SetTheme
} from "../../wailsjs/go/utils/Settings";
import Toast from "../components/Toast";
import {useToast} from "../hooks/useToast";

// @ts-ignore
export const Route = createFileRoute('/settings')({
  component: SettingsComponent,
})


function SettingsComponent(): JSXElement {
  return (
      <div class="pt-5 pl-6 pr-6 grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* 下载路径设置 */}
        <StorageDirectory/>
        <ConcurrencyNum/>
        <SleepAfterDownLoad/>
        <SavePreferences/>
        <ThemeChange/>
        <div class="flex flex-2">
          <AutoUpdate/>
          <AutoDownloadFFmpeg/>
        </div>
      </div>
  )
}

// 存储位置
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
    } catch (err) {
      showToast(String(err), 'error');
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
  const [allowGroup, setAllowGroup] = createSignal<boolean>(true);
  const {message, type, showToast} = useToast();
  const [text, setText] = createSignal<string>("保存时自动按照作者名称 | 合集 | 收藏夹名称分目录存储");
  onMount(async () => {
    try {
      const allow: boolean = await GetSavePreference();
      setAllowGroup(allow);
    } catch (error) {
      showToast("获取保存偏好失败" + error, 'error');
    }
  })

  async function setPreference(allow: boolean) {
    try {
      setAllowGroup(allow);
      if (!allow) {
        setText("保存时全部存储至公共下载目录")
      } else {
        setText("保存时自动按照作者名称 | 合集 | 收藏夹名称分目录存储")
      }
      showToast('下载目录设置成功', 'success');
    } catch (error) {
      showToast("更新保存偏好失败", 'error');
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
              保存时自动分组
            </h2>
            <div class="form-control">
              <div class="flex gap-2">
                <input
                    type="checkbox"
                    checked={allowGroup()}
                    class="toggle toggle-warning"
                    onClick={() => setPreference(!allowGroup())}
                />
              </div>
              <div class="mt-5">
              <span class="text-base-content border-b">
                {text()}
              </span>
              </div>
            </div>
          </div>
        </div>
        <Toast message={message()} type={type()}/>
      </>
  )
}

// 下载后休眠时间设置，防止速度过快被封
function SleepAfterDownLoad(): JSXElement {
  const [seconds, setSeconds] = createSignal<number>(0);
  const {message, type, showToast} = useToast();
  const presets = [0, 10, 30, 60, 120, 300];


  let saveTimer: number | undefined;
  let saveSeq: number = 0;

  function normalizeSeconds(value: number): number {
    if (!Number.isFinite(value)) return 0;
    return Math.max(0, Math.min(3600, Math.round(value)));
  }

  function updateSeconds(value: number): void {
    const next = normalizeSeconds(value);
    setSeconds(next);
    scheduleSave(next);
  }


  function scheduleSave(next: number): void {
    if (saveTimer !== undefined) {
      window.clearTimeout(saveTimer);
    }
    saveTimer = window.setTimeout(() => {
      void persistSeconds(next);
    }, 300);
  }

  async function persistSeconds(next: number): Promise<void> {
    const seq = ++saveSeq;
    try {
      await SetSleepTime(next);
    } catch (error) {
      if (seq === saveSeq) {
        showToast("保存休眠时间失败" + error, "error");
      }
    }
  }

  onMount(async () => {
    try {
      const time = await GetSleepTime();
      const loaded = normalizeSeconds(Number(time));
      setSeconds(loaded);
    } catch (error) {
      showToast("获取下载后休眠时间失败" + error, 'error');
    }
  })

  onCleanup(() => {
    if (saveTimer !== undefined) {
      window.clearTimeout(saveTimer);
    }
  });

  return (
      <>
        <div class="card bg-base-100 shadow-xl">
          <div class="card-body">
            <div class="flex items-start justify-between gap-4">
              <div>
                <h2 class="card-title">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-primary" fill="none"
                       viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                          d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"/>
                  </svg>
                  下载后休眠
                </h2>
                <div class="flex items-center justify-between text-sm text-base-content/70">
                  <div>每个视频下载完成后随机休眠&nbsp;
                    <span class="text-secondary text-xl font-bold">
                    [0-{seconds()}]
                  </span>
                    &nbsp;秒
                  </div>
                </div>
              </div>
              <div class="badge badge-primary badge-outline shrink-0">
                {seconds()} 秒
              </div>
            </div>
            <div class="flex items-center gap-3">
              <input
                  type="range" min="0" max="600" step="5" value={seconds()}
                  class="range range-primary flex-1"
                  onInput={(e) => updateSeconds(Number(e.currentTarget.value))}
              />
              <label class="input input-bordered flex w-32 items-center gap-2">
                <input
                    type="number"
                    min="0"
                    max="3600"
                    step="1"
                    class="w-full"
                    value={seconds()}
                    onInput={(e) => updateSeconds(Number(e.currentTarget.value))}
                />
                <span class="text-sm text-base-content/60">秒</span>
              </label>
            </div>

            <div class="flex flex-wrap gap-2">
              <For each={presets}>
                {(value) => (
                    <button
                        type="button"
                        class={`btn btn-xs ${seconds() === value ? "btn-primary" : "btn-outline"}`}
                        onClick={() => updateSeconds(value)}
                    >
                      {value === 0 ? "不休眠" : `${value} 秒`}
                    </button>
                )}
              </For>
            </div>

          </div>
        </div>
        <Toast message={message()} type={type()}/>
      </>
  )
}

// 并发下载数量设置
function ConcurrencyNum(): JSXElement {
  const [num, setNum] = createSignal<number>(0);
  const {message, type, showToast} = useToast();

  onMount(async () => {
    try {
      const config: number = await GetConcurrencyNum();
      setNum(config);
    } catch (error) {
      showToast("获取并发下载数量配置失败" + error, 'error');
    }
  })

  async function saveNum(e: Event) {
    setNum(Number((e.currentTarget as HTMLInputElement).value));
    try {
      await SetConcurrencyNum(num());
      showToast(`并发下载数量设置成功: ${num()}`, 'success');
    } catch (error) {
      showToast("保存并发下载数量配置失败" + error, 'error');
    }
  }

  return (
      <div class="card bg-base-100 shadow-xl">
        <div class="card-body">
          <h2 class="card-title mb-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-accent" fill="none"
                 viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M13 10V3L4 14h7v7l9-11h-7z"/>
            </svg>
            下载控制
          </h2>
          <span>
          同时下载视频的数量 [ {num()} ]（被封了不关我事🤣）
        </span>
          <div class="w-full max-w-xl">
            {/* 滑块 */}
            <input
                type="range"
                min={1}
                max={10}
                step={1}
                value={num()}
                onInput={saveNum}
                class="range range-secondary w-full"
            />

            {/* 刻度线 */}
            <div class="flex justify-between px-2.5 mt-2 text-xs pointer-events-none select-none">
              {Array.from({length: 10}).map(() => (
                  <span>|</span>
              ))}
            </div>

            {/* 数字 */}
            <div class="flex justify-between px-2.5 mt-2 text-xs pointer-events-none select-none">
              {Array.from({length: 10}).map((_, i) => (
                  <span>{i + 1}</span>
              ))}
            </div>
          </div>
        </div>
        <Toast message={message()} type={type()}/>
      </div>
  )
}

// 自动更新设置，暂时没什么好设置的，先放这占位
// TODO
function AutoUpdate(): JSXElement {
  return (
      <></>
  )
}

// 自动下载FFmpeg设置，暂时没什么好设置的，先放这占位
// TODO
function AutoDownloadFFmpeg(): JSXElement {
  return (
      <></>
  )
}

// 主题设置
function ThemeChange(): JSXElement {
  const [theme, setTheme] = createSignal<string>('');
  const {message, type, showToast} = useToast();

  // 页面加载时读取当前主题
  onMount(async () => {
    const savedTheme: string = await GetTheme().catch(() => 'light');
    setTheme(savedTheme);
    document.documentElement.setAttribute('data-theme', savedTheme);
  });

  async function handleThemeChange(event: Event): Promise<void> {
    const target = event.target as HTMLSelectElement;
    const newTheme: string = target.value;
    setTheme(newTheme);
    document.documentElement.setAttribute('data-theme', newTheme);
    try {
      await SetTheme(newTheme);
      showToast(`已切换到 ${newTheme} 主题`, 'success');
    } catch (err) {
      showToast(String(err), 'error');
    }
  }

  return (
      <section>
        {/* 主题设置 */}
        <div class="card bg-base-100 shadow-xl">
          <div class="card-body">
            <h2 class="card-title mb-4">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-warning" fill="none"
                   viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01"/>
              </svg>
              界面主题
            </h2>
            <div class="form-control">
              <label class="label cursor-pointer justify-between">
                <span class="label-text">主题模式</span>
                <select value={theme()} onchange={handleThemeChange} class="select select-accent">
                  <option value="dark">dark - 深色模式</option>
                  <option value="light">light - 浅色模式</option>
                  <option value="cupcake">cupcake - 纸杯蛋糕</option>
                  <option value="caramellatte">caramellatte 焦糖</option>
                </select>
              </label>
              <label class="label">
                <span class="label-text-alt pl-2">当前主题：
                    <span class="text-accent font-semibold">{theme()}</span>
                </span>
              </label>
            </div>
          </div>
        </div>
        <Toast message={message()} type={type()}/>
      </section>
  )
}
