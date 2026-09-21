import {type JSXElement} from "solid-js";
import {
  bilibiliBatchSessions,
  dismissBilibiliBatch,
  resumeBilibiliBatch,
  stopBilibiliBatch,
} from "../../lib/bilibili/batchDownload.ts";
import {BatchSessionCards, type BatchStateView, type BatchStatusBadge} from "../BatchSessionCards.tsx";

function statusBadge(state: BatchStateView): BatchStatusBadge {
  // 停止是协作式的：置位后要等当前批（含任务后休眠）跑完，期间显示“停止中”。
  if (state.stopping) return {text: "停止中", badgeClass: "badge-warning", spinning: true};
  switch (state.status) {
    case "queued":
      return {text: "排队中", badgeClass: "badge-info", spinning: false};
    case "resolving":
      return {text: "解析地址", badgeClass: "badge-info", spinning: true};
    case "downloading":
      return {text: "下载中", badgeClass: "badge-primary", spinning: true};
    case "loadingPage":
      return {text: "加载下一页", badgeClass: "badge-info", spinning: true};
    case "retrying":
      return {text: "重试失败项", badgeClass: "badge-warning", spinning: true};
    case "done":
      return state.failedCount > 0
          ? {text: "部分失败", badgeClass: "badge-warning", spinning: false}
          : {text: "已完成", badgeClass: "badge-success", spinning: false};
    case "stopped":
      return {text: "已停止", badgeClass: "badge-neutral", spinning: false};
    default:
      return {text: "未知", badgeClass: "badge-ghost", spinning: false};
  }
}

/** 下载页顶部的批量任务卡片：布局与交互见共用的 BatchSessionCards，这里只提供 B 站的状态文案（多一个解析阶段）。 */
export default function BilibiliBatchStatusCard(): JSXElement {
  return (
      <BatchSessionCards
          sessions={bilibiliBatchSessions}
          statusBadge={statusBadge}
          onDismiss={dismissBilibiliBatch}
          onResume={resumeBilibiliBatch}
          onStop={stopBilibiliBatch}
      />
  );
}
