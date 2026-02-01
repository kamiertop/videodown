import {For, JSXElement} from "solid-js";
import {A} from "@solidjs/router";

export default function Header():JSXElement {
    interface Link {
        name: string;
        path: string;
        icon: string;
    }

    const list: Link[] = [
        {
            name: "首页",
            path: "/",
            icon: "🏠"
        },
        {
            name: "设置",
            path: "/settings",
            icon: "⚙️"
        },
        {
            name: "关于",
            path: "/about",
            icon: "ℹ️"
        }
    ];

    return (
        // 优化 Header 整体样式：固定顶部、增加内边距、优化对齐、提升层级
        <header class="flex flex-row justify-between items-center px-4 py-3 text-3xl backdrop-blur-md bg-white/20 shadow-lg sticky top-0 z-50 w-full">
            <h1 class="text-2xl font-bold text-gray-800 dark:text-gray-100 drop-shadow-sm">
                视频下载器
            </h1>
            {/* 优化导航容器：调整间距、垂直居中 */}
            <nav class="flex items-center space-x-2">
                <For each={list}>
                    {(item) => (
                        <A
                            href={item.path}
                            // 优化链接默认样式：调整内边距、优化颜色、提升磨砂背景质感、圆角更柔和
                            class="inline-flex items-center px-5 py-2.5 text-sm text-gray-700 dark:text-gray-300 bg-white/40 dark:bg-gray-800/40 hover:bg-white/60 dark:hover:bg-gray-800/60 hover:text-blue-600 dark:hover:text-blue-400 transition-all duration-200 rounded-xl"
                            activeClass="font-extrabold shadow-xl bg-white/80 dark:bg-gray-900/70 text-blue-700 dark:text-blue-300 ring-2 ring-blue-300 dark:ring-blue-500"
                            end={item.path === "/"}
                        >
                            <span class="mr-2">{item.icon}</span>{item.name}
                        </A>
                    )}
                </For>
            </nav>
        </header>
    )
}