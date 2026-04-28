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
// 合集内容分页大小
const COLLECTION_ITEM_PAGE_SIZE = 20;
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
  const [collectionPage, setCollectionPage] = createSignal(COLLECTION_PAGE);
  const [collectionHasMore, setCollectionHasMore] = createSignal(false);
  const [collectionLoadingMore, setCollectionLoadingMore] = createSignal(false);
  const [selectedItem, setSelectedItem] = createSignal<model.CollectionDataList | null>(null);
  const [detailLoading, setDetailLoading] = createSignal(false);
  const [detailError, setDetailError] = createSignal('');
  const [detail, setDetail] = createSignal<model.CollectionItemData | null>(null);
  const [detailPage, setDetailPage] = createSignal(COLLECTION_PAGE);
  const [detailHasMore, setDetailHasMore] = createSignal(false);
  const [loadingMore, setLoadingMore] = createSignal(false);
  const [loadedOnce, setLoadedOnce] = createSignal(false);
  const [detailVersion, setDetailVersion] = createSignal(0);

  let collectionRequestSeq = 0;
  let detailRequestSeq = 0;

  const sidebarItems = () => toColSidebarItems(collections()?.list ?? EMPTY_COLLECTION_LIST);
  const mediaCards = () => toCollectionMediaCards(detail()?.medias ?? [], selectedItem()?.title ?? "");

  const updateDetailHasMore = (data: model.CollectionItemData | null, receivedCount: number) => {
    const total = data?.info?.media_count ?? selectedItem()?.media_count ?? 0;
    const loaded = data?.medias?.length ?? 0;
    setDetailHasMore(receivedCount > 0 && loaded < total);
  };

  const loadCollectionDetail = async (item: model.CollectionDataList, append = false) => {
    const targetPage = append ? detailPage() + 1 : COLLECTION_PAGE;
    setSelectedItem(item);
    if (append) {
      setLoadingMore(true);
    } else {
      setLoadingMore(false);
      setDetailVersion(prev => prev + 1);
      setDetailLoading(true);
      setDetailError('');
      setDetail(null);
      setDetailPage(COLLECTION_PAGE);
      setDetailHasMore(false);
    }

    const seq = ++detailRequestSeq;
    try {
      const data = await CollectionItem(String(item.id), targetPage, COLLECTION_ITEM_PAGE_SIZE);
      if (seq !== detailRequestSeq) return;
      const receivedCount = data.medias?.length ?? 0;
      if (append) {
        const prevMedias = detail()?.medias ?? [];
        const next = model.CollectionItemData.createFrom({...data, medias: [...prevMedias, ...(data.medias ?? [])]});
        setDetail(next);
        updateDetailHasMore(next, receivedCount);
      } else {
        setDetail(data);
        updateDetailHasMore(data, receivedCount);
      }
      setDetailPage(targetPage);
    } catch (error) {
      if (seq !== detailRequestSeq) return;
      const msg = error instanceof Error ? error.message : String(error);
      if (append) props.showToast(`加载更多失败: ${msg}`, 'warning');
      else setDetailError(msg);
    } finally {
      if (seq === detailRequestSeq) {
        if (append) setLoadingMore(false);
        else setDetailLoading(false);
      }
    }
  };

  const loadCollections = async (append = false) => {
    const targetPage = append ? collectionPage() + 1 : COLLECTION_PAGE;
    if (append) {
      setCollectionLoadingMore(true);
    } else {
      setLoading(true);
      setCollectionLoadingMore(false);
      setCollectionPage(COLLECTION_PAGE);
      setCollectionHasMore(false);
    }

    const seq = ++collectionRequestSeq;
    try {
      const data = await Collection(targetPage, COLLECTION_PAGE_SIZE);
      if (seq !== collectionRequestSeq) return;
      if (append) {
        setCollections(prev => model.CollectionData.createFrom({
          ...data,
          list: [...(prev?.list ?? []), ...(data.list ?? [])],
        }));
      } else {
        setCollections(data);
      }
      setCollectionPage(targetPage);
      setCollectionHasMore(data.has_more);
      setLoadedOnce(true);

      if (append) return;

      const current = selectedItem()?.id;
      const nextItem = data.list?.find(item => item.id === current) ?? data.list?.[0];
      if (nextItem) {
        void loadCollectionDetail(nextItem);
      } else {
        setSelectedItem(null);
        setDetail(null);
        setDetailError('');
        setDetailPage(COLLECTION_PAGE);
        setDetailHasMore(false);
        setDetailVersion(prev => prev + 1);
      }
    } catch (error) {
      props.showToast(error instanceof Error ? error.message : String(error), append ? 'warning' : 'error');
    } finally {
      if (seq === collectionRequestSeq) {
        if (append) setCollectionLoadingMore(false);
        else setLoading(false);
      }
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
        sidebarIcon={<StackIcon />}
        sidebarLabel={'合集'}
        sidebarCount={() => collections()?.count ?? sidebarItems().length}
        sidebarLoading={loading}
        onRefresh={() => void loadCollections()}
        sidebarHasMore={collectionHasMore}
        sidebarLoadingMore={collectionLoadingMore}
        onLoadMoreSidebar={() => {
          if (!collectionHasMore() || collectionLoadingMore()) return;
          void loadCollections(true);
        }}
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
        hasMore={detailHasMore}
        loadingMore={loadingMore}
        onLoadMore={() => {
          const item = selectedItem();
          if (!item || !detailHasMore() || loadingMore()) return;
          void loadCollectionDetail(item, true);
        }}
      />
    </div>
  );
}
