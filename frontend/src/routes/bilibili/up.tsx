import {createFileRoute} from '@tanstack/solid-router'

export const Route = createFileRoute('/bilibili/up')({
    component: RouteComponent,
})

function RouteComponent() {
    return <div>Hello "/bilibili/up"!</div>
}
