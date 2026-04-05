import {createFileRoute} from '@tanstack/solid-router'
import type {JSXElement} from "solid-js";

export const Route = createFileRoute('/bilibili/download')({
    component: DownLoad,
})

function DownLoad(): JSXElement {
    return (
        <>
           
        </>
    )
}
