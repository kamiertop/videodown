import type {JSXElement} from "solid-js";
import {createSignal, For, onCleanup, onMount} from "solid-js";
import {Link} from "@tanstack/solid-router";
import {IsLoggedIn, MyInfo} from "../../wailsjs/go/api/BiliBili";
import bilibiliAvatarFallback from "../assets/bilibili_256_256.svg";

const menuItems = [
    {
        name: "单视频",
        link: "/bilibili/single",
        icon: <svg x="1775035399268" class="icon" viewBox="0 0 1024 1024"
                   xmlns="http://www.w3.org/2000/svg" p-id="17180" width="18" height="18">
            <path
                d="M347.2384 196.5056l110.592 110.6944h108.2368l110.6944-110.6944c13.312-13.312 34.9184-13.312 48.3328 0 13.312 13.312 13.312 34.9184 0 48.3328l-62.464 62.3616h71.168c65.9456 0 119.5008 53.4528 119.5008 119.5008v273.1008c0 65.9456-53.4528 119.5008-119.5008 119.5008H290.0992c-65.9456 0-119.5008-53.4528-119.5008-119.5008V426.7008c0-65.9456 53.4528-119.5008 119.5008-119.5008h71.168l-62.3616-62.464c-13.312-13.312-13.312-34.9184 0-48.3328 13.312-13.312 34.9184-13.312 48.3328 0.1024z m386.6624 178.9952H290.0992c-26.4192 0-48.4352 20.0704-50.9952 46.2848l-0.2048 4.9152v273.1008c0 26.4192 20.0704 48.4352 46.2848 50.9952l4.9152 0.2048h443.6992c26.4192 0 48.4352-20.0704 50.9952-46.2848l0.2048-4.9152V426.7008c0.1024-28.2624-22.8352-51.2-51.0976-51.2z m-358.4 102.4c18.8416 0 34.0992 15.2576 34.0992 34.0992v68.3008c0 18.8416-15.2576 34.0992-34.0992 34.0992s-34.0992-15.2576-34.0992-34.0992V512c-0.1024-18.8416 15.2576-34.0992 34.0992-34.0992z m272.9984 0c18.8416 0 34.0992 15.2576 34.0992 34.0992v68.3008c0 18.8416-15.2576 34.0992-34.0992 34.0992S614.4 599.1424 614.4 580.3008V512c0-18.8416 15.2576-34.0992 34.0992-34.0992z"
                fill="#040000" p-id="17181"></path>
        </svg>,
    },
    {
        name: "收藏",
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
        name: "UP主",
        link: "/bilibili/up",
        icon: <svg x="1775035354017" class="icon" viewBox="0 0 1024 1024" version="1.1"
                   xmlns="http://www.w3.org/2000/svg" p-id="15993" width="18" height="18">
            <path
                d="M262.4 351.914667a32 32 0 0 1 32 32v161.109333a62.890667 62.890667 0 0 0 125.781333 0V383.914667a32 32 0 0 1 64 0v161.109333a126.890667 126.890667 0 1 1-253.781333 0V383.914667a32 32 0 0 1 32-32zM539.818667 383.914667a32 32 0 0 1 32-32h99.114666a122.666667 122.666667 0 1 1 0 245.333333h-67.114666v42.666667a32 32 0 0 1-64 0v-256z m64 149.333333h67.114666a58.666667 58.666667 0 1 0 0-117.333333h-67.114666v117.333333z"
                fill="#2c2c2c" p-id="15994"></path>
            <path
                d="M512 213.248c-107.434667 0-201.258667 5.461333-267.52 10.837333a92.672 92.672 0 0 0-85.76 84.48 2301.738667 2301.738667 0 0 0-9.386667 203.349334c0 77.952 4.522667 149.248 9.386667 203.349333a92.672 92.672 0 0 0 85.76 84.48 3354.026667 3354.026667 0 0 0 267.52 10.837333c107.434667 0 201.258667-5.461333 267.52-10.837333a92.672 92.672 0 0 0 85.76-84.48 2301.44 2301.44 0 0 0 9.386667-203.349333c0-77.994667-4.522667-149.333333-9.386667-203.349334a92.672 92.672 0 0 0-85.76-84.48A3352.746667 3352.746667 0 0 0 512 213.248zM239.36 160.298667A3416.576 3416.576 0 0 1 512 149.248c109.568 0 205.141333 5.546667 272.64 11.050667a156.672 156.672 0 0 1 144.384 142.506666c4.992 55.466667 9.642667 128.725333 9.642667 209.109334s-4.693333 153.642667-9.642667 209.066666a156.672 156.672 0 0 1-144.341333 142.549334 3416.746667 3416.746667 0 0 1-272.682667 11.050666 3417.6 3417.6 0 0 1-272.64-11.050666 156.672 156.672 0 0 1-144.384-142.549334 2365.610667 2365.610667 0 0 1-9.642667-209.066666c0-80.341333 4.693333-153.6 9.642667-209.066667A156.672 156.672 0 0 1 239.36 160.298667z"
                fill="#2c2c2c" p-id="15995"></path>
        </svg>,
    },
    {
        name: "合集|系列",
        link: "/bilibili/series",
        icon: <svg class="vui_icon fav-sidebar-item__icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"
                   width="18" height="18" xmlns:xlink="http://www.w3.org/1999/xlink">
            <path
                d="M8.298191666666666 2.9306916666666667C9.374833333333335 2.4180166666666665 10.625208333333333 2.4180166666666665 11.701833333333333 2.9306916666666667L17.280708333333333 5.5873083333333335C18.4525 6.145308333333333 18.4525 7.813125 17.280708333333333 8.371108333333334L11.701833333333333 11.027750000000001C10.625208333333333 11.540416666666667 9.374833333333335 11.540416666666667 8.298191666666666 11.027750000000001L2.719316666666667 8.371108333333334C1.5475166666666667 7.813125 1.5475166666666667 6.145308333333333 2.719316666666667 5.5873083333333335L8.298191666666666 2.9306916666666667zM11.164416666666666 4.059275C10.427791666666668 3.7084916666666667 9.57225 3.7084916666666667 8.835625 4.059275L3.2567333333333335 6.715883333333334C3.0349916666666665 6.821475 3.0349916666666665 7.136958333333333 3.2567333333333335 7.242541666666668L8.835625 9.899166666666666C9.57225 10.249916666666666 10.427791666666668 10.249916666666666 11.164416666666666 9.899166666666666L16.743291666666668 7.242541666666668C16.963958333333334 7.137466666666667 16.965791666666668 6.821833333333333 16.743291666666668 6.715883333333334L11.164416666666666 4.059275z"
                fill="currentColor"></path>
            <path
                d="M1.7903916666666668 10.256208333333333C1.9388000000000003 9.944583333333334 2.311741666666667 9.81225 2.623391666666667 9.960666666666667L8.835583333333334 12.918833333333334C9.57225 13.269625000000001 10.42775 13.269625000000001 11.164416666666666 12.918833333333334L17.376583333333336 9.960666666666667C17.68825 9.81225 18.061208333333333 9.944583333333334 18.209583333333335 10.256208333333333C18.358 10.567916666666667 18.225666666666665 10.940833333333334 17.914 11.089208333333334L11.701833333333333 14.047458333333335C10.625166666666667 14.560125 9.374791666666667 14.560125 8.298175 14.047458333333335L2.085975 11.089208333333334C1.774325 10.940833333333334 1.6419833333333334 10.567916666666667 1.7903916666666668 10.256208333333333z"
                fill="currentColor"></path>
            <path
                d="M1.7903916666666668 13.173000000000002C1.9388000000000003 12.861333333333334 2.311741666666667 12.729000000000001 2.623391666666667 12.877458333333333L8.835583333333334 15.835625C9.57225 16.186374999999998 10.42775 16.186374999999998 11.164416666666666 15.835625L17.376583333333336 12.877458333333333C17.68825 12.729000000000001 18.061208333333333 12.861333333333334 18.209583333333335 13.173000000000002C18.358 13.484666666666667 18.225666666666665 13.857583333333334 17.914 14.006000000000002L11.701833333333333 16.964208333333335C10.625166666666667 17.476875 9.374791666666667 17.476875 8.298175 16.964208333333335L2.085975 14.006000000000002C1.774325 13.857583333333334 1.6419833333333334 13.484666666666667 1.7903916666666668 13.173000000000002z"
                fill="currentColor"></path>
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

export default function BiliBiliMenu(): JSXElement {
    const [avatar, setAvatar] = createSignal(bilibiliAvatarFallback);
    const [avatarLoadFailed, setAvatarLoadFailed] = createSignal(false);

    const resetAvatar = () => {
        setAvatarLoadFailed(false);
        setAvatar(bilibiliAvatarFallback);
    }

    const loadAccountAvatar = async () => {
        setAvatarLoadFailed(false);

        try {
            const loggedIn = await IsLoggedIn();
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
            class="flex h-full w-[5.5rem] shrink-0 flex-col gap-1.5 overflow-hidden border-r border-base-300/90 bg-gradient-to-b from-base-100 via-base-100 to-base-200/50 px-2 py-3">
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
                            class="flex h-10 w-10 items-center justify-center rounded-2xl bg-gradient-to-br from-primary to-secondary text-[0] text-white shadow-inner transition-transform duration-200 group-hover:scale-[1.04] group-active:scale-[0.98] group-data-[status=active]:brightness-110">
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
