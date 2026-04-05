import {For, Show, type JSXElement} from "solid-js";

export interface SidebarListItem {
    id: number;
    title: string;
    count: number;
    subtitle?: string;
}

export default function SidebarList<T extends SidebarListItem>(props: {
    list: () => readonly T[];
    selectedId: () => number | null;
    onSelect: (item: T) => void;
    icon: () => JSXElement;
}): JSXElement {
    return (
        <div class="min-h-0 flex-1 overflow-y-auto overscroll-y-contain px-2 py-1">
            <For each={props.list()}>
                {(item) => {
                    const isSelected = () => props.selectedId() === item.id;
                    return (
                        <button
                            class={`group my-0.5 flex h-[42px] w-full items-center gap-2.5 rounded-lg px-2.5 text-left transition-all duration-100 ${
                                isSelected()
                                    ? 'bg-success/15 ring-1 ring-success/30'
                                    : 'hover:bg-accent/10'
                            }`}
                            onClick={() => props.onSelect(item)}
                        >
                            <div
                                class={`flex h-6 w-6 shrink-0 items-center justify-center rounded-md border transition-colors ${
                                    isSelected()
                                        ? 'border-success/40 bg-success/20 text-success'
                                        : 'border-base-300 bg-base-200/50 text-accent group-hover:border-accent/50 group-hover:bg-accent/10'
                                }`}>
                                {props.icon()}
                            </div>
                            <div class="min-w-0 flex-1">
                                <span class={`block truncate text-sm font-semibold ${
                                    isSelected() ? 'text-success' : 'text-base-content'
                                }`}>{item.title}</span>
                                <Show when={item.subtitle}>
                                    <span class="block truncate text-xs text-base-content/50">{item.subtitle}</span>
                                </Show>
                            </div>
                            <span class={`shrink-0 rounded-full px-1.5 py-0.5 text-xs tabular-nums font-bold ${
                                isSelected()
                                    ? 'bg-success/15 text-success'
                                    : 'bg-base-200 text-base-content/70'
                            }`}>{item.count}</span>
                        </button>
                    );
                }}
            </For>
        </div>
    );
}
