import {createFileRoute} from '@tanstack/solid-router'
import type {JSXElement} from "solid-js";

export const Route = createFileRoute('/douyin')({
  component: DouyinComponent,
})

function DouyinComponent(): JSXElement {
  return (
    <div class="flex flex-col gap-6 p-8">
      <h1 class="text-2xl font-bold">正在赶来~</h1>
    </div>
  )
}
