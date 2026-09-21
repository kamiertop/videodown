import {createSignal, type JSXElement, onMount, Show} from "solid-js";
import Toast from "./Toast.tsx";
import {useToast} from "../hooks/useToast.ts";

/**
 * 设置页布尔开关卡片的通用骨架：onMount 读取 → 渲染 toggle → 乐观写 + 失败回滚 + toast。
 * 各设置项只提供标题、文案、图标和读写函数。
 */
export default function SettingsToggleCard(props: {
  title: string;
  /** 左侧图标，渲染在带底色的圆角容器里。 */
  icon: JSXElement;
  /** 图标容器底色，如 "bg-success/10 text-success"。 */
  iconClass: string;
  /** 开关行内的主文案。 */
  label: string;
  /** 开关行内的说明文案。 */
  description: string;
  /** 额外的警告色提示行（如“功能异常请关闭”），可选。 */
  hint?: string;
  /** toggle 配色，如 "toggle-success"。 */
  toggleClass: string;
  /** 设置名，用于失败提示文案，如“自动更新”。 */
  name: string;
  get: () => Promise<boolean>;
  set: (value: boolean) => Promise<void>;
}): JSXElement {
  const [enabled, setEnabled] = createSignal(false);
  const [loaded, setLoaded] = createSignal(false);
  const {message, type, showToast} = useToast();

  onMount(async () => {
    try {
      setEnabled(await props.get());
    } catch (e) {
      showToast(`获取${props.name}设置失败: ` + (e instanceof Error ? e.message : String(e)), "error");
    } finally {
      setLoaded(true);
    }
  });

  async function handleToggle() {
    const next = !enabled();
    setEnabled(next);
    try {
      await props.set(next);
    } catch (e) {
      setEnabled(!next);
      showToast(`保存${props.name}设置失败: ` + (e instanceof Error ? e.message : String(e)), "error");
    }
  }

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
              <div class="flex items-start gap-3">
                <div class={`mt-1 flex h-9 w-9 shrink-0 items-center justify-center rounded-md ${props.iconClass}`}>
                  {props.icon}
                </div>
                <div class="min-w-0 flex-1">
                  <h2 class="text-lg font-semibold leading-8">{props.title}</h2>
                  <div class="mt-4 flex items-center justify-between gap-6 rounded-md border border-base-300 bg-base-200/40 px-4 py-3">
                    <div>
                      <p class="font-medium leading-6">{props.label}</p>
                      <p class="text-sm leading-6 text-base-content/60">{props.description}</p>
                      <Show when={props.hint}>
                        <p class="mt-1 text-sm leading-6 text-warning">{props.hint}</p>
                      </Show>
                    </div>
                    <input type="checkbox" class={`toggle ${props.toggleClass}`} checked={enabled()} onChange={handleToggle}/>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </Show>
        <Toast message={message()} type={type()}/>
      </>
  );
}
