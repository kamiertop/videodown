import {createMemo, createSignal, For, Index, type JSXElement, onCleanup, onMount, Show} from "solid-js";
import DouyinVideoCard, {type DouyinVideoCardItem} from "./DouyinVideoCard.tsx";

const GRID_GAP = 8;
const GRID_ROW_GAP = 16;
const GRID_PADDING = 8;

// 与下方虚拟行的响应式列布局保持一致。
function columnCount(width: number): number {
  if (width >= 1536) return 8;
  if (width >= 1280) return 7;
  if (width >= 1024) return 6;
  if (width >= 768) return 5;
  if (width >= 640) return 3;
  return 2;
}

export default function DouyinVideoGrid(props: {
  title: string;
  countText: string;
  items: DouyinVideoCardItem[];
  selectedCount: number;
  allSelected: boolean;
  selectedClass: (id: string) => string;
  onToggleItem: (id: string) => void;
  onToggleAll: () => void;
  onClearSelection: () => void;
  onDownloadSelected: () => void;
  onDownloadAll?: () => void;
  hasMore?: boolean;
  loadingMore?: boolean;
  onLoadMore?: () => void;
}): JSXElement {
  const [scrollElement, setScrollElement] = createSignal<HTMLDivElement>();
  const [viewportWidth, setViewportWidth] = createSignal(0);
  const [viewportHeight, setViewportHeight] = createSignal(720);
  const [scrollTop, setScrollTop] = createSignal(0);
  const columns = createMemo(() => columnCount(viewportWidth()));
  const rowCount = createMemo(() => Math.ceil(props.items.length / columns()));

  // 卡片封面比例固定为 3:5，文字区高度稳定，因此可以按容器宽度估算行高。
  const estimatedRowHeight = createMemo(() => {
    const availableWidth = Math.max(0, viewportWidth() - GRID_PADDING * 2 - GRID_GAP * (columns() - 1));
    const cardWidth = columns() > 0 ? availableWidth / columns() : 160;
    return Math.ceil(cardWidth * 5 / 3 + 72 + GRID_ROW_GAP);
  });
  const virtualRows = createMemo(() => {
    const count = rowCount();
    if (count === 0) return [];

    const rowHeight = estimatedRowHeight();
    const overscan = 4;
    const first = Math.max(0, Math.floor(scrollTop() / rowHeight) - overscan);
    const last = Math.min(
      count - 1,
      Math.ceil((scrollTop() + viewportHeight()) / rowHeight) + overscan,
    );
    const rows: number[] = [];
    for (let index = first; index <= last; index += 1) {
      rows.push(index);
    }
    return rows;
  });

  let resizeObserver: ResizeObserver | undefined;
  onMount(() => {
    const element = scrollElement();
    if (!element) return;

    const updateSize = () => {
      setViewportWidth(element.clientWidth);
      setViewportHeight(element.clientHeight);
    };
    updateSize();
    resizeObserver = new ResizeObserver(updateSize);
    resizeObserver.observe(element);
  });
  onCleanup(() => resizeObserver?.disconnect());

  function rowItems(rowIndex: number): DouyinVideoCardItem[] {
    const start = rowIndex * columns();
    return props.items.slice(start, start + columns());
  }

  return (
    <div class={`flex h-full min-h-0 flex-col ${props.allSelected ? "ring-1 ring-success/40" : ""}`}>
      {/* 固定操作栏：选择与下载按钮不跟随卡片区域滚动。 */}
      <div class={`flex shrink-0 items-center gap-2 border-b px-4 py-3 ${
        props.allSelected ? "border-success/40 bg-success/5" : "border-base-300"
      }`}>
        <div class="min-w-0 flex-1">
          <h3 class="truncate text-sm font-bold text-base-content">{props.title}</h3>
          <p class="text-xs text-base-content/55">{props.countText}</p>
        </div>
        <button class="btn btn-ghost btn-sm" type="button" onClick={props.onToggleAll}>
          {props.allSelected ? "取消全选" : "全选"}
        </button>
        <Show when={props.selectedCount > 0}>
          <button class="btn btn-ghost btn-sm text-error" type="button" onClick={props.onClearSelection}>
            取消已选
          </button>
        </Show>
        <button
          class="btn btn-outline btn-primary btn-sm"
          type="button"
          onClick={props.onDownloadSelected}
          disabled={props.selectedCount === 0}
        >
          去下载 ({props.selectedCount})
        </button>
        <Show when={props.onDownloadAll}>
          {(onDownloadAll) => (
            <button
              class="btn btn-primary btn-sm"
              type="button"
              onClick={onDownloadAll()}
              disabled={props.items.length === 0}
            >
              下载全部
            </button>
          )}
        </Show>
      </div>

      <div
        ref={setScrollElement}
        class="min-h-0 flex-1 overflow-auto"
        onScroll={(event) => setScrollTop(event.currentTarget.scrollTop)}
      >
        {/* 虚拟网格：保留完整滚动高度，但只挂载可视区域附近的行。 */}
        <div
          class="relative"
          style={{height: `${rowCount() * estimatedRowHeight() + GRID_PADDING * 2}px`}}
        >
          <For each={virtualRows()}>
            {(rowIndex) => (
              <div
                data-index={rowIndex}
                class="absolute left-0 top-0 grid w-full gap-x-2 px-2"
                style={{
                  "grid-template-columns": `repeat(${columns()}, minmax(0, 1fr))`,
                  transform: `translateY(${rowIndex * estimatedRowHeight() + GRID_PADDING}px)`,
                }}
              >
                <Index each={rowItems(rowIndex)}>
                  {(item) => (
                    <DouyinVideoCard
                      item={item()}
                      onClick={() => props.onToggleItem(item().id)}
                      selectedClass={props.selectedClass(item().id)}
                    />
                  )}
                </Index>
              </div>
            )}
          </For>
        </div>
        <Show when={props.hasMore}>
          <div class="border-t border-base-200/90 p-2">
            <button
              class="btn btn-ghost btn-sm h-9 w-full"
              type="button"
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
      </div>
    </div>
  );
}
