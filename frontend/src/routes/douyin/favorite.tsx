import {createFileRoute} from '@tanstack/solid-router'
import {createEffect, createSignal, For, type JSXElement, onCleanup, onMount,} from "solid-js";
import {
  Collection as FavoriteMixCollection,
  CollectionList,
  CollectList,
  FavoritesVideoList,
  FavoriteVideo,
} from "../../../wailsjs/go/api/Douyin";
import {model} from "../../../wailsjs/go/models";
import CollectionVideoPanel, {
  type DouyinListItem,
  type ListPage,
  type VideoPage,
} from "../../components/douyin/CollectionVideoPanel.tsx";
import VideoContentPanel from "../../components/douyin/VideoContentPanel.tsx";
import Toast from "../../components/Toast.tsx";
import {useToast} from "../../hooks/useToast.ts";

export const Route = createFileRoute('/douyin/favorite')({
  component: DouyinFavoritePage,
})

type FavoriteTab = 'collection' | 'video' | 'mix';

const MIX_PAGE_SIZE = 12;
const MIX_VIDEO_PAGE_SIZE = 10;

type DouyinMixItem = model.CollectionItem | model.SeriesInfoItem;

function collectionId(item: model.CollectsList | null | undefined): string {
  return item?.collects_id_str?.trim() || "";
}

function isUserSeries(item: DouyinMixItem): item is model.SeriesInfoItem {
  return "series_id" in item;
}

function mixId(item: DouyinMixItem): string {
  return isUserSeries(item) ? item.series_id : item.mix_id;
}

async function loadCollections(cursor: number): Promise<ListPage> {
  const data = await CollectList(cursor);
  const items = data.collects_list ?? [];
  return {
    items,
    cursor: data.cursor ?? cursor + items.length,
    hasMore: Boolean(data.has_more),
  };
}

async function loadCollectionVideos(item: DouyinListItem, cursor: number): Promise<VideoPage> {
  const id = collectionId(item as model.CollectsList);
  if (!id) throw new Error("收藏夹 ID 无效");

  const data = await FavoritesVideoList(id, cursor, 10);  // 默认每页10个视频
  return {
    items: data.aweme_list ?? [],
    hasMore: Number(data.has_more ?? 0) > 0,
  };
}

async function loadFavoriteMixes(cursor: number): Promise<ListPage> {
  const data = await FavoriteMixCollection(MIX_PAGE_SIZE, cursor);
  const items = data.mix_infos ?? [];
  return {
    items,
    cursor: data.cursor ?? cursor + items.length,
    hasMore: Number(data.has_more ?? 0) > 0,
  };
}

async function loadFavoriteMixVideos(item: DouyinListItem, cursor: number): Promise<VideoPage> {
  const mixItem = item as DouyinMixItem;
  const seriesID = mixId(mixItem);
  if (!seriesID) throw new Error("合集 ID 无效");

  const data = await CollectionList(mixItem.author?.sec_uid || "", seriesID, cursor, MIX_VIDEO_PAGE_SIZE);
  return {
    items: data.aweme_list ?? [],
    hasMore: Number(data.has_more ?? 0) > 0,
  };
}

function tabLabel(tab: FavoriteTab): string {
  const labels: Record<FavoriteTab, string> = {
    collection: "收藏夹",
    video: "视频",
    mix: "合集",
  };

  return labels[tab];
}

function FavoriteHeader(props: {
  activeTab: FavoriteTab;
  onTabChange: (tab: FavoriteTab) => void;
}): JSXElement {
  return (
    <header class="mb-2 shrink-0">
      <nav
        class="flex min-w-0 flex-row justify-between rounded-2xl border border-base-300/90 bg-linear-to-b from-base-100 via-base-100 to-base-200/20 p-1.5 shadow-sm"
        role="tablist"
        aria-label="收藏分类"
      >
        <For each={['collection', 'video', 'mix'] as const}>
          {(tab) => (
            <button
              class={'min-h-6 flex-1 rounded-xl px-2 text-center text-xs font-semibold sm:text-sm'}
              classList={{
                "bg-primary/12 text-primary shadow-sm ring-1 ring-primary/25": props.activeTab === tab,
                "text-base-content/50 hover:bg-base-200/80 hover:text-base-content": props.activeTab !== tab,
                "opacity-90": tab === 'mix',
              }}
              type="button"
              role="tab"
              aria-selected={props.activeTab === tab}
              onClick={() => props.onTabChange(tab)}
            >
              {tabLabel(tab)}
            </button>
          )}
        </For>
      </nav>
    </header>
  );
}

function DouyinFavoritePage(): JSXElement {
  const [activeTab, setActiveTab] = createSignal<FavoriteTab>('collection');
  const [videoLoading, setVideoLoading] = createSignal(false);
  const [videoError, setVideoError] = createSignal("");
  const [videoLoadingMore, setVideoLoadingMore] = createSignal(false);
  const [videoLoaded, setVideoLoaded] = createSignal(false);
  const [videoResult, setVideoResult] = createSignal<model.FavoriteVideoResponse>();
  const {message, type, showToast} = useToast();
  let videoRequestSeq = 0;

  const videos = () => videoResult()?.aweme_list ?? [];

  function switchTab(event: KeyboardEvent): void {
    if (event.key === 'ArrowLeft') {
      event.preventDefault();
      setActiveTab((prev) => {
        if (prev === 'video') return 'collection';
        if (prev === 'mix') return 'video';
        return 'collection';
      });
      return;
    }

    if (event.key === 'ArrowRight') {
      event.preventDefault();
      setActiveTab((prev) => {
        if (prev === 'collection') return 'video';
        if (prev === 'video') return 'mix';
        return 'mix';
      });
    }
  }

  onMount(() => {
    window.addEventListener("keydown", switchTab);
  });
  onCleanup(() => {
    window.removeEventListener("keydown", switchTab);
  });

  createEffect(() => {
    if (activeTab() === 'video' && !videoLoaded() && !videoLoading()) {
      void loadFavoriteVideos();
    }
  });

  function videoHasMore(): boolean {
    return Number(videoResult()?.has_more ?? 0) > 0;
  }

  async function loadFavoriteVideos(): Promise<void> {
    const seq = ++videoRequestSeq;
    setVideoLoading(true);
    setVideoError("");
    try {
      const data = await FavoriteVideo(10, 0);
      if (seq !== videoRequestSeq) return;
      setVideoResult(data);
      setVideoLoaded(true);
    } catch (error) {
      if (seq !== videoRequestSeq) return;
      setVideoError(error instanceof Error ? error.message : String(error));
      setVideoLoaded(true);
    } finally {
      if (seq === videoRequestSeq) setVideoLoading(false);
    }
  }

  async function loadMoreVideos(): Promise<void> {
    const current = videoResult();
    if (!current || Number(current.has_more ?? 0) <= 0 || videoLoadingMore()) {
      return;
    }

    setVideoLoadingMore(true);
    try {
      const next = await FavoriteVideo(20, current.cursor ?? videos().length);
      setVideoResult(model.FavoriteVideoResponse.createFrom({
        ...next,
        aweme_list: [...videos(), ...(next.aweme_list ?? [])],
        uid: current.uid ?? next.uid,
        sec_uid: current.sec_uid ?? next.sec_uid,
      }));
    } catch (error) {
      showToast(error instanceof Error ? error.message : String(error), 'warning');
    } finally {
      setVideoLoadingMore(false);
    }
  }

  return (
    <section class="flex h-full min-h-0 flex-col p-2">
      <FavoriteHeader
        activeTab={activeTab()}
        onTabChange={setActiveTab}
      />

      <div class="min-h-0 flex-1 overflow-hidden rounded-xl border border-base-300 bg-base-100 shadow-sm">
        <div class="flex h-full min-h-0 flex-col">
          <CollectionVideoPanel
            active={activeTab() === 'collection'}
            kind="favorite-collection"
            sourceKey="favorite-collections"
            showToast={showToast}
            loadList={loadCollections}
            loadVideos={loadCollectionVideos}
          />

          <div classList={{
            "flex h-full min-h-0 w-full min-w-0 flex-1 flex-col overflow-hidden": activeTab() === 'video',
            "hidden": activeTab() !== 'video',
          }}>
            <VideoContentPanel
              kind="favorite-video"
              loading={videoLoading()}
              error={videoError()}
              onRetry={() => void loadFavoriteVideos()}
              items={videos()}
              sourceName="收藏视频"
              fallbackAuthor="未知作者"
              showToast={showToast}
              refreshing={videoLoading()}
              onRefresh={() => void loadFavoriteVideos()}
              hasMore={videoHasMore()}
              loadingMore={videoLoadingMore()}
              onLoadMore={() => void loadMoreVideos()}
            />
          </div>

          <CollectionVideoPanel
            active={activeTab() === 'mix'}
            kind="favorite-mix"
            sourceKey="favorite-mixes"
            showToast={showToast}
            loadList={loadFavoriteMixes}
            loadVideos={loadFavoriteMixVideos}
          />
        </div>
      </div>
      <Toast message={message()} type={type()}/>
    </section>
  );
}
