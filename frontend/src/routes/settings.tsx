import {createFileRoute} from '@tanstack/solid-router'
import {createSignal, For, type JSXElement, onCleanup, onMount} from "solid-js";
import {SetStorage} from "../../wailsjs/go/main/App";
import {
  GetConcurrencyNum,
  GetSavePreference,
  GetSleepTime,
  GetStorage,
  SetConcurrencyNum,
  SetSleepTime
} from "../../wailsjs/go/utils/Settings";
import ThemeChange from "../components/ThemeChange.tsx";
import Toast from "../components/Toast";
import {useToast} from "../hooks/useToast";

// @ts-ignore
export const Route = createFileRoute('/settings')({
  component: SettingsComponent,
})


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
      <ConcurrencyNum/>
      <SleepAfterDownLoad/>
      <SavePreferences/>
      <ThemeChange/>
    </div>
  )
}

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
          同时下载视频的数量（被封了不关我事🤣）
        </span>

        <div class="w-full max-w-xs">
          <input type="range" min={1} max="5" value={num()}
                 onInput={(e) => saveNum(e)}
                 class="range range-secondary" step="1"/>
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
      <Toast message={message()} type={type()}/>
    </div>
  )
}
