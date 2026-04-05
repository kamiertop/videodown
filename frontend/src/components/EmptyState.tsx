import {Show, type JSXElement} from "solid-js";

export default function EmptyState(props: { title: string; description?: string; compact?: boolean }): JSXElement {
    return (
        <div class={`flex h-full items-center justify-center text-base-content/40 ${props.compact ? 'py-10' : ''}`}>
            <div class="text-center">
                <svg xmlns="http://www.w3.org/2000/svg" class="mx-auto h-14 w-14" fill="none" viewBox="0 0 24 24"
                     stroke="currentColor" stroke-width="1">
                    <path stroke-linecap="round" stroke-linejoin="round"
                          d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
                </svg>
                <p class="mt-3 text-sm font-semibold text-base-content/60">{props.title}</p>
                <Show when={props.description}>
                    <p class="mt-1 text-xs text-base-content/50">{props.description}</p>
                </Show>
            </div>
        </div>
    );
}
