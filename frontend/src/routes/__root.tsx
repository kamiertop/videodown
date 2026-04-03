import {Outlet, createRootRoute} from '@tanstack/solid-router'
import type {JSXElement} from "solid-js";
import HomeHeader from "../components/Header.tsx";


export const Route = createRootRoute({
    component: RootComponent,
})

function RootComponent(): JSXElement {
    return (
        <div class="h-screen bg-base-200 flex flex-col overflow-hidden">
            <HomeHeader/>

            <div class="flex-1 min-h-0 overflow-hidden">
                <Outlet/>
            </div>

        </div>
    )
}
