import {Show, type JSXElement} from "solid-js";

/**
 * 收藏夹|合集视频卡片展示页面上面的小组件，显示选中数量和下载按钮等
 */

export default function DetailToolbar(props: {
    title: string;  // 收藏夹|合集名称
    mediaCount: number; // 媒体数量
    selectedCount: number;  // 已选数量
    allSelected: boolean;   // 是否全选
    onToggleSelectAll: () => void;  // 全选
    onClearSelection: () => void;   // 取消已选
    onDownloadSelected: () => void; // 下载已选
    onDownloadAll: () => void;      // 下载全部
}): JSXElement {
    return (
        <div class="flex shrink-0 items-center gap-2 border-b border-base-300 px-5 py-3.5">
            <div class="min-w-0 flex-1">
                <h2 class="truncate text-sm font-bold text-base-content">{props.title}</h2>
                <p class="text-xs text-orange-500">{props.mediaCount} 个视频</p>
            </div>
            <button class="btn btn-ghost btn-sm" onClick={props.onToggleSelectAll}>
                {props.allSelected ? '取消全选' : '全选'}
            </button>
            <Show when={props.selectedCount > 0}>
                <button class="btn btn-ghost btn-sm text-error" onClick={props.onClearSelection}>
                    取消已选
                </button>
            </Show>
            <button class="btn btn-outline btn-primary btn-sm" onClick={props.onDownloadSelected}
                    disabled={props.selectedCount === 0}>
                下载已选 ({props.selectedCount})
            </button>
            <button class="btn btn-primary btn-sm" onClick={props.onDownloadAll}>下载全部</button>
        </div>
    );
}
