import {GetBulkDownloadSleepTime} from "@bindings/github.com/kamiertop/videodown/utils/settings";

/** 一键提取分页前的随机休眠；仅在一键下载流程中调用。 */
export async function waitBulkDownloadPage(onStatus?: (message: string) => void): Promise<void> {
  const configuredSeconds = Number(await GetBulkDownloadSleepTime()) || 0;
  if (configuredSeconds <= 0) return;
  const delayMs = Math.floor(Math.random() * configuredSeconds * 1000);
  if (delayMs <= 0) return;
  onStatus?.(`正在休眠，预计 ${(delayMs / 1000).toFixed(1)} 秒后继续`);
  await new Promise<void>((resolve) => window.setTimeout(resolve, delayMs));
  onStatus?.("正在加载下一页");
}
