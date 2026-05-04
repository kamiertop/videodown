import {createMemo, createSignal, For, Index, type JSXElement, onCleanup, onMount, Show} from "solid-js";
import type {DouyinDownloadItem} from "../../lib/douyinStore.ts";
import IconLivePhoto from "../icons/IconLivePhoto.tsx";
import NoCover from "../NoCover.tsx";

const GRID_GAP = 8;
const GRID_ROW_GAP = 24;
const GRID_PADDING = 8;

export interface DouyinVideoCardItem {
  // VideoContentPanel 已经把 AwemeItem 压平，这里不再依赖后端模型字段。
  id: string;
  cover: string;
  title: string;
  author: string;
  publishText: string;
  durationText: string;
  downloadItem: DouyinDownloadItem;
  mediaBadge?: "image" | "live-photo";
}

// 与下方虚拟行的响应式列布局保持一致。
function columnCount(width: number): number {
  if (width >= 1500) return 8;
  if (width >= 1240) return 7;
  if (width >= 1040) return 6;
  if (width >= 840) return 5;
  if (width >= 620) return 4;
  if (width >= 420) return 3;
  return 2;
}

function VideoCard(props: {
  item: DouyinVideoCardItem;
  onClick: () => void;
  selectedClass: string;
}): JSXElement {
  return (
    <article
      class={`flex cursor-pointer flex-col overflow-hidden rounded-lg transition-colors ${props.selectedClass}`}
      // 虚拟滚动已经减少 DOM 数量；content-visibility 再降低离屏卡片绘制成本。
      style={{"content-visibility": "auto", "contain-intrinsic-size": "220px"}}
      onClick={props.onClick}
    >
      <div class="relative aspect-3/5 w-full overflow-hidden bg-base-200">
        <Show when={props.item.cover} fallback={<NoCover/>}>
          <img
            src={props.item.cover}
            class="h-full w-full object-cover"
            alt={props.item.title}
            loading="lazy"
            decoding="async"
            referrerPolicy="no-referrer"
          />
        </Show>
        <Show when={props.item.mediaBadge === "image"}>
          <svg
            width="12"
            height="12"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
            class="absolute right-1.5 top-1.5 z-10 h-3.5 w-3.5 drop-shadow-[0_1px_2px_rgba(0,0,0,0.65)]"
            viewBox="0 0 12 12"
            aria-label="图文"
          >
            <path
              d="M1.455 0C.65 0 0 .651 0 1.455V8c0 .803.651 1.455 1.455 1.455H8c.803 0 1.455-.652 1.455-1.455V1.455C9.455.65 8.803 0 8 0H1.455z"
              fill="#fff"></path>
            <path
              d="M4 12a1.455 1.455 0 0 1-1.455-1.454h5.819a2.182 2.182 0 0 0 2.181-2.182V2.545C11.35 2.545 12 3.197 12 4v5.09A2.909 2.909 0 0 1 9.09 12H4z"
              fill="#fff"></path>
          </svg>
        </Show>
        <Show when={props.item.mediaBadge === "live-photo"}>
          <IconLivePhoto/>
        </Show>
        <Show when={!props.item.mediaBadge}>
          <span
            class="absolute bottom-1 right-1 rounded-md bg-black/65 px-1.5 py-1 text-[0.65rem] font-medium leading-none text-white">
            {props.item.durationText}
          </span>
        </Show>
      </div>

      <div class="flex flex-col p-2">
        <h3 class="line-clamp-2 text-[12px] font-semibold leading-4.5 text-base-content">
          {props.item.title}
        </h3>
        <p class="mt-1 line-clamp-1 text-[10px] text-base-content/50">
          @{props.item.author}
        </p>
        <p class="mt-1.5 text-[10px] text-base-content/45">
          发布 {props.item.publishText}
        </p>
      </div>
    </article>
  );
}

// 视频卡片网格：包含固定操作栏和虚拟滚动区域。操作栏提供全选、下载等功能，滚动区域仅渲染可视范围内的行以提升性能。
export default function VideoGrid(props: {
  title: string;
  items: DouyinVideoCardItem[];
  selectedCount: number;
  allSelected: boolean;
  selectedClass: (id: string) => string;
  onToggleItem: (id: string) => void;
  onToggleAll: () => void;
  onClearSelection: () => void;
  onDownloadSelected: () => void;
  onDownloadAll?: () => void;
  refreshing?: boolean;
  onRefresh?: () => void;
  hasMore?: boolean;
  loadingMore?: boolean;
  onLoadMore?: () => void;
}): JSXElement {
  // scrollElement 是网格滚动容器，ResizeObserver 和 onScroll 都围绕它工作。
  const [scrollElement, setScrollElement] = createSignal<HTMLDivElement>();
  // viewportWidth 决定列数；viewportHeight 决定需要渲染多少虚拟行。
  const [viewportWidth, setViewportWidth] = createSignal(0);
  const [viewportHeight, setViewportHeight] = createSignal(720);
  // scrollTop 是虚拟列表计算第一行/最后一行的核心输入。
  const [scrollTop, setScrollTop] = createSignal(0);
  // columns 跟容器宽度走，不直接跟窗口宽度走，避免侧栏宽度变化后错列。
  const columns = createMemo(() => columnCount(viewportWidth()));
  // rowCount 是完整数据应有的总行数，虚拟滚动用它撑出真实滚动高度。
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
    // 多渲染几行，滚动时不容易露白；行高是估算值，所以 overscan 不能太小。
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

    // ResizeObserver 保证窗口变化、侧栏宽度变化后列数和行高能重新计算。
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
    // 每个虚拟行只切出当前行需要展示的卡片。
    const start = rowIndex * columns();
    return props.items.slice(start, start + columns());
  }

  return (
    <div class={`flex h-full min-h-0 w-full min-w-0 flex-col ${props.allSelected ? "ring-1 ring-success/40" : ""}`}>
      {/* 固定操作栏：选择与下载按钮不跟随卡片区域滚动。 */}
      <div class="flex shrink-0 items-center gap-2 border-b px-4 py-3"
           classList={{"border-success/40 bg-success/5": props.allSelected, "border-base-300": !props.allSelected}}>
        <div class="min-w-0 flex-1">
          <h3 class="truncate text-sm font-bold text-base-content">{props.title}</h3>
          <p class="text-xs text-base-content/55">{props.items.length} 个已加载</p>
        </div>
        <button class="btn btn-ghost btn-sm" type="button" onClick={props.onToggleAll}>
          {props.allSelected ? "取消全选" : "全选"}
        </button>
        <Show when={props.onRefresh}>
          {(onRefresh) => (
            <button class="btn btn-ghost btn-sm" type="button" onClick={onRefresh()} title="刷新">
              <Show when={!props.refreshing} fallback={<span class="loading loading-spinner loading-xs"/>}>
                刷新
              </Show>
            </button>
          )}
        </Show>
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
              // 每一行用 absolute + translateY 定位，避免为不可见行创建 DOM。
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
                    <VideoCard
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
