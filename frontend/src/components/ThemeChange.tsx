import {createSignal, type JSXElement, onMount} from "solid-js";


/**
 * 主题切换组件，允许用户切换主题
 */

export default function ThemeChange(): JSXElement {
    const [theme, setTheme] = createSignal<string>('light');

    // 页面加载时读取当前主题
    onMount(() => {
        const savedTheme = localStorage.getItem('theme') || 'light';
        setTheme(savedTheme);
    });

    function handleThemeChange(event: Event): void {
        const target = event.target as HTMLSelectElement;
        const newTheme: string = target.value;
        setTheme(newTheme);
        document.documentElement.setAttribute('data-theme', newTheme);
    }

    return (
        <>
            {/* 主题设置 */}
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
                                <option disabled={true}>选择一个主题</option>
                                <option value="dark">dark - 深色模式</option>
                                <option value="light">light - 浅色模式</option>
                                <option value="cupcake">cupcake - 纸杯蛋糕</option>
                            </select>
                        </label>
                        <label class="label">
                            <span class="label-text-alt">当前主题：<span
                                class="text-accent font-semibold">{theme()}</span></span>
                        </label>
                    </div>
                </div>
            </div>
        </>
    )
}
