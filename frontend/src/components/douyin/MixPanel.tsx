import type {JSXElement} from "solid-js";
import {Collection as FavoriteMixCollection, CollectionList, UserSeries} from "../../../wailsjs/go/api/Douyin";
import {model} from "../../../wailsjs/go/models";
import ListVideoPanel, {type DouyinListItem, type ListPage, type VideoPage} from "./ListVideoPanel.tsx";

const MIX_PAGE_SIZE = 12;
const MIX_VIDEO_PAGE_SIZE = 20;

type DouyinMixItem = model.CollectionItem | model.SeriesInfoItem;

function isUserSeries(item: DouyinMixItem): item is model.SeriesInfoItem {
  // 收藏合集接口返回 mix_id，用户合集接口返回 series_id；用字段存在性区分类型。
  return "series_id" in item;
}

function mixId(item: DouyinMixItem): string {
  return isUserSeries(item) ? item.series_id : item.mix_id;
}

async function loadFavoriteMixes(cursor: number): Promise<ListPage> {
  // 我的收藏合集列表。
  const data = await FavoriteMixCollection(MIX_PAGE_SIZE, cursor);
  const items = data.mix_infos ?? [];
  return {
    items,
    cursor: data.cursor ?? cursor + items.length,
    hasMore: Number(data.has_more ?? 0) > 0,
  };
}

async function loadUserMixes(secUserId: string, cursor: number): Promise<ListPage> {
  // 某个用户主页下的合集列表。
  const data = await UserSeries(secUserId, cursor, MIX_PAGE_SIZE);
  const items = data.series_infos ?? [];
  return {
    items,
    cursor: data.cursor ?? cursor + items.length,
    hasMore: Number(data.has_more ?? 0) > 0,
  };
}

async function loadMixVideos(secUserId: string | undefined, item: DouyinListItem, cursor: number): Promise<VideoPage> {
  const mixItem = item as DouyinMixItem;
  const seriesID = mixId(mixItem);
  if (!seriesID) throw new Error("合集 ID 无效");

  // 收藏合集和用户合集详情使用同一个 series/aweme 接口，只是 referer 所需 secUserId 来源不同。
  const data = await CollectionList(secUserId || mixItem.author?.sec_uid || "", seriesID, cursor, MIX_VIDEO_PAGE_SIZE);
  return {
    items: data.aweme_list ?? [],
    hasMore: Number(data.has_more ?? 0) > 0,
  };
}

// 收藏合集和用户合集共用同一个面板，只在列表接口、sourceKind 和文案上有差异。
export default function MixPanel(props: {
  active: boolean;
  showToast: (message: string, type?: "success" | "error" | "warning" | "info") => void;
  userSecUserId?: string;
  refreshKey?: unknown;
}): JSXElement {
  const isUserSource = () => !!props.userSecUserId;

  return (
    <ListVideoPanel
      active={props.active}
      // kind 决定 UI 文案和下载 sourceKind；接口差异留在本适配层处理。
      kind={isUserSource() ? "user-mix" : "favorite-mix"}
      sourceKey={isUserSource() ? `user:${props.userSecUserId}` : "favorite-mixes"}
      refreshKey={props.refreshKey}
      showToast={props.showToast}
      loadList={(cursor) => props.userSecUserId ? loadUserMixes(props.userSecUserId, cursor) : loadFavoriteMixes(cursor)}
      loadVideos={(item, cursor) => loadMixVideos(props.userSecUserId, item, cursor)}
    />
  );
}
