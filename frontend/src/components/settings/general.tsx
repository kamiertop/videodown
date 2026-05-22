import {createSignal, type JSXElement, onMount} from "solid-js";
import {GetStorage, GetTheme, SetStorage, SetTheme} from "../../../wailsjs/go/utils/Settings";
import {useToast} from "../../hooks/useToast.ts";
import Toast from "../Toast";

export function GeneralSection(): JSXElement {
  return (
      <div class="space-y-6 max-w-2xl mx-auto">
        <StorageDirectory/>
        <ThemeChange/>
      </div>
  )
}

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

function ThemeChange(): JSXElement {
  const [theme, setTheme] = createSignal<string>('');
  const {message, type, showToast} = useToast();

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
        <Toast message={message()} type={type()}/>
      </div>
  )
}
