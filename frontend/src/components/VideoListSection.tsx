import {useNavigate} from "@tanstack/solid-router";
import {createEffect, createMemo, createSignal, type JSXElement, Show} from "solid-js";
import {addVideos} from "../lib/bilibiliStore.ts";
import type {MediaCardItem} from "../lib/model.ts";
import DetailToolbar from "./DetailToolbar";
import VideoCardGrid from "./VideoCardGrid";

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
    const selectedSet = createMemo(() => new Set(selectedMediaIds()));

    const allSelected = () => {
        const cards = props.medias();
        if (cards.length === 0) return false;
        const s = selectedSet();
        return cards.every(c => s.has(c.id));
    };

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
        const list = medias.filter(m => m.bvid?.trim());
        if (list.length === 0) {
            props.showToast?.("没有可下载的视频（缺少 BV 号）", "warning");
            return;
        }
        addVideos(list);
        await navigate({to: "/bilibili/download"});
    }

    return (
        <>
            <DetailToolbar
                title={props.title}
                mediaCount={props.mediaCount}
                selectedCount={selectedMediaIds().length}
                allSelected={allSelected()}
                onToggleSelectAll={toggleSelectAllMedia}
                onClearSelection={clearSelection}
                onDownloadSelected={() => {
                    const picked = props.medias().filter(m => selectedSet().has(m.id));
                    void enqueueAndGoDownload(picked);
                }}
                onDownloadAll={() => void enqueueAndGoDownload(props.medias())}
            />
            <div class="min-h-0 flex-1 overflow-auto p-4">
                <VideoCardGrid
                    medias={props.medias()}
                    selectedSet={selectedSet}
                    onToggleSelect={toggleSelectMedia}
                    onDownloadOne={(m) => void enqueueAndGoDownload([m])}
                />
                <Show when={() => !!props.hasMore?.()}>
                    <div class="mt-4 flex justify-center pb-2">
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
        </>
    );
}
