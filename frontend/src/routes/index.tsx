import {createFileRoute} from '@tanstack/solid-router'
import type {JSXElement} from "solid-js";
import bilibili from "../assets/bilibili_256_256.svg"
import douyin from "../assets/douyin_256_256.svg"

export const Route = createFileRoute('/')({
    component: HomeComponent,
})

function HomeComponent(): JSXElement {
    return (
        <div class={"items-center flex-1 flex"}>
            <div class="w-full h-full flex gap-6 flex-row items-center justify-center">
                <a href="/bilibili" class="flex items-center gap-4 px-8 py-6">
                    <img src={bilibili} alt="Bilibili" class="w-64 h-64"/>
                </a>
                <a href="/douyin" class="flex items-center gap-4 px-8 py-6">
                    <img src={douyin} alt="Douyin" class="w-64 h-64"/>
                </a>
            </div>
        </div>
    )
}
