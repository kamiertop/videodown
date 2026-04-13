import {createFileRoute} from '@tanstack/solid-router'
import type {JSXElement} from "solid-js";

export const Route = createFileRoute('/bilibili/history')({
    component: History,
})

function History(): JSXElement {
    return <div></div>
}
