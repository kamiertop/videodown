/* @refresh reload */
import {render} from 'solid-js/web'
import {RouterProvider, createRouter} from '@tanstack/solid-router'
import {routeTree} from './routeTree.gen'
import './index.css'

// 创建路由器实例
const router = createRouter({
    routeTree,
    defaultPreload: 'intent',
})

const root = document.getElementById('root')

render(() => <RouterProvider router={router}/>, root!)
