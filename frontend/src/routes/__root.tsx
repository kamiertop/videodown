import {Outlet, createRootRoute} from '@tanstack/solid-router'
import type {JSXElement} from "solid-js";
import HomeHeader from "../components/Header.tsx";


export const Route = createRootRoute({
    component: RootComponent,
})

function RootComponent(): JSXElement {
    return (
        <div class="min-h-screen bg-base-200 flex flex-col">
            <HomeHeader/>

            <Outlet/>

        </div>
    )
}
