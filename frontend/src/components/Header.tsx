import type {JSXElement} from "solid-js";
import {createMemo, For} from "solid-js";
import {Link, useLocation} from "@tanstack/solid-router";
import app from "../assets/app.png";

// Simple, data-driven Header: desktop-only, flat nav with emojis
type RoutePath = '/' | '/bilibili' | '/douyin' | '/settings' | '/about';

const NAV_ITEMS: { to: RoutePath; label: string; emoji: string }[] = [
    {to: '/', label: '首页', emoji: '🏠'},
    {to: '/bilibili', label: 'B 站', emoji: '📺'},
    {to: '/douyin', label: '抖音', emoji: '🎵'},
    {to: '/settings', label: '设置', emoji: '⚙️'},
    {to: '/about', label: '关于', emoji: 'ℹ️'},
];


export default function HomeHeader(): JSXElement {
    const location = useLocation();
    const currentPath = createMemo(() => (location as any)().pathname || '/');

    const baseClass = 'px-3 py-2 rounded text-sm text-base-content';

    // 判断当前路径是否匹配导航项（支持子路由）
    const isActive = (path: string) => {
        if (path === '/') {
            return currentPath() === '/'
        }
        // 对于有子路由的路径，检查是否以该路径开头
        return currentPath() === path || currentPath().startsWith(path + '/')
    }

    return (
        <nav class="navbar sticky top-0 z-50 bg-base-100 shadow-md">
            <div class="w-full flex items-center justify-between px-4">
                {/* left: logo */}
                <Link to="/" class="btn btn-ghost normal-case text-xl text-primary flex items-center gap-3">
                    <img src={app} alt="VideoDown" class="h-8 w-8 object-contain"/>
                    <span class="font-semibold">VideoDown</span>
                </Link>

                {/* right: navigation */}
                <div class="flex items-center gap-2">
                    <For each={NAV_ITEMS}>{(item) => (
                        <Link to={item.to}
                              preload="intent"
                              class={`${baseClass} ${isActive(item.to) ? 'bg-accent text-white font-semibold' : ''}`}>
                            <span class="text-lg mr-1">{item.emoji}</span>
                            {item.label}
                        </Link>
                    )}</For>
                </div>
            </div>
        </nav>
    )
}
