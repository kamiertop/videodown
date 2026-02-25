import { createSignal, JSXElement, onMount, Show } from "solid-js";
import { invoke } from "@tauri-apps/api/core";
import SolidQR from "solid-qr";
import { onCleanup } from "solid-js";

interface BilibiliQrcodeResponse {
  code: number;
  message: string;
  ttl: number;
  data: { url: string; qrcode_key: string };
}

interface PollQrCodeDataResponse {
  url: string;
  refresh_token: string;
  timestamp: number;
  code: number;
  message: string;
}

export default function Bilibili(): JSXElement {
  const [qrUrl, setQrUrl] = createSignal<string | null>(null);
  const [_, setQrcodeKey] = createSignal<string>("");
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);
  const [polling, setPolling] = createSignal(false);
  const [loginStatus, setLoginStatus] = createSignal<string>(''); // 登录状态消息
  const [qrExpired, setQrExpired] = createSignal(false); // 二维码是否失效

  async function refreshQrcode() {
    try {
      // 清理之前的轮询
      stopPolling();
      setQrExpired(false);
      setError(null);
      setLoginStatus('');
        
      const resp = await invoke<BilibiliQrcodeResponse>("qrcode");
      setQrUrl(resp.data.url);
      setQrcodeKey(resp.data.qrcode_key);
        
      // 获取二维码后立即开始轮询
      if (resp.data.qrcode_key) {
        await startPolling(resp.data.qrcode_key);
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    }
  }

  // 轮询定时器引用
  let pollTimer: number | null = null;
  async function poll(key: string) {
    try {
      const result = await invoke<PollQrCodeDataResponse>("poll_qrcode", { qrcodeKey: key });
      setLoginStatus(result.message);
      if (result.code === 0) { // 扫码成功, 保存参数, 跳转到主页
        console.log("扫码成功, 停止轮询")
        stopPolling()
        setQrUrl(null);
      } else if (result.code === 86038) { // 二维码已失效, 需要停止轮询, 点击重新获取
        stopPolling()
        setQrExpired(true);
        setLoginStatus('二维码已失效，请刷新重新获取');
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    }
  }

  function stopPolling() {
    setPolling(false);
    if (pollTimer) {
      clearTimeout(pollTimer);
      pollTimer = null;
    }
  }
  async function startPolling(key: string): Promise<void> {
    setPolling(true);
    pollTimer = setInterval(async () => {
      await poll(key)
    }, 1000)
  }

  onMount(async () => {
    try {
      const loggedIn = await invoke<boolean>("is_logged_in");
      if (loggedIn) {
        console.log("已登录");
        return;
      }
      setLoading(true);
      const resp = await invoke<BilibiliQrcodeResponse>("qrcode");
      setQrUrl(resp.data.url);
      setQrcodeKey(resp.data.qrcode_key);

      if (resp.data.qrcode_key) {
        await startPolling(resp.data.qrcode_key);
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setLoading(false);
    }
  });

  onCleanup(() => {
    stopPolling();
  });

  return (
    <div class="h-full flex flex-col items-center justify-center">
        <Show when={error()}>
          <p class="text-sm text-red-600">{error()}</p>
        </Show>
        <Show when={loading()}>
          <p class="text-sm text-neutral-500">获取二维码中…</p>
        </Show>
        <Show when={qrUrl()}>
          <div class="flex flex-col items-center gap-4 p-6 bg-neutral-100 dark:bg-neutral-800 rounded-lg w-fit min-w-75 relative">
            {/* 二维码容器 */}
            <div class="relative">
              <SolidQR
                text={qrUrl()!}
                options={{ width: 220, margin: 2 }}
                class={`rounded border-2 transition-all duration-300 ${
                  qrExpired() 
                    ? 'border-red-500 opacity-50 grayscale' 
                    : 'border-neutral-200 dark:border-neutral-600'
                }`}
              />
              
              {/* 刷新按钮覆盖层 */}
              <Show when={qrExpired()}>
                <div class="absolute inset-0 bg-black bg-opacity-50 rounded flex items-center justify-center">
                  <button
                    onClick={refreshQrcode}
                    class="bg-white hover:bg-gray-100 text-red-500 font-bold py-2 px-4 rounded-full flex items-center gap-2 transition-colors"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                      <path fill-rule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" clip-rule="evenodd" />
                    </svg>
                    刷新二维码
                  </button>
                </div>
              </Show>
            </div>
            
            <div class="text-center">
              <p class="text-xl font-bold text-red-500 mb-2">请使用 B站 App 扫描二维码登录</p>
              <Show when={loginStatus()}>
                <p class={`text-sm mt-2 ${
                  qrExpired() ? 'text-red-500' : 'text-neutral-600 dark:text-neutral-300'
                }`}>
                  {loginStatus()}
                </p>
              </Show>
              <Show when={polling() && !qrExpired()}>
                <div class="flex items-center justify-center mt-3">
                  <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-red-500 mr-2"></div>
                  <span class="text-xs text-neutral-500">正在检测登录状态...</span>
                </div>
              </Show>
            </div>
          </div>
        </Show>
    </div>
  );
}
