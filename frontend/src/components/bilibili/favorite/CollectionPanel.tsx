import {createEffect, createSignal, type JSXElement} from "solid-js";
import {Collection, CollectionItem} from "../../../../wailsjs/go/api/BiliBili";
import {model} from "../../../../wailsjs/go/models";
import type {MediaCardItem} from "../../../lib/model.ts";
import {StackIcon} from "../../icons/IconStack";
import {type SidebarListItem} from "../../SidebarList";
import FavoriteCollectionView from "./FavoriteCollectionView";

const COLLECTION_PAGE = 1;
// 合集列表分页大小
const COLLECTION_PAGE_SIZE = 20;
const EMPTY_COLLECTION_LIST: readonly model.CollectionDataList[] = [];

interface ColSidebarItem extends SidebarListItem {
  raw: model.CollectionDataList;
}

function toColSidebarItems(list: readonly model.CollectionDataList[]): ColSidebarItem[] {
  return list.map(item => ({
    id: item.id,
    title: item.title,
    count: item.media_count,
    subtitle: item.upper?.name,
    raw: item,
  }));
}

function toCollectionMediaCards(medias: model.CollectionItemMedias[], listName: string): MediaCardItem[] {
  const name = listName.trim();
  return medias.map(media => ({
    id: Number(media.id),
    title: media.title,
    cover: media.cover,
    duration: media.duration,
    bvid: media.bvid,
    upperName: media.upper?.name || '未知',
    play: media.cnt_info?.play,
    danmaku: media.cnt_info?.danmaku,
    pubtime: media.pubtime,
    sourceListName: name,
    sourceListKind: '合集',
  }));
}

export default function CollectionPanel(props: {
  active: boolean;
  showToast: (message: string, type?: "success" | "error" | "warning" | "info") => void;
}): JSXElement {
  const [loading, setLoading] = createSignal(false);
  const [collections, setCollections] = createSignal<model.CollectionData | null>(null);
  const [selectedItem, setSelectedItem] = createSignal<model.CollectionDataList | null>(null);
  const [detailLoading, setDetailLoading] = createSignal(false);
  const [detailError, setDetailError] = createSignal('');
  const [detail, setDetail] = createSignal<model.CollectionItemData | null>(null);
  const [loadedOnce, setLoadedOnce] = createSignal(false);
  const [detailVersion, setDetailVersion] = createSignal(0);

  let detailRequestSeq = 0;

  const sidebarItems = () => toColSidebarItems(collections()?.list ?? EMPTY_COLLECTION_LIST);
  const mediaCards = () => toCollectionMediaCards(detail()?.medias ?? [], selectedItem()?.title ?? "");

  const loadCollectionDetail = async (item: model.CollectionDataList) => {
    setSelectedItem(item);
    setDetailVersion(prev => prev + 1);
    setDetailLoading(true);
    setDetailError('');
    setDetail(null);

    const seq = ++detailRequestSeq;
    try {
      const data = await CollectionItem(String(item.id), 1, COLLECTION_PAGE_SIZE);
      if (seq !== detailRequestSeq) return;
      setDetail(data);
    } catch (error) {
      if (seq !== detailRequestSeq) return;
      setDetailError(error instanceof Error ? error.message : String(error));
    } finally {
      if (seq === detailRequestSeq) setDetailLoading(false);
    }
  };

  const loadCollections = async () => {
    setLoading(true);
    try {
      const data = await Collection(COLLECTION_PAGE, COLLECTION_PAGE_SIZE);
      setCollections(data);
      setLoadedOnce(true);

      const current = selectedItem()?.id;
      const nextItem = data.list?.find(item => item.id === current) ?? data.list?.[0];
      if (nextItem) {
        void loadCollectionDetail(nextItem);
      } else {
        setSelectedItem(null);
        setDetail(null);
        setDetailError('');
        setDetailVersion(prev => prev + 1);
      }
    } catch (error) {
      props.showToast(error instanceof Error ? error.message : String(error), 'error');
    } finally {
      setLoading(false);
    }
  };

  createEffect(() => {
    if (props.active && !loadedOnce()) {
      void loadCollections();
    }
  });

  return (
    <div class={props.active ? "flex h-full min-h-0 flex-1" : "hidden"}>
      <FavoriteCollectionView
        sidebarItems={sidebarItems}
        selectedSidebarId={() => selectedItem()?.id ?? null}
        onSelectSidebar={(item) => {
          if (selectedItem()?.id === item.raw.id && detail() && !detailLoading()) return;
          void loadCollectionDetail(item.raw);
        }}
        sidebarIcon={StackIcon()}
        sidebarLabel={'合集'}
        sidebarCount={() => collections()?.count ?? sidebarItems().length}
        sidebarLoading={loading}
        onRefresh={() => void loadCollections()}
        hasSelection={() => selectedItem() !== null}
        detailLoading={detailLoading}
        detailError={detailError}
        onRetryDetail={() => {
          const item = selectedItem();
          if (item) void loadCollectionDetail(item);
        }}
        detailTitle={() => selectedItem()?.title ?? ''}
        detailMediaCount={() => detail()?.info?.media_count ?? selectedItem()?.media_count ?? 0}
        mediaCards={mediaCards}
        detailVersion={detailVersion}
      />
    </div>
  );
}
