import type {JSXElement} from "solid-js";

export default function DetailLoading(): JSXElement {
    return (
        <div class="flex h-full items-center justify-center">
            <span class="loading loading-spinner loading-md text-primary"></span>
        </div>
    );
}
