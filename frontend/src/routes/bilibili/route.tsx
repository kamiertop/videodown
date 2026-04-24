import {createFileRoute, Outlet} from '@tanstack/solid-router'
import type {JSXElement} from "solid-js";
import Menu from "../../components/bilibili/Menu.tsx";

export const Route = createFileRoute('/bilibili')({
  component: BiliBiliLayout,
})

function BiliBiliLayout(): JSXElement {
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

