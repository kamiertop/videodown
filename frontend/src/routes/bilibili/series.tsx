import {createFileRoute} from '@tanstack/solid-router'
import type {JSXElement} from "solid-js";

export const Route = createFileRoute('/bilibili/series')({
    component: Series,
})

function Series(): JSXElement {
    return <div>Hello "/bilibili/series"!</div>
}
