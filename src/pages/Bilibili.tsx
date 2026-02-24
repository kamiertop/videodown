import { createSignal, JSXElement, onMount, Show } from "solid-js";
import { invoke } from "@tauri-apps/api/core";
import SolidQR from "solid-qr";

interface BilibiliQrcodeResponse {
  code: number;
  message: string;
  ttl: number;
  data: { url: string; qrcode_key: string };
}

export default function Bilibili(): JSXElement {
  const [qrUrl, setQrUrl] = createSignal<string | null>(null);
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);
  const [showContent, setShowContent] = createSignal(true); // 未登录才为 true

  onMount(async () => {
    try {
      const loggedIn = await invoke<boolean>("is_logged_in");
      if (loggedIn) {
        setShowContent(false);
        return;
      }
      setLoading(true);
      const res = await invoke<BilibiliQrcodeResponse>("qrcode");
      setQrUrl(res.data.url);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setLoading(false);
    }
  });

  if (!showContent()) return null;

  return (
    <div class="h-full flex flex-col items-center justify-center">
        <Show when={error()}>
          <p class="text-sm text-red-600">{error()}</p>
        </Show>
        <Show when={loading()}>
          <p class="text-sm text-neutral-500">获取二维码中…</p>
        </Show>
        <Show when={qrUrl()}>
          <div class="flex flex-col items-center gap-2 p-4 bg-neutral-100 dark:bg-neutral-800 rounded-lg w-fit">
            <SolidQR
              text={qrUrl()!}
              options={{ width: 220, margin: 2 }}
              class="rounded border border-neutral-200 dark:border-neutral-600"
            />
            <p class="text-xl font-bold text-red-500">请使用 B站 App 扫描二维码登录</p>
          </div>
        </Show>
    </div>
  );
}
