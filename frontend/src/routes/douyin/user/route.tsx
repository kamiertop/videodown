import {createFileRoute, Outlet} from '@tanstack/solid-router'
import type {JSXElement} from "solid-js";

export const Route = createFileRoute('/douyin/user')({
  component: DouyinUserLayout,
})

function DouyinUserLayout(): JSXElement {
  return <Outlet/>
}
