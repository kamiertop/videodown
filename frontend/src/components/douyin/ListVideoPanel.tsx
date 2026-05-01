import {createEffect, createSignal, For, type JSXElement, Match, Show, Switch} from "solid-js";
import {model} from "../../../wailsjs/go/models";
import DetailError from "../DetailError.tsx";
import DetailLoading from "../DetailLoading.tsx";
import EmptyState from "../EmptyState.tsx";
import NoCover from "../NoCover.tsx";
import VideoContentPanel, {type DouyinVideoContentKind} from "./VideoContentPanel.tsx";

export type DouyinListItem = model.CollectsList | model.CollectionItem | model.SeriesInfoItem;

export interface ListPage {
  // 左侧列表分页结果。T 可以是收藏夹，也可以是合集条目。
  items: readonly DouyinListItem[];
  // 下一页请求使用的游标；不同接口字段名不完全一致，适配层统一成 cursor。
  cursor: number;
  hasMore: boolean;
}

export interface VideoPage {
  // 右侧详情统一只关心 AwemeItem，下载项转换放到 VideoContentPanel 内。
  items: readonly model.AwemeItem[];
  hasMore: boolean;
}

// 左侧列表 + 右侧视频详情目前只服务这三类集合页。
// 文案、下载 sourceKind 和空态都可以从 kind 推导，避免调用方重复传一堆字符串。
export type DouyinListVideoKind = "favorite-collection" | "favorite-mix" | "user-mix";

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

export interface ListVideoPanelProps {
  active: boolean;
  kind: DouyinListVideoKind;
  // 数据源身份。切换收藏合集/用户合集，或者切换用户时，用它判断是否要清掉旧状态。
  sourceKey: string;
  // 手动刷新用。值变化时强制重新拉左侧列表，但组件不关心具体值是什么。
  refreshKey?: unknown;
  showToast: (message: string, type?: "success" | "error" | "warning" | "info") => void;
  // 只拉左侧列表，不自动拉第一项详情；详情请求等用户点击列表项后触发。
  loadList: (cursor: number) => Promise<ListPage>;
  loadVideos: (item: DouyinListItem, cursor: number) => Promise<VideoPage>;
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

// 只服务 ListVideoPanel 的私有布局：左侧列表负责选集合，右侧复用 VideoContentPanel 展示视频。
function SidebarDetailLayout(props: {
  active: boolean;
  loading: boolean;
  error: string;
  emptyTitle: string;
  onRetry: () => void;
  items: readonly DouyinListItem[];
  listTitle: string;
  isSelected: (item: DouyinListItem) => boolean;
  onSelect: (item: DouyinListItem) => void;
  hasMore: boolean;
  loadingMore: boolean;
  onLoadMore: () => void;
  hasSelection: boolean;
  selectionTitle: string;
  selectionDescription: string;
  children: JSXElement;
}): JSXElement {
  return (
    // active=false 时保留组件实例但隐藏 DOM，避免 tab 切换丢掉已加载列表。
    <div classList={{"flex h-full min-h-0 flex-1": props.active, "hidden": !props.active}}>
      <Switch>
        <Match when={props.loading}>
          {/* 左侧列表首次加载中，整块区域用统一加载态占位。 */}
          <DetailLoading/>
        </Match>
        <Match when={!!props.error}>
          {/* 左侧列表加载失败；重试只重新拉左侧列表，不触碰下载队列。 */}
          <DetailError message={props.error} onRetry={props.onRetry}/>
        </Match>
        <Match when={props.items.length === 0}>
          {/* 左侧列表为空，右侧没有可选对象，所以只展示列表空态。 */}
          <div class="m-auto">
            <EmptyState title={props.emptyTitle}/>
          </div>
        </Match>
        <Match when={props.items.length > 0}>
          {/* 左侧只负责集合列表和分页；右侧详情不在这里写死，由 children 注入。 */}
          <aside class="flex w-48 shrink-0 flex-col border-r border-base-300 bg-base-100">
            {/* 左侧列表头：固定在上方，展示集合类型和当前已加载数量。 */}
            <div class="flex shrink-0 items-center justify-between gap-2 border-b border-base-200 px-3 py-2">
              <div class="min-w-0">
                <h3 class="truncate text-sm font-bold text-base-content">{props.listTitle}</h3>
                <p class="text-xs text-base-content/50">{props.items.length} 个已加载</p>
              </div>
            </div>

            {/* 左侧列表主体：只有这里滚动，右侧视频网格有自己的滚动容器。 */}
            <div class="min-h-0 flex-1 overflow-auto">
              <For each={props.items}>
                {(item, index) => {
                  // selected 包成函数，确保 Solid 能在选中项变化时更新当前行样式。
                  const selected = () => props.isSelected(item);
                  return (
                    <button
                      type="button"
                      class="flex w-full gap-3 border-b border-base-200/80 p-3 text-left transition-colors"
                      classList={{"bg-primary/10": selected(), "hover:bg-base-200/35": !selected()}}
                      data-key={itemKey(item, index())}
                      onClick={() => props.onSelect(item)}
                    >
                      {/* 左侧条目统一布局：封面 + 标题 + 副标题 + 数量。字段差异由 kind/item 类型推导。 */}
                      <div class="h-16 w-16 shrink-0 overflow-hidden rounded-xl border border-base-300 bg-base-200">
                        <Show when={itemCover(item)} fallback={<NoCover/>}>
                          <img
                            src={itemCover(item)}
                            alt=""
                            class="h-full w-full object-cover"
                            loading="lazy"
                            decoding="async"
                            referrerPolicy="no-referrer"
                          />
                        </Show>
                      </div>
                      <div class="min-w-0 flex-1 self-center">
                        <h4 class={`line-clamp-2 text-sm font-semibold ${selected() ? "text-primary" : "text-base-content"}`}>
                          {itemTitle(item)}
                        </h4>
                        <p class="mt-1 truncate text-xs text-base-content/50">
                          {itemSubtitle(item)}
                        </p>
                        <p class="mt-1 text-xs text-base-content/45">
                          {itemCount(item)} 个作品
                        </p>
                      </div>
                    </button>
                  );
                }}
              </For>
            </div>

            <Show when={props.hasMore}>
              {/* 左侧列表分页按钮固定在侧栏底部，不随列表内容滚走。 */}
              <div class="shrink-0 border-t border-base-200 p-2">
                <button
                  class="btn btn-ghost btn-sm h-9 w-full"
                  type="button"
                  onClick={props.onLoadMore}
                  disabled={props.loadingMore}
                >
                  <Show
                    when={!props.loadingMore}
                    fallback={<span class="loading loading-spinner loading-sm text-primary"/>}
                  >
                    加载更多
                  </Show>
                </button>
              </div>
            </Show>
          </aside>

          {/* 右侧详情区域：未选择时显示提示，选择后渲染调用方传入的视频面板。 */}
          <main class="flex min-w-0 flex-1 flex-col overflow-hidden">
            <Show
              when={props.hasSelection}
              fallback={<EmptyState title={props.selectionTitle} description={props.selectionDescription}/>}
            >
              {props.children}
            </Show>
          </main>
        </Match>
      </Switch>
    </div>
  );
}

export default function ListVideoPanel(props: ListVideoPanelProps): JSXElement {
  // 左侧列表状态：收藏夹/合集列表、列表分页和列表加载错误。
  // listLoading 只表示首次加载或刷新整页列表。
  const [listLoading, setListLoading] = createSignal(false);
  // listError 只展示左侧列表请求失败；右侧详情错误用 detailError。
  const [listError, setListError] = createSignal("");
  // items 是左侧集合列表，不包含右侧视频详情。
  const [items, setItems] = createSignal<readonly DouyinListItem[]>(EMPTY_LIST);
  // listCursor 保存下一次“加载更多左侧列表”的 cursor。
  const [listCursor, setListCursor] = createSignal(0);
  // listHasMore 控制左侧底部“加载更多”按钮是否出现。
  const [listHasMore, setListHasMore] = createSignal(false);
  // listLoadingMore 只锁住左侧加载更多按钮，不影响右侧详情区域。
  const [listLoadingMore, setListLoadingMore] = createSignal(false);
  // loadedOnce 防止 tab 反复切换时重复拉列表。
  const [loadedOnce, setLoadedOnce] = createSignal(false);

  // 右侧详情状态：当前选中的左侧条目，以及该条目下的视频分页。
  // selectedItem 为空时右侧显示“请选择收藏夹/合集”的占位文案。
  const [selectedItem, setSelectedItem] = createSignal<DouyinListItem | null>(null);
  // detailLoading 表示当前选中条目的第一页视频正在加载。
  const [detailLoading, setDetailLoading] = createSignal(false);
  // detailError 只展示右侧视频详情请求失败。
  const [detailError, setDetailError] = createSignal("");
  // detailVideos 是右侧网格的数据源，会被传给 VideoContentPanel。
  const [detailVideos, setDetailVideos] = createSignal<readonly model.AwemeItem[]>(EMPTY_LIST);
  // detailHasMore 控制右侧视频网格底部“加载更多”按钮。
  const [detailHasMore, setDetailHasMore] = createSignal(false);
  // detailLoadingMore 只锁住右侧加载更多按钮，不清空当前视频。
  const [detailLoadingMore, setDetailLoadingMore] = createSignal(false);

  // 请求序号不是响应式状态，只用于丢弃乱序返回。
  let listRequestSeq = 0;
  let detailRequestSeq = 0;
  // 记录上一次实际加载的数据源，和 props.sourceKey 对比判断是否换源。
  let loadedSourceKey = "";
  // 记录 refreshKey 的旧值，值变化即视为外部要求刷新。
  let observedRefreshKey: unknown = props.refreshKey;

  // sourceKey/refreshKey 变化代表换了数据源，例如从收藏合集切到某个用户合集，需要清掉旧列表和旧详情。
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
    // append=true 是左侧列表“加载更多”；append=false 是首次加载/刷新列表。
    if (append) setListLoadingMore(true);
    else {
      setListLoading(true);
      setListError("");
      setListHasMore(false);
      setListLoadingMore(false);
    }

    // 请求序号用于忽略慢返回的旧请求，避免快速切换 tab 或用户时覆盖新状态。
    const seq = ++listRequestSeq;
    try {
      const page = await props.loadList(append ? listCursor() : 0);
      if (seq !== listRequestSeq) return;
      const nextItems = page.items;
      // 左侧分页直接拼接；刷新列表则整体替换。
      const merged = append ? [...items(), ...nextItems] : nextItems;
      setItems(merged);
      setListCursor(page.cursor);
      setListHasMore(page.hasMore);
      setLoadedOnce(true);

      if (!append) {
        // 列表刷新时只保留已有选择，不默认选择第一项，避免进入页面就请求详情接口。
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
    // append=true 是右侧视频“加载更多”；append=false 表示切换了左侧条目。
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
      // 详情分页追加视频；切换条目时重置为第一页。
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

  createEffect(() => {
    // inactive 的 tab 不拉数据。收藏页初始在收藏夹 tab，合集 tab 等切过去后再请求。
    if (!props.active) return;
    if (loadedSourceKey !== props.sourceKey) {
      loadedSourceKey = props.sourceKey;
      resetState();
    }
    if (observedRefreshKey !== props.refreshKey) {
      observedRefreshKey = props.refreshKey;
      resetState();
    }
    if (!loadedOnce()) void loadItems();
  });

  return (
    <SidebarDetailLayout
      // 下面这一组 props 只服务左侧列表区域。
      active={props.active}
      loading={listLoading()}
      error={listError()}
      emptyTitle={listEmptyTitle(props.kind)}
      onRetry={() => void loadItems()}
      items={items()}
      listTitle={listTitle(props.kind)}
      isSelected={(item) => {
        const selected = selectedItem();
        return !!selected && isSameItem(selected, item);
      }}
      onSelect={(item) => {
        const selected = selectedItem();
        if (selected && isSameItem(selected, item) && detailVideos().length > 0 && !detailLoading()) return;
        void loadDetail(item);
      }}
      hasMore={listHasMore()}
      loadingMore={listLoadingMore()}
      onLoadMore={() => void loadItems(true)}
      hasSelection={!!selectedItem()}
      selectionTitle={selectionTitle(props.kind)}
      selectionDescription={selectionDescription(props.kind)}
    >
      {/* 右侧详情区统一交给 VideoContentPanel，ListVideoPanel 只负责给它准备数据和分页回调。 */}
      <VideoContentPanel
        kind={props.kind as DouyinVideoContentKind}
        loading={detailLoading()}
        error={detailError()}
        onRetry={() => {
          const item = selectedItem();
          if (item) void loadDetail(item);
        }}
        title={selectedItem() ? itemTitle(selectedItem()!) : listTitle(props.kind)}
        items={detailVideos()}
        sourceName={selectedItem() ? itemTitle(selectedItem()!) : listTitle(props.kind)}
        fallbackAuthor={selectedItem() ? itemSubtitle(selectedItem()!) : "未知作者"}
        showToast={props.showToast}
        hasMore={detailHasMore()}
        loadingMore={detailLoadingMore()}
        onLoadMore={() => {
          const item = selectedItem();
          if (item) void loadDetail(item, true);
        }}
      />
    </SidebarDetailLayout>
  );
}
