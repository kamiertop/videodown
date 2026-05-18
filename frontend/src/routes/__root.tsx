import {createRootRoute, Outlet} from '@tanstack/solid-router'
import {type JSXElement, onMount} from "solid-js";
import {GetTheme} from "../../wailsjs/go/utils/Settings";
import HomeHeader from "../components/Header.tsx";

export const Route = createRootRoute({
  component: RootComponent,
})

// 根组件，负责布局，顶部HomeHeader，内容区Outlet
function RootComponent(): JSXElement {
  // 挂载时获取主题设置并应用到 documentElement 上，供全局使用
  onMount(async () => {
    const theme: string = await GetTheme().catch(() => 'light');
    document.documentElement.setAttribute('data-theme', theme);
  });
  return (
      <div class="h-screen bg-base-200 flex flex-col">
        <HomeHeader/>
        <div class="flex-1 min-h-0 ">
          <Outlet/>
        </div>
      </div>
  )
}
