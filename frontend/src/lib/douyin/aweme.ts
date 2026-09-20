import * as model from "@bindings/github.com/kamiertop/videodown/douyin/model/models";
import {
  defaultDouyinVideoOption,
  douyinCoverCandidates,
  douyinDownloadAssets,
  douyinImageURLs,
  douyinMediaBadge,
  douyinMusicURL,
  douyinVideoOptions,
} from "./media.ts";
import type {DouyinDownloadItem} from "./store.ts";

function normalizeDouyinDuration(value?: number): number {
  // 抖音部分接口返回毫秒，部分字段可能已经是秒；展示层统一转成秒。
  if (!value || value <= 0) return 0;
  return value >= 1000 ? Math.floor(value / 1000) : value;
}

export function awemeTitle(item: model.AwemeItem): string {
  return item.item_title || item.desc || item.caption || `作品 ${item.aweme_id || ""}`.trim();
}

/** 展示用封面候选（字符串 URL）；图文作品用第一张图兜底。 */
export function awemeCoverURLCandidates(item: model.AwemeItem): string[] {
  const candidates = [
    ...(item.video?.raw_cover?.url_list ?? []),
    ...(item.video?.cover?.url_list ?? []),
    ...(item.video?.origin_cover?.url_list ?? []),
    ...(item.images?.[0]?.url_list ?? []),
  ];
  return [...new Set(candidates.filter(Boolean))];
}

/** 作品下载 ID：aweme_id 最稳定，其余只用于接口缺字段时兜底。 */
export function awemeDownloadID(item: model.AwemeItem, index?: number): string {
  if (item.aweme_id) return item.aweme_id;
  if (item.group_id) return item.group_id;
  if (item.sec_item_id) return item.sec_item_id;
  return index !== undefined ? `${item.author_user_id || "item"}-${index}` : "";
}

/**
 * 把后端 AwemeItem 转成下载页使用的轻量任务。所有入口（视频网格、链接解析、批量翻页）
 * 都走这一份转换，保证清晰度和图文素材的解析结果一致。
 */
export function awemeToDownloadItem(
    item: model.AwemeItem,
    sourceName: string,
    fallbackAuthor: string,
    index?: number,
): DouyinDownloadItem {
  const awemeId = awemeDownloadID(item, index);
  const covers = awemeCoverURLCandidates(item);
  const mediaBadge = douyinMediaBadge(item);
  const videoOptions = douyinVideoOptions(item);
  const selectedVideoOption = defaultDouyinVideoOption(videoOptions);

  return {
    awemeId,
    sourceName,
    title: awemeTitle(item),
    cover: covers[0] ?? "",
    coverCandidates: douyinCoverCandidates(item),
    duration: normalizeDouyinDuration(item.video?.duration ?? item.duration ?? 0),
    authorName: item.author?.nickname || item.author?.uid || fallbackAuthor,
    publishTime: item.create_time ?? 0,
    diggCount: item.statistics?.digg_count ?? 0,
    collectCount: item.statistics?.collect_count ?? 0,
    link: awemeId ? `https://www.douyin.com/video/${awemeId}` : undefined,
    videoURL: selectedVideoOption?.url,
    videoOptions,
    selectedVideoOptionId: selectedVideoOption?.id,
    imageURLs: douyinImageURLs(item),
    assets: mediaBadge ? douyinDownloadAssets(item) : undefined,
    musicURL: mediaBadge ? douyinMusicURL(item) : undefined,
    mediaBadge,
  };
}
