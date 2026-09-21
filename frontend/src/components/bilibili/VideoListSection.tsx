import {useNavigate} from "@tanstack/solid-router";
import {createEffect, createMemo, createSignal, type JSXElement, Show} from "solid-js";
import {type BilibiliBatchPageLoader, startBilibiliBatch} from "../../lib/bilibili/batchDownload.ts";
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
  /**
   * 一键下载全部的批量会话描述。传入后点击按钮会立即跳转下载页并自动开始：
   * 当前已加载内容作为首批，后续分页由 loader 在会话里“解析地址 → 下载”推进。
   * 不传时不渲染一键下载按钮。
   */
  batchDownload?: {
    title: string;
    totalCount?: number;
    createLoader: () => BilibiliBatchPageLoader;
  };
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
      // 不再预展开分 P 视频（expandBilibiliDownloadItems），直接将原始列表加入下载队列。
      // 后端 BatchResolvePlayUrl 会在解析播放地址时一并获取视频详情，避免重复请求。
      // 后端逐条推送 bilibili-playurl-resolved 事件，前端渐进式更新每张卡片的解析状态。
      addVideos(list);
      await navigate({to: "/bilibili/download"});
    } finally {
      setEnqueueLoading(false);
    }
  }

  // 一键下载全部：立即跳转下载页并自动开始；后续分页由批量会话在下载页继续“解析 → 下载”。
  async function startBatchDownload(): Promise<void> {
    const batch = props.batchDownload;
    if (!batch || enqueueLoading()) return;

    const list = props.medias().filter(m => m.bvid?.trim());
    const started = startBilibiliBatch({
      title: batch.title,
      totalCount: batch.totalCount,
      initialItems: list,
      loader: batch.createLoader(),
    });
    if (!started) {
      props.showToast?.("该内容已在批量下载队列中，或没有可下载的内容", "warning");
      return;
    }
    clearSelection();
    await navigate({to: "/bilibili/download"});
  }

  return (
      <>
        {/*功能区域*/}
        <div class="flex shrink-0 items-center gap-2 border-b border-base-300 px-5 py-3.5">
          <div class="min-w-0 flex-1">
            <h2 class="truncate text-sm font-bold text-base-content">{props.title}</h2>
            <p class="text-xs text-orange-500">
              {props.mediaCount} 个视频
              {props.medias().length < props.mediaCount ? `（已加载 ${props.medias().length}）` : ""}
            </p>
          </div>
          <button class="btn btn-ghost btn-sm" onClick={toggleSelectAllMedia}>
            {allSelected() ? '取消全选' : '全选'}
          </button>
          <Show when={selectedMediaIds().length > 0}>
            <button class="btn btn-ghost btn-sm text-error" onClick={clearSelection}>
              取消已选
            </button>
          </Show>
          <button class="btn btn-outline btn-primary btn-sm"
                  onClick={() => {
                    const picked = props.medias().filter(m => selectedSet().has(m.id));
                    void enqueueAndGoDownload(picked);
                  }}
                  disabled={selectedMediaIds().length === 0 || enqueueLoading()}>
            {enqueueLoading() ? "处理中..." : `下载已选 (${selectedMediaIds().length})`}
          </button>
          <Show when={props.batchDownload}>
            <div class="group relative shrink-0">
              <button class="btn btn-primary btn-sm"
                      type="button"
                      onClick={() => void startBatchDownload()}
                      disabled={props.medias().length === 0 || enqueueLoading()}
                      aria-label="一键下载全部（跳转下载页自动分页下载）">
                {enqueueLoading() ? "处理中..." : "一键下载全部"}
              </button>
              <span
                  role="tooltip"
                  class="pointer-events-none absolute right-0 top-full z-50 mt-1.5 w-max max-w-[min(220px,calc(100vw-24px))] whitespace-normal rounded-md bg-neutral px-2.5 py-1.5 text-center text-[11px] leading-4 text-neutral-content opacity-0 shadow-lg transition-opacity group-hover:opacity-100 group-focus-within:opacity-100"
              >
                跳转下载页，自动分页解析并下载全部
              </span>
            </div>
          </Show>
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
