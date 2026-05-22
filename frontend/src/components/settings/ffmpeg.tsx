import {createSignal, type JSXElement, onMount, Show} from "solid-js";
import {FFmpegPath, HasFFmpeg, SelectFFmpegPath} from "../../../wailsjs/go/utils/Settings";
import {BrowserOpenURL} from "../../../wailsjs/runtime";
import {useToast} from "../../hooks/useToast";
import Toast from "../Toast";

export function FFmpegSection(): JSXElement {
  const [detected, setDetected] = createSignal<boolean | null>(null);
  const [path, setPath] = createSignal("");
  const [checking, setChecking] = createSignal(false);
  const [customPath, setCustomPath] = createSignal("");
  const {message, type, showToast} = useToast();

  async function check() {
    setChecking(true);
    try {
      const [has, p] = await Promise.all([HasFFmpeg(), FFmpegPath().catch(() => "")]);
      setDetected(has);
      setPath(p);
    } catch {
      setDetected(false);
      setPath("");
    } finally {
      setChecking(false);
    }
  }

  onMount(() => {
    void check();
  });

  async function handleRecheck() {
    await check();
    if (detected()) {
      showToast("FFmpeg 已就绪", "success");
    } else {
      showToast("未检测到 FFmpeg", "warning");
    }
  }

  async function handleBrowse() {
    try {
      const selected = await SelectFFmpegPath();
      if (selected) {
        setCustomPath(selected);
        showToast("FFmpeg 路径设置成功", "success");
        void check();
      }
    } catch (e) {
      showToast(e instanceof Error ? e.message : "选择失败", "error");
    }
  }

  return (
      <>
        <div class="space-y-6 max-w-2xl mx-auto">
          {/* 状态卡片 */}
          <div class="card bg-base-100 shadow-xl">
            <div class="card-body">
              <h2 class="card-title mb-4">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-accent" fill="none"
                     viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                </svg>
                FFmpeg 状态
              </h2>

              <Show when={checking()}>
                <div class="flex items-center gap-3 text-sm text-base-content/60">
                  <span class="loading loading-spinner loading-sm"/>
                  检测中...
                </div>
              </Show>

              <Show when={!checking() && detected() != null}>
                <div class="space-y-3">
                  <div class="flex items-center gap-3">
                    <Show when={detected()}
                          fallback={<span class="badge badge-error gap-1">未检测到</span>}
                    >
                      <span class="badge badge-success gap-1">已就绪</span>
                    </Show>
                    <span class="text-sm text-base-content/70">
                      {detected() ? "FFmpeg 可正常使用" : "请安装或手动指定 FFmpeg 路径"}
                    </span>
                  </div>

                  <Show when={path()}>
                    <div class="bg-base-200 rounded-lg p-3">
                      <div class="text-xs text-base-content/50 mb-1">当前路径</div>
                      <code class="text-sm break-all">{path()}</code>
                    </div>
                  </Show>

                  <button class="btn btn-sm btn-outline" onClick={handleRecheck}>
                    重新检测
                  </button>
                </div>
              </Show>
            </div>
          </div>

          {/* 自定义路径 */}
          <div class="card bg-base-100 shadow-xl">
            <div class="card-body">
              <h2 class="card-title mb-4">自定义路径</h2>
              <p class="text-sm text-base-content/60 mb-4">
                自动检测不到时，可在此手动指定 FFmpeg 的完整路径，
                也可将 ffmpeg 放到应用同目录或 <code class="text-xs bg-base-200 px-1 rounded">bin/</code> 子目录下自动识别。
              </p>
              <div class="space-y-3">
                <div class="flex gap-2">
                  <input
                      type="text"
                      value={customPath()}
                      onInput={(e) => setCustomPath(e.currentTarget.value)}
                      placeholder="例如 /usr/bin/ffmpeg 或 C:\ffmpeg\bin\ffmpeg.exe"
                      class="input input-bordered flex-1 text-sm"
                  />
                  <button
                      class="btn btn-primary"
                      onClick={handleBrowse}
                  >
                    浏览
                  </button>
                </div>
              </div>
            </div>
          </div>

          {/* 安装说明 */}
          <div class="card bg-base-100 shadow-xl">
            <div class="card-body">
              <h2 class="card-title mb-4">安装指南</h2>
              <div class="space-y-4 text-sm text-base-content/70">
                <div>
                  <h3 class="font-semibold text-base-content mb-1">Windows</h3>
                  <p>
                    访问 <span class="link link-primary cursor-pointer"
                              onClick={() => BrowserOpenURL("https://ffmpeg.org/download.html")}>ffmpeg.org</span> 下载，
                    将 ffmpeg.exe 放到应用同目录或 <code class="text-xs bg-base-200 px-1 rounded">bin/</code> 子目录下，
                    或添加到系统 PATH 环境变量。
                  </p>
                </div>
                <div>
                  <h3 class="font-semibold text-base-content mb-1">macOS</h3>
                  <code class="text-xs bg-base-200 px-1 rounded">brew install ffmpeg</code>
                </div>
                <div>
                  <h3 class="font-semibold text-base-content mb-1">Linux</h3>
                  <code class="text-xs bg-base-200 px-1 rounded">sudo apt install ffmpeg</code>
                  &nbsp;或&nbsp;
                  <code class="text-xs bg-base-200 px-1 rounded">sudo dnf install ffmpeg</code>
                </div>
              </div>
            </div>
          </div>
        </div>
        <Toast message={message()} type={type()}/>
      </>
  )
}
