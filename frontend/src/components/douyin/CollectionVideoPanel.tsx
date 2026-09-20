import {createEffect, createMemo, createSignal, type JSXElement, Match, Show, Switch} from "solid-js";
import * as model from "@bindings/github.com/kamiertop/videodown/douyin/model/models";
import DetailError from "../DetailError.tsx";
import DetailLoading from "../DetailLoading.tsx";
import EmptyState from "../EmptyState.tsx";
import CollectionSidebar, {type CollectionSidebarItem} from "./CollectionSidebar.tsx";
import VideoContentPanel, {type DouyinVideoContentKind} from "./VideoContentPanel.tsx";
import {awemeToDownloadItem} from "../../lib/douyin/aweme.ts";
import type {DouyinBatchPageLoader} from "../../lib/douyin/batchDownload.ts";

export type DouyinListItem = model.CollectsList | model.CollectionItem | model.SeriesInfoItem;

export interface ListPage {
  items: readonly DouyinListItem[];
  cursor: number;
  hasMore: boolean;
}

export interface VideoPage {
  items: readonly model.AwemeItem[];
  hasMore: boolean;
}

export type DouyinListVideoKind = "favorite-collection" | "favorite-mix" | "user-mix";

function CollectionPage<T>(props: {
  active: boolean;
  loading: boolean;
  error: string;
  emptyTitle: string;
  onRetry: () => void;
  sidebarItems: readonly CollectionSidebarItem<T>[];
  listTitle: string;
  isSelected: (item: T) => boolean;
  onSelect: (item: T) => void;
  sidebarHasMore: boolean;
  sidebarLoadingMore: boolean;
  onSidebarLoadMore: () => void;
  hasSelection: boolean;
  selectionTitle: string;
  selectionDescription: string;
  contentKind: DouyinVideoContentKind;
  contentLoading: boolean;
  contentError: string;
  onContentRetry: () => void;
  contentTitle: string;
  contentItems: readonly model.AwemeItem[];
  sourceName: string;
  fallbackAuthor: string;
  showToast: (message: string, type?: "success" | "error" | "warning" | "info") => void;
  contentRefreshing?: boolean;
  onContentRefresh?: () => void;
  contentHasMore: boolean;
  contentLoadingMore: boolean;
  onContentLoadMore: () => void;
  /** 一键下载全部的批量会话描述，随选中项变化；没有分页取数器时不传。 */
  contentBatchDownload?: {
    title: string;
    totalCount?: number;
    createLoader: () => DouyinBatchPageLoader;
  };
}): JSXElement {
  return (
    <div classList={{"flex h-full min-h-0 w-full flex-1": props.active, "hidden": !props.active}}>
      <Switch>
        <Match when={props.loading}>
          <DetailLoading/>
        </Match>
        <Match when={!!props.error}>
          <DetailError message={props.error} onRetry={props.onRetry}/>
        </Match>
        <Match when={props.sidebarItems.length === 0}>
          <div class="m-auto">
            <EmptyState title={props.emptyTitle}/>
          </div>
        </Match>
        <Match when={props.sidebarItems.length > 0}>
          <CollectionSidebar
            items={props.sidebarItems}
            listTitle={props.listTitle}
            isSelected={props.isSelected}
            onSelect={props.onSelect}
            hasMore={props.sidebarHasMore}
            loadingMore={props.sidebarLoadingMore}
            onLoadMore={props.onSidebarLoadMore}
          />
          <main class="flex min-w-0 flex-1 flex-col overflow-hidden">
            <Show
              when={props.hasSelection}
              fallback={<EmptyState title={props.selectionTitle} description={props.selectionDescription}/>}
            >
              <VideoContentPanel
                kind={props.contentKind}
                loading={props.contentLoading}
                error={props.contentError}
                onRetry={props.onContentRetry}
                title={props.contentTitle}
                items={props.contentItems}
                sourceName={props.sourceName}
                fallbackAuthor={props.fallbackAuthor}
                showToast={props.showToast}
                refreshing={props.contentRefreshing}
                onRefresh={props.onContentRefresh}
                hasMore={props.contentHasMore}
                loadingMore={props.contentLoadingMore}
                onLoadMore={props.onContentLoadMore}
                batchDownload={props.contentBatchDownload}
              />
            </Show>
          </main>
        </Match>
      </Switch>
    </div>
  );
}

function listTitle(kind: DouyinListVideoKind): string {
  switch (kind) {
    case "favorite-collection":
      return "收藏夹";
    case "favorite-mix":
      return "收藏合集";
    case "user-mix":
      return "全部合集";
  }
}

function listEmptyTitle(kind: DouyinListVideoKind): string {
  switch (kind) {
    case "favorite-collection":
      return "暂无收藏夹";
    case "favorite-mix":
      return "暂无收藏合集";
    case "user-mix":
      return "暂无合集";
  }
}

function selectionTitle(kind: DouyinListVideoKind): string {
  return kind === "favorite-collection" ? "选择一个收藏夹" : "选择一个合集";
}

function selectionDescription(kind: DouyinListVideoKind): string {
  return kind === "favorite-collection"
    ? "右侧将展示收藏夹内视频列表。"
    : "右侧将展示合集内视频列表。";
}

const EMPTY_LIST: readonly never[] = [];

function isCollection(item: DouyinListItem): item is model.CollectsList {
  return "collects_id_str" in item || "collects_name" in item;
}

function isUserSeries(item: DouyinListItem): item is model.SeriesInfoItem {
  return "series_id" in item;
}

function itemId(item: DouyinListItem): string {
  if (isCollection(item)) return item.collects_id_str?.trim() || "";
  if (isUserSeries(item)) return item.series_id || "";
  return item.mix_id || "";
}

function itemKey(item: DouyinListItem, index: number): string {
  return itemId(item) || `list-item-${index}`;
}

function itemTitle(item: DouyinListItem): string {
  if (isCollection(item)) return item.collects_name || "未命名收藏夹";
  if (isUserSeries(item)) return item.series_name || item.real_name || "未命名合集";
  return item.mix_name || "未命名合集";
}

function itemCover(item: DouyinListItem): string {
  if (isCollection(item)) return item.collects_cover?.url_list?.[0] ?? "";
  return item.cover_url?.url_list?.[0] ?? "";
}

function itemSubtitle(item: DouyinListItem): string {
  if (isCollection(item)) return item.user_info?.nickname || item.user_id_str || "未知用户";
  return item.author?.nickname || item.author?.uid || "未知作者";
}

function itemCount(item: DouyinListItem): number {
  if (isCollection(item)) return item.total_number ?? 0;
  if (isUserSeries(item)) return item.stats?.updated_to_episode ?? item.stats?.total_episode ?? 0;
  return item.statis?.updated_to_episode ?? item.statis?.has_updated_episode ?? 0;
}

function isSameItem(a: DouyinListItem, b: DouyinListItem): boolean {
  return itemId(a) === itemId(b);
}

export default function CollectionVideoPanel(props: {
  active: boolean;
  kind: DouyinListVideoKind;
  sourceKey: string;
  refreshKey?: unknown;
  showToast: (message: string, type?: "success" | "error" | "warning" | "info") => void;
  loadList: (cursor: number) => Promise<ListPage>;
  loadVideos: (item: DouyinListItem, cursor: number) => Promise<VideoPage>;
  /**
   * 构造收藏夹/合集详情的纯分页取数器，供批量下载会话在离开本页后继续翻页。
   * 在点击“一键下载全部”时调用（组件仍挂载），实现里必须把所需 ID 烘焙进闭包。
   */
  createDetailPageFetcher?: (item: DouyinListItem) => (offset: number) => Promise<VideoPage>;
}): JSXElement {
  const [listLoading, setListLoading] = createSignal(false);
  const [listError, setListError] = createSignal("");
  const [items, setItems] = createSignal<readonly DouyinListItem[]>(EMPTY_LIST);
  const [listCursor, setListCursor] = createSignal(0);
  const [listHasMore, setListHasMore] = createSignal(false);
  const [listLoadingMore, setListLoadingMore] = createSignal(false);
  const [loadedOnce, setLoadedOnce] = createSignal(false);

  const [selectedItem, setSelectedItem] = createSignal<DouyinListItem | null>(null);
  const [detailLoading, setDetailLoading] = createSignal(false);
  const [detailError, setDetailError] = createSignal("");
  const [detailVideos, setDetailVideos] = createSignal<readonly model.AwemeItem[]>(EMPTY_LIST);
  const [detailHasMore, setDetailHasMore] = createSignal(false);
  const [detailLoadingMore, setDetailLoadingMore] = createSignal(false);

  let listRequestSeq = 0;
  let detailRequestSeq = 0;
  let loadedSourceKey = "";
  let observedRefreshKey: unknown = props.refreshKey;
  let autoLoadingList = false;

  function resetState(): void {
    setItems(EMPTY_LIST);
    setListCursor(0);
    setListHasMore(false);
    setListLoadingMore(false);
    setLoadedOnce(false);
    setSelectedItem(null);
    setDetailVideos(EMPTY_LIST);
    setDetailError("");
    setDetailHasMore(false);
    setDetailLoadingMore(false);
    detailRequestSeq += 1;
  }

  async function loadItems(append = false): Promise<void> {
    if (append) setListLoadingMore(true);
    else {
      setListLoading(true);
      setListError("");
      setListHasMore(false);
      setListLoadingMore(false);
    }

    const seq = ++listRequestSeq;
    try {
      const page = await props.loadList(append ? listCursor() : 0);
      if (seq !== listRequestSeq) return;
      const nextItems = page.items;
      const merged = append ? [...items(), ...nextItems] : nextItems;
      setItems(merged);
      setListCursor(page.cursor);
      setListHasMore(page.hasMore);
      setLoadedOnce(true);

      if (!append) {
        const current = selectedItem();
        const nextSelected = current
          ? nextItems.find((item) => isSameItem(current, item))
          : undefined;
        if (nextSelected) void loadDetail(nextSelected);
      }
    } catch (error) {
      if (seq !== listRequestSeq) return;
      const msg = error instanceof Error ? error.message : String(error);
      if (append) props.showToast(`加载更多${listTitle(props.kind)}失败: ${msg}`, "warning");
      else setListError(msg);
    } finally {
      if (seq === listRequestSeq) {
        if (append) setListLoadingMore(false);
        else setListLoading(false);
      }
    }
  }

  async function loadDetail(item: DouyinListItem, append = false): Promise<void> {
    // Solid setter 接收函数有特殊语义；这里用函数形式避免泛型 T 被误判成 updater。
    setSelectedItem(() => item);
    if (append) setDetailLoadingMore(true);
    else {
      setDetailLoading(true);
      setDetailError("");
      setDetailVideos(EMPTY_LIST);
      setDetailHasMore(false);
      setDetailLoadingMore(false);
    }

    const seq = ++detailRequestSeq;
    try {
      const page = await props.loadVideos(item, append ? detailVideos().length : 0);
      if (seq !== detailRequestSeq) return;
      const nextVideos = page.items;
      setDetailVideos(append ? [...detailVideos(), ...nextVideos] : nextVideos);
      setDetailHasMore(nextVideos.length > 0 && page.hasMore);
    } catch (error) {
      if (seq !== detailRequestSeq) return;
      const msg = error instanceof Error ? error.message : String(error);
      if (append) props.showToast(`加载更多视频失败: ${msg}`, "warning");
      else setDetailError(msg);
    } finally {
      if (seq === detailRequestSeq) {
        if (append) setDetailLoadingMore(false);
        else setDetailLoading(false);
      }
    }
  }

  async function loadAllListItems(): Promise<void> {
    if (autoLoadingList) return;
    autoLoadingList = true;
    try {
      while (listHasMore()) {
        const before = items().length;
        await loadItems(true);
        if (items().length === before) break;
      }
    } finally {
      autoLoadingList = false;
    }
  }

  // 批量下载描述随选中项变化；createLoader 在点击时才执行，此时组件仍挂载，
  // 可以安全读取当前分页状态并烘焙成闭包值。
  const detailBatchDownload = createMemo(() => {
    const item = selectedItem();
    const fetcherFactory = props.createDetailPageFetcher;
    if (!item || !fetcherFactory) return undefined;

    return {
      title: itemTitle(item),
      totalCount: itemCount(item) || undefined,
      createLoader: (): DouyinBatchPageLoader => {
        const fetchPage = fetcherFactory(item);
        const sourceName = itemTitle(item);
        const fallbackAuthor = itemSubtitle(item);
        let offset = detailVideos().length;
        let hasMore = detailHasMore();
        return {
          hasMore: () => hasMore,
          loadNext: async () => {
            const page = await fetchPage(offset);
            offset += page.items.length;
            hasMore = page.items.length > 0 && page.hasMore;
            return page.items.map((aweme, index) => awemeToDownloadItem(aweme, sourceName, fallbackAuthor, index));
          },
        };
      },
    };
  });

  createEffect(() => {
    if (!props.active) return;
    if (loadedSourceKey !== props.sourceKey) {
      loadedSourceKey = props.sourceKey;
      resetState();
    }
    if (observedRefreshKey !== props.refreshKey) {
      observedRefreshKey = props.refreshKey;
      resetState();
    }
    if (!loadedOnce()) {
      void loadItems().then(() => void loadAllListItems());
    }
  });

  const sidebarItems = createMemo<CollectionSidebarItem<DouyinListItem>[]>(() =>
    items().map((item, index) => ({
      key: itemKey(item, index),
      title: itemTitle(item),
      cover: itemCover(item),
      subtitle: itemSubtitle(item),
      count: itemCount(item),
      raw: item,
    })),
  );

  function isSelected(item: DouyinListItem): boolean {
    const selected = selectedItem();
    return !!selected && isSameItem(selected, item);
  }

  function selectItem(item: DouyinListItem): void {
    const selected = selectedItem();
    if (selected && isSameItem(selected, item) && detailVideos().length > 0 && !detailLoading()) return;
    void loadDetail(item);
  }

  return (
    <CollectionPage
      active={props.active}
      loading={listLoading()}
      error={listError()}
      emptyTitle={listEmptyTitle(props.kind)}
      onRetry={() => void loadItems()}
      sidebarItems={sidebarItems()}
      listTitle={listTitle(props.kind)}
      isSelected={isSelected}
      onSelect={selectItem}
      sidebarHasMore={listHasMore()}
      sidebarLoadingMore={listLoadingMore()}
      onSidebarLoadMore={() => void loadItems(true)}
      hasSelection={!!selectedItem()}
      selectionTitle={selectionTitle(props.kind)}
      selectionDescription={selectionDescription(props.kind)}
      contentKind={props.kind as DouyinVideoContentKind}
      contentLoading={detailLoading()}
      contentError={detailError()}
      onContentRetry={() => {
        const item = selectedItem();
        if (item) void loadDetail(item);
      }}
      contentTitle={selectedItem() ? itemTitle(selectedItem()!) : listTitle(props.kind)}
      contentItems={detailVideos()}
      sourceName={selectedItem() ? itemTitle(selectedItem()!) : listTitle(props.kind)}
      fallbackAuthor={selectedItem() ? itemSubtitle(selectedItem()!) : "未知作者"}
      showToast={props.showToast}
      contentRefreshing={detailLoading()}
      onContentRefresh={() => {
        const item = selectedItem();
        if (item) void loadDetail(item);
      }}
      contentHasMore={detailHasMore()}
      contentLoadingMore={detailLoadingMore()}
      onContentLoadMore={() => {
        const item = selectedItem();
        if (item) void loadDetail(item, true);
      }}
      contentBatchDownload={detailBatchDownload()}
    />
  );
}
