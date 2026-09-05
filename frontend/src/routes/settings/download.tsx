import {
  GetBulkDownloadSleepTime,
  GetConcurrencyNum,
  GetSleepTime,
  SetBulkDownloadSleepTime,
  SetConcurrencyNum,
  SetSleepTime,
} from "@bindings/github.com/kamiertop/videodown/utils/settings";
import {createFileRoute} from '@tanstack/solid-router';
import {createSignal, For, type JSXElement, onCleanup, onMount, Show} from "solid-js";
import Toast from "../../components/Toast";
import {useToast} from "../../hooks/useToast";

export const Route = createFileRoute('/settings/download')({component: DownloadSection});

function DownloadSection(): JSXElement {
  return (
      <div class="space-y-6 max-w-2xl mx-auto">
        <ConcurrencyNum/>
        <SleepAfterDownLoad/>
        <BulkDownloadSleep/>
      </div>
  )
}

function BulkDownloadSleep(): JSXElement {
  const [seconds, setSeconds] = createSignal(0);
  const [loaded, setLoaded] = createSignal(false);
  const {message, type, showToast} = useToast();
  let timer: number | undefined;
  const normalize = (value: number) => Math.max(0, Math.min(600, Math.round(Number.isFinite(value) ? value : 0)));
  const update = (value: number) => {
    const next = normalize(value);
    setSeconds(next);
    if (timer !== undefined) window.clearTimeout(timer);
    timer = window.setTimeout(() => void SetBulkDownloadSleepTime(next).catch((e) => showToast(`保存一键下载休眠时间失败${e}`, "error")), 300);
  };
  onMount(async () => {
    try {
      setSeconds(normalize(Number(await GetBulkDownloadSleepTime())));
    } catch (e) {
      showToast(`获取一键下载休眠时间失败${e}`, "error");
    } finally {
      setLoaded(true);
    }
  });
  onCleanup(() => {
    if (timer !== undefined) window.clearTimeout(timer);
  });
  return (
      <Show when={loaded()}
            fallback={
              <div class="card bg-base-100 shadow-xl">
                <div class="card-body">
                  <span class="loading loading-dots loading-sm"/>
                </div>
              </div>
            }
      >
        <div class="card bg-base-100 shadow-xl">
          <div class="card-body">
            <div class="flex items-start justify-between gap-4">
              <h2 class="card-title">一键下载全部功能 分页间隔休眠</h2>
              <div class="badge badge-outline shrink-0">{seconds()} 秒</div>
            </div>
            <p class="text-sm text-base-content/70">
              使用一键下载功能，自动分页加载所有视频时，每次请求之间随机休眠 0-{seconds()} 秒，降低触发限流的概率
            </p>
            <div class="flex items-center gap-3">
              <input type="range" min="0" max="600" step="5"
                     value={seconds()}
                     class="range range-secondary flex-1"
                     onInput={(e) => update(Number(e.currentTarget.value))}/>
              <label class="input input-bordered flex w-32 items-center gap-2">
                <input type="number" min="0" max="600"
                       value={seconds()} class="w-full"
                       onInput={(e) => update(Number(e.currentTarget.value))}/>
                <span class="text-sm text-base-content/60">秒</span>
              </label>
            </div>
          </div>
        </div>
        <Toast message={message()} type={type()}/>
      </Show>
  );
}

function ConcurrencyNum(): JSXElement {
  const [num, setNum] = createSignal<number>(0);
  const [loaded, setLoaded] = createSignal(false);
  const {message, type, showToast} = useToast();

  onMount(async () => {
    try {
      const config: number = await GetConcurrencyNum();
      setNum(config);
    } catch (error) {
      showToast("获取并发下载数量配置失败" + error, 'error');
    } finally {
      setLoaded(true);
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
      <Show when={loaded()} fallback={<div class="card bg-base-100 shadow-xl">
        <div class="card-body">
          <span class="loading loading-dots loading-sm" aria-label="读取中"/>
        </div>
      </div>}>
        <div class="card bg-base-100 shadow-xl">
          <div class="card-body">
            <h2 class="card-title mb-2">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-accent" fill="none"
                   viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M13 10V3L4 14h7v7l9-11h-7z"/>
              </svg>
              并发数量
            </h2>
            <span class="text-sm text-base-content/70">
            同时下载视频的数量 [ {num()} ]
          </span>
            <div class="w-full max-w-xl">
              <input
                  type="range"
                  min={1}
                  max={10}
                  step={1}
                  value={num()}
                  onInput={saveNum}
                  class="range range-secondary w-full"
              />
              <div class="flex justify-between px-2.5 mt-2 text-xs pointer-events-none select-none">
                {Array.from({length: 10}).map(() => (
                    <span>|</span>
                ))}
              </div>
              <div class="flex justify-between px-2.5 mt-2 text-xs pointer-events-none select-none">
                {Array.from({length: 10}).map((_, i) => (
                    <span>{i + 1}</span>
                ))}
              </div>
            </div>
          </div>
          <Toast message={message()} type={type()}/>
        </div>
      </Show>
  )
}

function SleepAfterDownLoad(): JSXElement {
  const [seconds, setSeconds] = createSignal<number>(0);
  const [loaded, setLoaded] = createSignal(false);
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
    } finally {
      setLoaded(true);
    }
  })

  onCleanup(() => {
    if (saveTimer !== undefined) {
      window.clearTimeout(saveTimer);
    }
  });

  return (
      <>
        <Show when={loaded()}
              fallback={
                <div class="card bg-base-100 shadow-xl">
                  <div class="card-body">
                    <span class="loading loading-dots loading-sm" aria-label="读取中"/>
                  </div>
                </div>
              }>
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
        </Show>
        <Toast message={message()} type={type()}/>
      </>
  )
}
