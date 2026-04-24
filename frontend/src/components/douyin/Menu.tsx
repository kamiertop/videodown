import {Link} from "@tanstack/solid-router";
import type {JSXElement} from "solid-js";
import {For} from "solid-js";
import douyinAvatarFallback from "../../assets/douyin_256_256.svg";
import {StarIcon} from "../icons/IconStar.tsx";
import IconUsers from "../icons/IconUsers.tsx";

export default function Menu(): JSXElement {
  const menuItems = [
    {
      name: "下载",
      link: "/douyin/download",
      icon: (
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 3V16M12 16L7 11M12 16L17 11" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                stroke-linejoin="round"/>
          <path d="M4 20H20" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
        </svg>
      ),
    },
    {
      name: "关注",
      link: "/douyin/follow",
      icon: <IconUsers class="h-4.5 w-4.5"/>,
    },
    {
      name: "收藏",
      link: "/douyin/favorite",
      icon: <span class="scale-125"><StarIcon/></span>,
    },
  ];

  return (
    <div
      class="flex h-full w-22 shrink-0 flex-col gap-1.5 overflow-hidden border-r border-base-300/90 bg-linear-to-b from-base-100 via-base-100 to-base-200/50 px-2 py-3">
      <p
        class="mb-0.5 select-none px-0.5 text-center text-[9px] font-semibold uppercase tracking-widest text-base-content/40">
        抖音
      </p>

      <For each={menuItems}>
        {(item) => (
          <Link
            to={item.link}
            title={item.name}
            class="group relative flex flex-col items-center gap-1 rounded-2xl border border-base-300/50 bg-base-100/90 px-1.5 py-2.5 shadow-sm transition-all duration-200 hover:-translate-y-px hover:border-primary/40 hover:bg-base-100 hover:shadow-md"
            activeProps={{
              class:
                "border-primary/50 bg-gradient-to-b from-primary/[0.14] to-primary/[0.06] shadow-md ring-1 ring-primary/25",
            }}
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
          to="/douyin/profile"
          title="账号"
          class="group flex flex-col items-center gap-1 rounded-2xl border border-base-300/50 bg-base-100/95 p-2.5 shadow-sm transition-all duration-200 hover:border-primary/40 hover:shadow-md"
          activeProps={{
            class: "border-primary/50 bg-gradient-to-b from-primary/[0.12] to-transparent shadow-md ring-1 ring-primary/20",
          }}
        >
          <div class="avatar">
            <div
              class="h-10 w-10 rounded-full bg-base-200 p-0.5 shadow-inner ring-2 ring-base-100 ring-offset-2 ring-offset-base-100 transition group-hover:ring-primary/30 group-data-[status=active]:ring-primary/50">
              <img
                src={douyinAvatarFallback}
                alt=""
                class="h-full w-full rounded-full object-cover"
              />
            </div>
          </div>
          <span class="text-[10px] font-semibold tracking-tight text-base-content/80">账号</span>
        </Link>
      </div>
    </div>
  )
}
