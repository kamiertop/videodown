import {createFileRoute, Link} from '@tanstack/solid-router'
import type {JSXElement} from "solid-js";
import bilibili from "../assets/bilibili_256_256.svg"
import douyin from "../assets/douyin_256_256.svg"

export const Route = createFileRoute('/')({
    component: HomeComponent,
})

function HomeComponent(): JSXElement {
    return (
        <div class="flex h-full w-full flex-1 items-center justify-center">
            <div class="flex w-full flex-row items-center justify-center gap-6">
                <Link to="/bilibili/profile" preload="intent" class="flex items-center gap-4 px-8 py-6">
                    <img src={bilibili} alt="Bilibili" class="w-64 h-64"/>
                </Link>
                <Link to="/douyin" preload="intent" class="flex items-center gap-4 px-8 py-6">
                    <img src={douyin} alt="Douyin" class="w-64 h-64"/>
                </Link>
            </div>
        </div>
    )
}
