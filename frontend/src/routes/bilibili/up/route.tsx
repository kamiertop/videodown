import {createFileRoute, Outlet} from '@tanstack/solid-router'
import type {JSXElement} from "solid-js";

export const Route = createFileRoute('/bilibili/up')({
    component: UpLayout,
})

// UP主的2个页面往这里填充
function UpLayout(): JSXElement {
    return <Outlet/>
}
