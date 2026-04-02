import {createFileRoute} from '@tanstack/solid-router'

export const Route = createFileRoute('/bilibili/single')({
    component: RouteComponent,
})

function RouteComponent() {
    return <div>Hello "/bilibili/single"!</div>
}
