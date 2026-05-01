import {createRootRoute, Outlet} from '@tanstack/solid-router'
import {type JSXElement, onMount} from "solid-js";
import {GetTheme} from "../../wailsjs/go/utils/Settings";
import HomeHeader from "../components/Header.tsx";


export const Route = createRootRoute({
  component: RootComponent,
})

function RootComponent(): JSXElement {
  onMount(async () => {
    const savedTheme: string = await GetTheme();
    document.documentElement.setAttribute('data-theme', savedTheme);
  })
  return (
    <div class="h-screen bg-base-200 flex flex-col">
      <HomeHeader/>
      <div class="flex-1 min-h-0 ">
        <Outlet/>
      </div>
    </div>
  )
}
