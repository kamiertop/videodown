import {For, type JSXElement, Show} from "solid-js";
import NoCover from "../NoCover.tsx";

export interface CollectionSidebarItem<T> {
  key: string;
  title: string;
  cover: string;
  subtitle: string;
  count: number;
  raw: T;
}

export default function CollectionSidebar<T>(props: {
  items: readonly CollectionSidebarItem<T>[];
  listTitle: string;
  isSelected: (item: T) => boolean;
  onSelect: (item: T) => void;
  hasMore: boolean;
  loadingMore: boolean;
  onLoadMore: () => void;
}): JSXElement {
  return (
    <aside class="flex w-48 shrink-0 flex-col border-r border-base-300 bg-base-100">
      <div class="flex shrink-0 items-center justify-between gap-2 border-b border-base-200 px-3 py-2">
        <div class="min-w-0">
          <h3 class="truncate text-sm font-bold text-base-content">{props.listTitle}</h3>
          <p class="text-xs text-base-content/50">{props.items.length} 个已加载</p>
        </div>
      </div>

      <div class="min-h-0 flex-1 overflow-auto">
        <For each={props.items}>
          {(item) => {
            const selected = () => props.isSelected(item.raw);
            return (
              <button
                type="button"
                class="flex w-full gap-3 border-b border-base-200/80 p-3 text-left transition-colors"
                classList={{"bg-primary/10": selected(), "hover:bg-base-200/35": !selected()}}
                data-key={item.key}
                onClick={() => props.onSelect(item.raw)}
              >
                <div class="h-16 w-16 shrink-0 overflow-hidden rounded-xl border border-base-300 bg-base-200">
                  <Show when={item.cover} fallback={<NoCover/>}>
                    <img
                      src={item.cover}
                      alt=""
                      class="h-full w-full object-cover"
                      loading="lazy"
                      decoding="async"
                      referrerPolicy="no-referrer"
                    />
                  </Show>
                </div>
                <div class="min-w-0 flex-1 self-center">
                  <h4 class={`line-clamp-2 text-sm font-semibold ${selected() ? "text-primary" : "text-base-content"}`}>
                    {item.title}
                  </h4>
                  <p class="mt-1 truncate text-xs text-base-content/50">{item.subtitle}</p>
                  <p class="mt-1 text-xs text-base-content/45">{item.count} 个作品</p>
                </div>
              </button>
            );
          }}
        </For>
      </div>

      <Show when={props.hasMore}>
        <div class="shrink-0 border-t border-base-200 p-2">
          <button
            class="btn btn-ghost btn-sm h-9 w-full"
            onClick={props.onLoadMore}
            disabled={props.loadingMore}
          >
            <Show
              when={!props.loadingMore}
              fallback={<span class="loading loading-spinner loading-sm text-primary"/>}
            >
              加载更多
            </Show>
          </button>
        </div>
      </Show>
    </aside>
  );
}
