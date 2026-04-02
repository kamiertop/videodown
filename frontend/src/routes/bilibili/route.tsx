import {createFileRoute, Outlet} from '@tanstack/solid-router'
import type {JSXElement} from "solid-js";
import BiliBiliMenu from "../../components/BiliBiliMenu.tsx";

export const Route = createFileRoute('/bilibili')({
    component: BiliBiliLayout,
})

function BiliBiliLayout(): JSXElement {
    return (
        <section class={"flex h-full w-full flex-1"}>
            <BiliBiliMenu/>
            <main class={"flex-1 min-w-0 overflow-y-auto"}>
                <Outlet/>
            </main>
        </section>
    )
}

