import {createFileRoute, Outlet} from '@tanstack/solid-router'
import type {JSXElement} from "solid-js";
import {onMount} from "solid-js";
import Menu from "../../components/bilibili/Menu.tsx";
import {refreshBilibiliCookieDeduped} from "../../lib/bilibili/auth.ts";

export const Route = createFileRoute('/bilibili')({
  component: BiliBiliLayout,
})

function BiliBiliLayout(): JSXElement {
  onMount(() => {
    void refreshBilibiliCookieDeduped().catch((error: unknown) => {
      console.warn("refresh bilibili cookie failed", error);
    });
  });

  return (
      <section class={"flex h-full min-h-0 w-full flex-1 overflow-hidden"}>
        <aside class={"h-full shrink-0 contain-[layout_paint_style]"}>
          <Menu/>
        </aside>
        <main class="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden contain-[layout_paint]">
          <Outlet/>
        </main>
      </section>
  )
}
