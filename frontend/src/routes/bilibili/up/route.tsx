import {createFileRoute, Outlet} from '@tanstack/solid-router'
import type {JSXElement} from "solid-js";

export const Route = createFileRoute('/bilibili/up')({
    component: UpLayout,
})

function UpLayout(): JSXElement {
    return <Outlet/>
}
