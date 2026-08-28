import {SetStorage} from "@bindings/github.com/kamiertop/videodown/internal/app/controller";
import {
  GetCloseToTray,
  GetStorage,
  GetTheme,
  HasCloseToTrayChoice,
  SetCloseToTray,
  SetTheme
} from "@bindings/github.com/kamiertop/videodown/utils/settings";
import {IsAutoUpdate, SetAutoUpdate} from "@bindings/github.com/kamiertop/videodown/internal/updater/updater";
import {createSignal, type JSXElement, onMount} from "solid-js";
import {useToast} from "../../hooks/useToast.ts";
import Toast from "../Toast";

export function GeneralSection(): JSXElement {
  return (
      <div class="space-y-6 max-w-2xl mx-auto">
        <StorageDirectory/>
        <ThemeChange/>
        <AutoUpdate/>
        <CloseToTray/>
      </div>
  )
}

function AutoUpdate(): JSXElement {
  const [enabled, setEnabled] = createSignal(false);
  const [loaded, setLoaded] = createSignal(false);
  const {message, type, showToast} = useToast();

  onMount(async () => {
    try {
      setEnabled(await IsAutoUpdate());
    } catch (e) {
      showToast("获取自动更新设置失败: " + (e instanceof Error ? e.message : String(e)), "error");
    } finally {
      setLoaded(true);
    }
  });

  async function handleToggle() {
    const next = !enabled();
    setEnabled(next);
    try {
      await SetAutoUpdate(next);
    } catch (e) {
      setEnabled(!next);
      showToast("保存自动更新设置失败: " + (e instanceof Error ? e.message : String(e)), "error");
    }
  }

  return (
    <>
      <div class="card bg-base-100 shadow-xl" classList={{"invisible": !loaded()}}>
        <div class="card-body">
          <div class="flex items-start gap-3">
            <div class="mt-1 flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-success/10 text-success">↻</div>
            <div class="min-w-0 flex-1">
              <h2 class="text-lg font-semibold leading-8">自动更新</h2>
              <div class="mt-4 flex items-center justify-between gap-6 rounded-md border border-base-300 bg-base-200/40 px-4 py-3">
                <div>
                  <p class="font-medium leading-6">启动时检查新版本</p>
                  <p class="text-sm leading-6 text-base-content/60">在后台检查 GitHub 是否有可用更新</p>
                </div>
                <input type="checkbox" class="toggle toggle-success" checked={enabled()} disabled={!loaded()} onChange={handleToggle}/>
              </div>
            </div>
          </div>
        </div>
      </div>
      <Toast message={message()} type={type()}/>
    </>
  );
}

function CloseToTray(): JSXElement {
  const [enabled, setEnabled] = createSignal(false);
  const [configured, setConfigured] = createSignal(true);
  const [loaded, setLoaded] = createSignal(false);
  const {message, type, showToast} = useToast();

  onMount(async () => {
    try {
      const hasChoice = await HasCloseToTrayChoice();
      setConfigured(hasChoice);
      if (!hasChoice) {
        setEnabled(true);
        return;
      }
      const val = await GetCloseToTray();
      setEnabled(val);
    } catch {
      // key 不存在 = 用户从未选择 = 默认隐藏到托盘
      setEnabled(true);
    } finally {
      setLoaded(true);
    }
  });

  async function handleToggle() {
    const next = !enabled();
    setEnabled(next);
    try {
      await SetCloseToTray(next);
      setConfigured(true);
    } catch (e) {
      setEnabled(!next);
      showToast("保存失败: " + (e instanceof Error ? e.message : String(e)), "error");
    }
  }

  async function saveCurrentChoice() {
    try {
      await SetCloseToTray(enabled());
      setConfigured(true);
      showToast("保存成功", "success");
    } catch (e) {
      showToast("保存失败: " + (e instanceof Error ? e.message : String(e)), "error");
    }
  }

  return (
      <>
        <div class="card bg-base-100 shadow-xl" classList={{"invisible": !loaded()}}>
          <div class="card-body">
            <div class="flex items-start gap-3">
              <div class="mt-1 flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-info/10 text-info">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none"
                     viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z"/>
                </svg>
              </div>
              <div class="min-w-0 flex-1">
                <h2 class="text-lg font-semibold leading-8">窗口关闭行为</h2>
                <div
                    class="mt-4 flex items-center justify-between gap-6 rounded-md border border-base-300 bg-base-200/40 px-4 py-3">
                  <div class="min-w-0">
                    <p class="font-medium leading-6">关闭时缩小到任务栏托盘</p>
                    <p class="text-sm leading-6 text-base-content/60">
                      关闭时隐藏窗口到系统托盘，下载不受影响
                    </p>
                    {!configured() && (
                        <p class="text-sm leading-6 text-warning">
                          当前未设置，首次关闭窗口时仍会询问
                        </p>
                    )}
                  </div>
                  <div class="flex h-8 w-16 shrink-0 items-center justify-end">
                    <input
                        type="checkbox"
                        class="toggle toggle-info"
                        classList={{"invisible": !loaded()}}
                        checked={enabled()}
                        disabled={!loaded()}
                        onChange={handleToggle}
                    />
                  </div>
                </div>
                {!configured() && (
                    <button class="btn btn-sm btn-outline btn-warning mt-3" onClick={saveCurrentChoice}>
                      保存当前选择
                    </button>
                )}
              </div>
            </div>
          </div>
        </div>
        <Toast message={message()} type={type()}/>
      </>
  )
}

function StorageDirectory(): JSXElement {
  const [storagePath, setStoragePath] = createSignal<string>("");
  const [loaded, setLoaded] = createSignal(false);
  const {message, type, showToast} = useToast();

  onMount(async () => {
    try {
      const path = await GetStorage();
      setStoragePath(path);
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : '获取存储路径失败';
      showToast(errorMsg, 'error');
    } finally {
      setLoaded(true);
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
                    classList={{"invisible": !loaded()}}
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

// 界面主题
function ThemeChange(): JSXElement {
  const [theme, setTheme] = createSignal<string>('');
  const [loaded, setLoaded] = createSignal(false);
  const {message, type, showToast} = useToast();

  onMount(async () => {
    const savedTheme: string = await GetTheme().catch(() => 'light');
    setTheme(savedTheme);
    document.documentElement.setAttribute('data-theme', savedTheme);
    setLoaded(true);
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
              <select value={theme()} onchange={handleThemeChange} class="select select-accent"
                      classList={{"invisible": !loaded()}}>
                <option value="dark">dark - 深色模式</option>
                <option value="light">light - 浅色模式</option>
                <option value="cupcake">cupcake - 纸杯蛋糕</option>
                <option value="caramellatte">caramellatte 焦糖</option>
              </select>
            </label>
            <label class="label">
              <span class="label-text-alt pl-2">当前主题：
                <span class="text-accent font-semibold" classList={{"invisible": !loaded()}}>{theme()}</span>
              </span>
            </label>
          </div>
        </div>
        <Toast message={message()} type={type()}/>
      </div>
  )
}
