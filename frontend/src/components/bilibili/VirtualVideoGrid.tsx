import {createMemo, createSignal, For, onCleanup, onMount, Show, type JSXElement} from "solid-js";
import type {MediaCardItem} from "../../lib/model";
import VideoCard from "./VideoCard";

/**
 * 虚拟滚动视频网格。
 * 只渲染可视区域 ± overscan 行内的卡片，其余用 spacer div 撑开高度，
 * 避免大量 DOM 节点导致滚动卡顿——效果对标 B站官网 UP 主视频页。
 *
 * 原理：
 *  - ResizeObserver 监听容器宽高，计算出每行列数、行高
 *  - scroll 事件更新滚动位置，据此计算当前可见行范围
 *  - 外层 div 高度 = 总行高（撑出原生滚动条），内层网格绝对定位偏移到可见位置
 */

const GAP = 12;        // gap-3 = 0.75rem
const PAD = 16;        // p-4 padding
const OVERSCAN = 2;    // 视口外多渲染的行数，避免快速滚动时白屏

/** 根据容器宽度决定列数 */
function columnCount(w: number): number {
  if (w >= 1400) return 6;
  if (w >= 1100) return 5;
  if (w >= 900) return 4;
  if (w >= 600) return 3;
  return 2;
}

export default function VirtualVideoGrid(props: {
  medias: () => MediaCardItem[];
  selectedSet: () => Set<number>;
  onToggleSelect: (id: number) => void;
  onDownloadOne: (media: MediaCardItem) => void;
  hasMore?: () => boolean;
  loadingMore?: () => boolean;
  onLoadMore?: () => void;
}): JSXElement {
  let scrollEl!: HTMLDivElement;
  let ro: ResizeObserver;
  let onScroll: () => void;

  // 容器尺寸 & 滚动位置
  const [vw, setVw] = createSignal(800);
  const [vh, setVh] = createSignal(600);
  const [st, setSt] = createSignal(0);

  // 每行列数 / 卡片宽度 / 行高（封面 16:9 + 文字区 84px + 间距）
  const cols = createMemo(() => columnCount(vw()));
  const cw = createMemo(() => (vw() - PAD * 2 - (cols() - 1) * GAP) / cols());
  const rh = createMemo(() => cw() * 9 / 16 + 84 + GAP);

  // 总行数 & 总高度（至少比视口高 1px，保证可滚动）
  const totalRows = createMemo(() => Math.ceil(props.medias().length / cols()));
  const totalH = createMemo(() => Math.max(totalRows() * rh() + PAD * 2, vh() + 1));

  // 可见行范围（含 overscan）
  const startRow = createMemo(() => Math.max(0, Math.floor(st() / rh()) - OVERSCAN));
  const endRow = createMemo(() => Math.min(totalRows(), Math.ceil((st() + vh()) / rh()) + OVERSCAN));

  // 当前可见的卡片切片
  const visible = createMemo(() => {
    const s = startRow() * cols();
    const e = Math.min(endRow() * cols(), props.medias().length);
    return props.medias().slice(s, e);
  });

  // 可见网格的垂直偏移
  const offsetY = createMemo(() => startRow() * rh() + PAD);

  const showLoadMore = () => !!props.hasMore?.();

  // 挂载 ResizeObserver + scroll 监听
  onMount(() => {
    ro = new ResizeObserver(([e]) => {
      setVw(e.contentRect.width);
      setVh(e.contentRect.height);
    });
    ro.observe(scrollEl);

    onScroll = () => setSt(scrollEl.scrollTop);
    scrollEl.addEventListener('scroll', onScroll, {passive: true});

    // 初始尺寸
    setVw(scrollEl.clientWidth);
    setVh(scrollEl.clientHeight);
  });

  onCleanup(() => {
    ro?.disconnect();
    if (scrollEl && onScroll) scrollEl.removeEventListener('scroll', onScroll);
  });

  return (
    <div
      ref={scrollEl}
      class="min-h-0 flex-1 overflow-auto overscroll-contain"
    >
      {/* 撑高容器，让原生滚动条表现正确 */}
      <div style={{height: `${totalH()}px`, position: 'relative'}}>
        {/* 可见卡片网格，绝对定位到正确偏移 */}
        <div
          style={{
            position: 'absolute',
            top: `${offsetY()}px`,
            left: `${PAD}px`,
            right: `${PAD}px`,
            display: 'grid',
            'grid-template-columns': `repeat(${cols()}, 1fr)`,
            gap: `${GAP}px`,
          }}
        >
          <For each={visible()}>
            {(media) => (
              <VideoCard
                media={media}
                isSelected={props.selectedSet().has(media.id)}
                onToggleSelect={props.onToggleSelect}
                onDownloadOne={props.onDownloadOne}
              />
            )}
          </For>
        </div>
      </div>
      <Show when={showLoadMore()}>
        <div class="flex justify-center pb-3">
          <button
            type="button"
            class="btn btn-outline btn-sm"
            onClick={props.onLoadMore}
            disabled={props.loadingMore?.()}
          >
            {props.loadingMore?.() ? "加载中..." : "加载更多"}
          </button>
        </div>
      </Show>
    </div>
  );
}
