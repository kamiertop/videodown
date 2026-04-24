import {createFileRoute, redirect} from '@tanstack/solid-router'

export const Route = createFileRoute('/douyin/')({
  beforeLoad: () => {
    throw redirect({to: '/douyin/download'})
  },
})
