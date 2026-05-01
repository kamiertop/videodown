import type {JSXElement} from "solid-js";
import {CollectList, FavoritesVideoList} from "../../../wailsjs/go/api/Douyin";
import {model} from "../../../wailsjs/go/models";
import ListVideoPanel, {type DouyinListItem, type ListPage, type VideoPage} from "./ListVideoPanel.tsx";
import CollectsList = model.CollectsList;

const COLLECTION_VIDEO_PAGE_SIZE = 20;

function collectionId(item: model.CollectsList | null | undefined): string {
  // 收藏夹 ID 可能超过 JS 安全整数范围，必须优先用后端返回的字符串 ID。
  return item?.collects_id_str?.trim() || "";
}

async function loadCollections(cursor: number): Promise<ListPage> {
  // 适配层把 Douyin 原始响应统一成 ListVideoPanel 认识的 ListPage。
  const data = await CollectList(cursor);
  const items: CollectsList[] = data.collects_list ?? [];
  return {
    items,
    cursor: data.cursor ?? cursor + items.length,
    hasMore: Boolean(data.has_more),
  };
}

async function loadCollectionVideos(item: DouyinListItem, cursor: number): Promise<VideoPage> {
  const id = collectionId(item as model.CollectsList);
  if (!id) throw new Error("收藏夹 ID 无效");

  // 只在用户点击某个收藏夹后调用；首次进入只拉左侧列表。
  const data = await FavoritesVideoList(id, cursor, COLLECTION_VIDEO_PAGE_SIZE);
  return {
    items: data.aweme_list ?? [],
    hasMore: Number(data.has_more ?? 0) > 0,
  };
}

// 收藏夹只是 ListVideoPanel 的接口适配层：列表取收藏夹，详情取该收藏夹内的视频。
export default function FavoriteCollectionPanel(props: {
  active: boolean;
  showToast: (message: string, type?: "success" | "error" | "warning" | "info") => void;
}): JSXElement {
  return (
    <ListVideoPanel
      active={props.active}
      kind="favorite-collection"
      // 收藏夹列表与用户无关，固定 sourceKey；刷新逻辑由 ListVideoPanel 管
      sourceKey="favorite-collections"
      showToast={props.showToast}
      loadList={loadCollections}
      loadVideos={loadCollectionVideos}
    />
  );
}
