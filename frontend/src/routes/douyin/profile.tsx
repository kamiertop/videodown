import {createFileRoute} from '@tanstack/solid-router'
import {createSignal, type JSXElement, onMount} from "solid-js";
import {GetCookie, SetCookie} from "../../../wailsjs/go/api/Douyin";
import Toast from "../../components/Toast.tsx";
import {useToast} from "../../hooks/useToast.ts";

export const Route = createFileRoute('/douyin/profile')({
  component: RouteComponent,
})

function RouteComponent(): JSXElement {
  const [cookie, setCookie] = createSignal<string>('');
  const {message, type, showToast} = useToast();
  onMount(async () => {
    try {
      const c: string = await GetCookie();
      if (c == "") {
        showToast('Cookie为空，请设置Cookie', 'warning');
      } else {
        showToast('Cookie加载成功', 'success');
        setCookie(c);
      }
    } catch (error) {
      const errMsg: string = error as string;
      if (errMsg.includes("未设置")) {
        showToast(errMsg, 'warning');
      } else {
        showToast(errMsg, 'error');
      }
    }
  })

  async function saveCookie(c: string): Promise<void> {
    try {
      setCookie(c)
      await SetCookie(c);
    } catch (e) {
      showToast("保存Cookie失败：" + e, 'error');
    }
  }

  return (
    <section class="h-full overflow-hidden bg-base-200/40 px-4 py-4 md:px-6 md:py-5">
      <header class="mb-3 rounded-lg border border-base-300 bg-base-100 px-4 py-3">
        <h2 class="text-base font-bold">账号设置</h2>
        <p class="text-sm text-base-content/60">把浏览器里的抖音 Cookie 粘贴到这里，页面会直接复用这份登录态访问接口。</p>
      </header>

      <div class="h-[calc(100%-5.25rem)] overflow-auto rounded-lg border border-base-300 bg-base-100 p-5">
        <div class="mx-auto flex max-w-5xl flex-col gap-4">
          <div class="rounded-lg bg-base-200/80 px-4 py-3 text-sm text-base-content/65">
            建议直接复制浏览器完整 Cookie，至少包含 `sessionid`、`ttwid`、`UIFID`、`s_v_web_id`。
          </div>
          <textarea
            class="textarea min-h-[28rem] w-full border-base-300 bg-base-100 font-mono text-xs leading-6"
            value={cookie()}
            onInput={(e) => void saveCookie(e.currentTarget.value)}
            placeholder="粘贴抖音 Cookie 至此"
          />
        </div>
      </div>
      <Toast message={message()} type={type()}></Toast>
    </section>
  )
}
