import type {JSXElement} from "solid-js";
import {createSignal, For, onCleanup, onMount} from "solid-js";
import {Link} from "@tanstack/solid-router";
import {MyInfo} from "../../../wailsjs/go/api/BiliBili";
import {getLoggedInDeduped} from "../../lib/bilibiliAuth.ts";
import bilibiliAvatarFallback from "../../assets/bilibili_256_256.svg";

const menuItems = [
    {
        name: "UP主",
        link: "/bilibili/up",
        icon: <svg x="1775035354017" class="icon" viewBox="0 0 1024 1024"
                   xmlns="http://www.w3.org/2000/svg" p-id="15993" width="18" height="18">
            <path
                d="M262.4 351.914667a32 32 0 0 1 32 32v161.109333a62.890667 62.890667 0 0 0 125.781333 0V383.914667a32 32 0 0 1 64 0v161.109333a126.890667 126.890667 0 1 1-253.781333 0V383.914667a32 32 0 0 1 32-32zM539.818667 383.914667a32 32 0 0 1 32-32h99.114666a122.666667 122.666667 0 1 1 0 245.333333h-67.114666v42.666667a32 32 0 0 1-64 0v-256z m64 149.333333h67.114666a58.666667 58.666667 0 1 0 0-117.333333h-67.114666v117.333333z"
                fill="currentColor" p-id="15994"></path>
            <path
                d="M512 213.248c-107.434667 0-201.258667 5.461333-267.52 10.837333a92.672 92.672 0 0 0-85.76 84.48 2301.738667 2301.738667 0 0 0-9.386667 203.349334c0 77.952 4.522667 149.248 9.386667 203.349333a92.672 92.672 0 0 0 85.76 84.48 3354.026667 3354.026667 0 0 0 267.52 10.837333c107.434667 0 201.258667-5.461333 267.52-10.837333a92.672 92.672 0 0 0 85.76-84.48 2301.44 2301.44 0 0 0 9.386667-203.349333c0-77.994667-4.522667-149.333333-9.386667-203.349334a92.672 92.672 0 0 0-85.76-84.48A3352.746667 3352.746667 0 0 0 512 213.248zM239.36 160.298667A3416.576 3416.576 0 0 1 512 149.248c109.568 0 205.141333 5.546667 272.64 11.050667a156.672 156.672 0 0 1 144.384 142.506666c4.992 55.466667 9.642667 128.725333 9.642667 209.109334s-4.693333 153.642667-9.642667 209.066666a156.672 156.672 0 0 1-144.341333 142.549334 3416.746667 3416.746667 0 0 1-272.682667 11.050666 3417.6 3417.6 0 0 1-272.64-11.050666 156.672 156.672 0 0 1-144.384-142.549334 2365.610667 2365.610667 0 0 1-9.642667-209.066666c0-80.341333 4.693333-153.6 9.642667-209.066667A156.672 156.672 0 0 1 239.36 160.298667z"
                fill="currentColor" p-id="15995"></path>
        </svg>,
    },
    {
        name: "收藏|合集",
        link: "/bilibili/favorite",
        icon: <svg width="18" height="19" viewBox="0 0 20 21" fill="none" xmlns="http://www.w3.org/2000/svg"
                   class="right-entry-icon">
            <path fill-rule="evenodd" clip-rule="evenodd"
                  d="M11.0505 3.16759L12.7915 6.69573C12.954 7.02647 13.2702 7.25612 13.6349 7.30949L17.5294 7.87474C18.448 8.00817 18.8159 9.13785 18.1504 9.78639L15.3331 12.5334C15.0686 12.7905 14.9481 13.1609 15.0104 13.5256L15.6759 17.4031C15.8328 18.3184 14.8721 19.0171 14.0497 18.5845L10.5661 16.7537C10.2402 16.5823 9.85042 16.5823 9.52373 16.7537L6.04087 18.5845C5.21848 19.0171 4.2578 18.3184 4.41468 17.4031L5.07939 13.5256C5.14166 13.1609 5.02198 12.7905 4.75755 12.5334L1.9394 9.78639C1.27469 9.13785 1.64182 8.00817 2.56126 7.87474L6.4549 7.30949C6.82041 7.25612 7.13578 7.02647 7.29832 6.69573L9.04015 3.16759C9.45095 2.33468 10.6389 2.33468 11.0505 3.16759Z"
                  stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"></path>
            <path
                d="M11.603 11.8739C11.413 12.5556 10.7871 13.0554 10.0447 13.0554C9.29592 13.0554 8.66679 12.5467 8.48242 11.8569"
                stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"></path>
        </svg>,
    },
    {
        name: "下载",
        link: "/bilibili/download",
        icon: <svg width="18" height="18" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M12 3V16M12 16L7 11M12 16L17 11" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                  stroke-linejoin="round"/>
            <path d="M3 20H21" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
        </svg>,
    },
    {
        name: "下载历史",
        link: "/bilibili/history",
        icon: <svg width="18" height="19" viewBox="0 0 20 21" fill="none" xmlns="http://www.w3.org/2000/svg"
                   class="right-entry-icon">
            <path fill-rule="evenodd" clip-rule="evenodd"
                  d="M10 1.74286C5.02955 1.74286 1 5.7724 1 10.7429C1 15.7133 5.02955 19.7429 10 19.7429C14.9705 19.7429 19 15.7133 19 10.7429C19 5.7724 14.9705 1.74286 10 1.74286ZM10.0006 3.379C14.0612 3.379 17.3642 6.68282 17.3642 10.7426C17.3642 14.8033 14.0612 18.1063 10.0006 18.1063C5.93996 18.1063 2.63696 14.8033 2.63696 10.7426C2.63696 6.68282 5.93996 3.379 10.0006 3.379Z"
                  fill="currentColor"></path>
            <path d="M9.99985 6.6521V10.743" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"></path>
            <path d="M12.4545 10.7427H10" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"></path>
        </svg>,
    }
]

export default function Menu(): JSXElement {
    const [avatar, setAvatar] = createSignal(bilibiliAvatarFallback);
    const [avatarLoadFailed, setAvatarLoadFailed] = createSignal(false);

    function resetAvatar(): void {
        setAvatarLoadFailed(false);
        setAvatar(bilibiliAvatarFallback);
    }

    const loadAccountAvatar = async () => {
        setAvatarLoadFailed(false);

        try {
            const loggedIn = await getLoggedInDeduped();
            if (!loggedIn) {
                resetAvatar();
                return;
            }

            const profile = await MyInfo();
            setAvatar(profile.face?.trim() ? profile.face : bilibiliAvatarFallback);
        } catch {
            resetAvatar();
        }
    }

    const handleAuthChanged = (event: Event) => {
        const detail = (event as CustomEvent<{ loggedIn?: boolean }>).detail;
        if (detail?.loggedIn === false) {
            resetAvatar();
            return;
        }

        void loadAccountAvatar();
    }

    onMount(() => {
        void loadAccountAvatar();
        window.addEventListener("bilibili-auth-changed", handleAuthChanged);
    })

    onCleanup(() => {
        window.removeEventListener("bilibili-auth-changed", handleAuthChanged);
    })

    return (
        <div
            class="flex h-full w-22 shrink-0 flex-col gap-1.5 overflow-hidden border-r border-base-300/90 bg-linear-to-b from-base-100 via-base-100 to-base-200/50 px-2 py-3">
            <p class="mb-0.5 select-none px-0.5 text-center text-[9px] font-semibold uppercase tracking-widest text-base-content/40">
                B 站
            </p>
            <For each={menuItems}
                 fallback={<span class="px-1 text-center text-[10px] text-base-content/60">暂无</span>}>
                {item => (
                    <Link
                        title={item.name}
                        class="group relative flex flex-col items-center gap-1 rounded-2xl border border-base-300/50 bg-base-100/90 px-1.5 py-2.5 shadow-sm transition-all duration-200 hover:-translate-y-px hover:border-primary/40 hover:bg-base-100 hover:shadow-md"
                        activeProps={{
                            class:
                                "border-primary/50 bg-gradient-to-b from-primary/[0.14] to-primary/[0.06] shadow-md ring-1 ring-primary/25",
                        }}
                        to={item.link}
                    >
                        <span
                            class="flex h-10 w-10 items-center justify-center rounded-2xl bg-linear-to-br from-primary to-secondary text-[0] shadow-inner transition-transform duration-200 group-hover:scale-[1.04] group-active:scale-[0.98] group-data-[status=active]:brightness-110">
                            {item.icon}
                        </span>
                        <span
                            class="max-w-full truncate px-0.5 text-center text-[10px] font-semibold leading-tight tracking-tight text-base-content/90">
                            {item.name}
                        </span>
                    </Link>
                )}
            </For>
            <div class="mt-auto border-t border-base-300/70 pt-3">
                <Link
                    title="我的账号"
                    to="/bilibili/profile"
                    class="group flex flex-col items-center gap-1 rounded-2xl border border-base-300/50 bg-base-100/95 p-2.5 shadow-sm transition-all duration-200 hover:border-primary/40 hover:shadow-md"
                    activeProps={{
                        class: "border-primary/50 bg-gradient-to-b from-primary/[0.12] to-transparent shadow-md ring-1 ring-primary/20",
                    }}
                >
                    <div class="avatar">
                        <div
                            class="h-10 w-10 rounded-full bg-base-200 p-0.5 shadow-inner ring-2 ring-base-100 ring-offset-2 ring-offset-base-100 transition group-hover:ring-primary/30 group-data-[status=active]:ring-primary/50">
                            <img
                                src={avatarLoadFailed() ? bilibiliAvatarFallback : avatar()}
                                alt=""
                                referrerPolicy="no-referrer"
                                class="h-full w-full rounded-full object-cover"
                                onError={() => {
                                    setAvatarLoadFailed(true);
                                    setAvatar(bilibiliAvatarFallback);
                                }}
                            />
                        </div>
                    </div>
                    <span class="text-[10px] font-semibold tracking-tight text-base-content/80">账号</span>
                </Link>
            </div>
        </div>
    )
}
