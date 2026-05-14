import {useNavigate} from "@tanstack/solid-router";
import {createEffect, createMemo, createSignal, type JSXElement, Show} from "solid-js";
import {expandBilibiliDownloadItems} from "../../lib/bilibili/downloadExpand.ts";
import {addVideos} from "../../lib/bilibili/store.ts";
import type {MediaCardItem} from "../../lib/model.ts";
import VirtualVideoGrid from "./VirtualVideoGrid";

/** 工具栏 + 卡片网格 + 多选；入队后跳转下载页。`selectionResetKey` 变则清空勾选。 */
export default function VideoListSection(props: {
  title: string;
  mediaCount: number;
  medias: () => MediaCardItem[];
  selectionResetKey: () => string | number;
  /** 无 BV 时提示（有 Toast 的页面可传，如 UP 详情） */
  showToast?: (message: string, type?: "warning") => void;
  hasMore?: () => boolean;
  loadingMore?: () => boolean;
  onLoadMore?: () => void;
}): JSXElement {
  const navigate = useNavigate();
  const [selectedMediaIds, setSelectedMediaIds] = createSignal<number[]>([]);
  const [enqueueLoading, setEnqueueLoading] = createSignal(false);
  const selectedSet = createMemo(() => new Set(selectedMediaIds()));

  function allSelected(): boolean {
    const cards = props.medias();
    if (cards.length === 0) return false;
    const s = selectedSet();
    return cards.every(c => s.has(c.id));
  }

  createEffect(() => {
    props.selectionResetKey();
    setSelectedMediaIds([]);
  });

  const toggleSelectMedia = (id: number) => {
    setSelectedMediaIds(prev => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return Array.from(next);
    });
  };

  const toggleSelectAllMedia = () => {
    const cards = props.medias();
    setSelectedMediaIds(allSelected() ? [] : cards.map(c => c.id));
  };

  // 清空已选择的
  function clearSelection(): void {
    setSelectedMediaIds([]);
  }

  async function enqueueAndGoDownload(medias: MediaCardItem[]) {
    if (enqueueLoading()) return;
    const list = medias.filter(m => m.bvid?.trim());
    if (list.length === 0) {
      props.showToast?.("没有可下载的视频（缺少 BV 号）", "warning");
      return;
    }
    setEnqueueLoading(true);
    try {
      // 入队前才查详情展开分 P；列表渲染阶段不做额外请求。
      const expanded = await expandBilibiliDownloadItems(list);
      addVideos(expanded);
      await navigate({to: "/bilibili/download"});
    } finally {
      setEnqueueLoading(false);
    }
  }

  return (
      <>
        {/*功能区域*/}
        <div class="flex shrink-0 items-center gap-2 border-b border-base-300 px-5 py-3.5">
          <div class="min-w-0 flex-1">
            <h2 class="truncate text-sm font-bold text-base-content">{props.title}</h2>
            <p class="text-xs text-orange-500">{props.mediaCount} 个视频</p>
          </div>
          <button class="btn btn-ghost btn-sm" onClick={toggleSelectAllMedia}>
            {allSelected() ? '取消全选' : '全选'}
          </button>
          <Show when={selectedMediaIds().length > 0}>
            <button class="btn btn-ghost btn-sm text-error" onClick={clearSelection}>
              取消已选
            </button>
          </Show>
          <button class="btn btn-outline btn-primary btn-sm" onClick={() => {
            const picked = props.medias().filter(m => selectedSet().has(m.id));
            void enqueueAndGoDownload(picked);
          }}
                  disabled={selectedMediaIds().length === 0 || enqueueLoading()}>
            {enqueueLoading() ? "处理中..." : `下载已选 (${selectedMediaIds().length})`}
          </button>
          <button class="btn btn-primary btn-sm"
                  disabled={enqueueLoading()}
                  onClick={() => void enqueueAndGoDownload(props.medias())}>
            {enqueueLoading() ? "处理中..." : "下载全部"}
          </button>
        </div>
        {/* 视频卡片网格 — 虚拟滚动，只渲染可见行 */}
        <VirtualVideoGrid
            medias={props.medias}
            selectedSet={selectedSet}
            onToggleSelect={toggleSelectMedia}
            onDownloadOne={(m) => void enqueueAndGoDownload([m])}
            hasMore={props.hasMore}
            loadingMore={props.loadingMore}
            onLoadMore={props.onLoadMore}
        />
      </>
  );
}
