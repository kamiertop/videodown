import {createFileRoute, redirect} from '@tanstack/solid-router'

export const Route = createFileRoute('/bilibili/')({
    beforeLoad: () => {
        throw redirect({to: '/bilibili/profile'})
    },
})
