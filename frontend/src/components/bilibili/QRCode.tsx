import QRCodeGenerator from "qrcode";
import {createSignal, type JSXElement, onMount, Show} from "solid-js";
import {QRCode as GetQRCode} from "../../../wailsjs/go/api/BiliBili";
import type {model} from "../../../wailsjs/go/models.ts";
import {useToast} from "../../hooks/useToast";
import Toast from "../Toast";

interface BiliBiliQRCodeProps {
  expired?: boolean;
  onLoad?: (data: model.QRCodeData) => void;
}

interface QRCodeViewData {
  image: string;
  payload: model.QRCodeData;
}

/**
 * Bilibili 登录二维码组件：拉取登录二维码并渲染成图片。
 */
export default function BiliBiliQRCode(props: BiliBiliQRCodeProps): JSXElement {
  const [loadingQRCode, setLoadingQRCode] = createSignal(false);
  const [qrCode, setQRCode] = createSignal<QRCodeViewData | null>(null);
  const {message, type, showToast} = useToast();

  const loadQRCode = async () => {
    setLoadingQRCode(true);

    try {
      const payload = await GetQRCode();
      props.onLoad?.(payload);

      if (!payload?.url?.trim()) {
        showToast('获取二维码失败', 'error');
      }

      const image = await QRCodeGenerator.toDataURL(payload.url, {
        width: 240,
        margin: 1,
      });

      setQRCode({image, payload});
    } catch (error) {
      setQRCode(null);
      showToast(error instanceof Error ? error.message : String(error), 'error');
    } finally {
      setLoadingQRCode(false);
    }
  }

  onMount(() => {
    void loadQRCode();
  })

  return (
      <>
        <div class="card border border-base-300 bg-base-100 shadow-2xl">
          <div class="card-body items-center gap-4 p-6 text-center">
            <div class="rounded-3xl bg-base-200 p-3">
              <Show
                  when={qrCode()}
                  fallback={
                    <div class="flex h-56 w-56 items-center justify-center">
                      <span class="loading loading-spinner loading-lg text-primary"></span>
                    </div>
                  }
              >
                <div class="relative">
                  <img
                      src={qrCode()!.image}
                      alt="Bilibili 登录二维码"
                      class="h-56 w-56 rounded-2xl border bg-white p-2 transition"
                      classList={{"opacity-40 grayscale": props.expired, "border-base-300": !props.expired}}
                  />
                  <Show when={props.expired}>
                    <button
                        class="btn btn-primary absolute inset-x-6 top-1/2 -translate-y-1/2"
                        onClick={() => void loadQRCode()}
                    >
                      刷新二维码
                    </button>
                  </Show>
                </div>
              </Show>
            </div>

            <div class="space-y-1.5">
              <h2 class="text-lg font-semibold">请使用 B 站 App 扫码</h2>
              <p class="text-sm text-base-content/60">扫码后在手机上确认登录，页面会自动更新状态。</p>
            </div>

            <div class="flex gap-3">
              <button class="btn btn-outline" onClick={() => void loadQRCode()} disabled={loadingQRCode()}>
                {loadingQRCode() ? '获取中...' : '重新获取'}
              </button>
            </div>
          </div>
        </div>
        <Toast message={message()} type={type()}/>
      </>
  )
}
