import {createFileRoute} from '@tanstack/solid-router'
import type {JSXElement} from "solid-js";

export const Route = createFileRoute('/bilibili/favorite')({
    component: Favorite,
})

function Favorite(): JSXElement {
    return <div>Hello "/bilibili/favorite"!</div>
}
