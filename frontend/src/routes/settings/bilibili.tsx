import {batch, createSignal, For, type JSXElement, onMount, Show} from "solid-js";
import {createFileRoute} from '@tanstack/solid-router';
import {
  GetParsePlayURLNum,
  GetParsePlayURLSleep,
  SetParsePlayURLNum,
  SetParsePlayURLSleep,
} from "@bindings/github.com/kamiertop/videodown/utils/settings";
import {useToast} from "../../hooks/useToast";
import Toast from "../../components/Toast";

export const Route = createFileRoute('/settings/bilibili')({component: BilibiliSection});


function BilibiliSection(): JSXElement {
  const {message, type, showToast} = useToast();
  const [num, setNum] = createSignal<number>(5);
  const [sleep, setSleep] = createSignal<number>(0);
  const [loaded, setLoaded] = createSignal(false);

  onMount(async () => {
    try {
      const [n, s] = await Promise.all([
        GetParsePlayURLNum().catch(() => 5),
        GetParsePlayURLSleep().catch(() => 0),
      ]);
      batch(() => {
        setNum(n);
        setSleep(s);
      });
    } catch {
      setNum(5);
      setSleep(0);
    } finally {
      setLoaded(true);
    }
  });

  async function saveNum(value: number) {
    setNum(value);
    try {
      await SetParsePlayURLNum(value);
      showToast("保存成功", "success");
    } catch (e) {
      showToast("保存失败: " + (e instanceof Error ? e.message : String(e)), "error");
    }
  }

  async function saveSleep(value: number) {
    setSleep(value);
    try {
      await SetParsePlayURLSleep(value);
    } catch (e) {
      showToast("保存失败: " + (e instanceof Error ? e.message : String(e)), "error");
    }
  }

  return (
      <>
        <Show when={loaded()} fallback={<div class="space-y-6 max-w-2xl mx-auto"><div class="card bg-base-100 shadow-xl"><div class="card-body"><span class="loading loading-dots loading-sm" aria-label="读取中"/></div></div></div>}>
        <div class="space-y-6 max-w-2xl mx-auto">
          {/* 并发数 */}
          <div class="card bg-base-100 shadow-xl">
            <div class="card-body">
              <h2 class="card-title mb-2">同时解析数 [ {num()} ]</h2>
              <p class="text-sm text-base-content/60 mb-4">
                批量下载时并发请求播放地址的数量。越小越不容易触发风控，越大解析越快。
              </p>
              <div class="w-full max-w-xl">
                <input
                    type="range"
                    min={1}
                    max={5}
                    step={1}
                    value={num()}
                    onInput={(e) => setNum(Number(e.currentTarget.value))}
                    onChange={(e) => saveNum(Number(e.currentTarget.value))}
                    class="range range-secondary w-full"
                />
                <div class="flex justify-between px-2.5 mt-2 text-xs pointer-events-none select-none">
                  {Array.from({length: 5}).map((_, i) => (
                      <span>{i + 1}</span>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* 请求间隔随机休眠 */}
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
                    请求间隔随机休眠
                  </h2>
                  <p class="text-sm text-base-content/70">
                    每个播放地址解析任务请求前随机等待 <span
                      class="text-secondary font-bold">[0-{sleep()}]</span> 秒，打散批量解析节奏，降低风控概率

                  </p>
                </div>
                <div class="badge badge-primary badge-outline shrink-0">
                  {sleep()} 秒
                </div>
              </div>
              <div class="flex items-center gap-3">
                <input
                    type="range"
                    min="0"
                    max="60" step="1"
                    value={sleep()}
                    onInput={(e) => setSleep(Number(e.currentTarget.value))}
                    onChange={(e) => saveSleep(Number(e.currentTarget.value))}
                    class="range range-primary flex-1"
                />
              </div>
              <div class="flex flex-wrap gap-2">
                <For each={[0, 1, 2, 3, 5, 10, 15, 20, 30, 60]}>
                  {(v) => (
                      <button
                          type="button"
                          class={`btn btn-xs ${sleep() === v ? "btn-primary" : "btn-outline"}`}
                          onClick={() => saveSleep(v)}
                      >
                        {v === 0 ? "不休眠" : `${v} 秒`}
                      </button>
                  )}
                </For>
              </div>
            </div>
          </div>
        </div>
        </Show>
        <Toast message={message()} type={type()}/>
      </>
  )
}
