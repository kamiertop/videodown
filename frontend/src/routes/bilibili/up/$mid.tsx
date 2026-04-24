import {createFileRoute, Link} from '@tanstack/solid-router'
import {createEffect, createMemo, createResource, createSignal, type JSXElement, Match, onMount, Show, Switch} from "solid-js";
import {Info, SeasonsArchivesList, SeasonsSeriesList, SeriesList, VideoList} from "../../../../wailsjs/go/api/BiliBili";
import {model} from "../../../../wailsjs/go/models";
import UpCommonLayout from "../../../components/bilibili/up/UpCommonLayout.tsx";
import DetailError from "../../../components/DetailError.tsx";
import DetailLoading from "../../../components/DetailLoading.tsx";
import EmptyState from "../../../components/EmptyState.tsx";
import IconBook from "../../../components/icons/IconBook.tsx";
import IconChevronLeft from "../../../components/icons/IconChevronLeft.tsx";
import IconRefresh from "../../../components/icons/IconRefresh.tsx";
import type {SidebarListItem} from "../../../components/SidebarList.tsx";
import SidebarList from "../../../components/SidebarList.tsx";
import Toast from "../../../components/Toast";
import VideoListSection from "../../../components/VideoListSection.tsx";
import {useToast} from "../../../hooks/useToast";
import {parseBilibiliLengthToSeconds} from "../../../lib/format";
import type {MediaCardItem} from "../../../lib/model.ts";

export const Route = createFileRoute('/bilibili/up/$mid')({
  component: UpDetail,
})

function UpDetail(): JSXElement {
  const params = Route.useParams();
  const {message, type, showToast} = useToast();
  const logic = createUpDetailLogic(() => params().mid, showToast);

  onMount(() => {
    logic.init();
  })

  return (
    <UpCommonLayout
      headerLeft={
        <>
          <Link
            to="/bilibili/up"
            class="btn btn-ghost btn-sm gap-1"
          >
            <IconChevronLeft class="h-4 w-4"/>
            返回
          </Link>
          <div class="h-5 w-px bg-base-300"></div>
          <h2 class="text-sm font-bold text-base-content">UP主详情</h2>
          <span class="rounded-full bg-base-200 px-2 py-0.5 text-xs tabular-nums text-base-content/60">
                        mid: {params().mid}
                    </span>
        </>
      }
      headerRight={
        <Switch>
          <Match when={logic.infoLoading()}>
            <div class="flex items-center gap-2">
              <span class="loading loading-spinner loading-xs text-primary"></span>
              <span class="text-xs text-base-content/50">获取UP主信息...</span>
            </div>
          </Match>
          <Match when={!logic.infoLoading() && logic.info()}>
            <div class="flex min-w-0 items-center gap-2">
              <div class="h-8 w-8 shrink-0 overflow-hidden rounded-full bg-base-200 ring-2 ring-base-200">
                <img
                  src={logic.info()!.face}
                  alt={logic.info()!.name}
                  referrerPolicy="no-referrer"
                  class="h-full w-full object-cover"
                />
              </div>
              <div class="min-w-0">
                <div class="flex min-w-0 items-center gap-2">
                                    <span class="max-w-[16rem] truncate text-sm font-black text-base-content">
                                        {logic.info()!.name}
                                    </span>
                  <span class="badge badge-outline badge-sm">Lv.{logic.info()!.level}</span>
                  <span
                    class={`badge badge-sm ${logic.info()!.is_followed ? 'badge-primary' : 'badge-ghost'}`}
                  >
                                        {logic.info()!.is_followed ? '已关注' : '未关注'}
                                    </span>
                </div>
              </div>
            </div>
          </Match>
        </Switch>
      }
    >
      <section class="flex h-full min-h-0 gap-3 overflow-hidden bg-base-200/40 p-3">
        <main
          class="flex min-w-0 flex-1 flex-col overflow-hidden rounded-xl border border-base-300 bg-base-100">
          <div class="flex shrink-0 border-b border-base-300">
            <button
              class={`flex-1 py-3 text-center text-sm font-bold transition-colors ${
                logic.activeTab() === 'videos'
                  ? 'border-b-2 border-success text-success'
                  : 'text-base-content/60 hover:text-base-content'
              }`}
              onClick={() => logic.setActiveTab('videos')}
            >
              全部视频
            </button>
            <button
              class={`flex-1 py-3 text-center text-sm font-bold transition-colors ${
                logic.activeTab() === 'lists'
                  ? 'border-b-2 border-success text-success'
                  : 'text-base-content/60 hover:text-base-content'
              }`}
              onClick={() => {
                logic.setActiveTab('lists');
                if (!logic.ssLoadedOnce()) void logic.loadSeasonsSeriesAll();
              }}
            >
              合集 | 系列
            </button>
          </div>

          <Switch>
            <Match when={logic.activeTab() === 'videos'}>
              <Switch>
                <Match when={logic.videoLoading()}>
                  <DetailLoading/>
                </Match>
                <Match when={!!logic.videoError()}>
                  <DetailError message={logic.videoError()}
                               onRetry={() => void logic.loadVideoList(false)}/>
                </Match>
                <Match when={!logic.videoLoading() && logic.videoCards().length === 0}>
                  <EmptyState title="暂无视频" description="该 UP 主暂无投稿视频或接口返回为空"/>
                </Match>
                <Match when={true}>
                  <VideoListSection
                    title="全部投稿视频"
                    mediaCount={logic.videoTotal() || logic.videoCards().length}
                    medias={() => logic.videoCards()}
                    selectionResetKey={() => `${params().mid}-v-${logic.videoListEpoch()}`}
                    showToast={showToast}
                    hasMore={logic.hasMoreVideos}
                    loadingMore={logic.videoLoadingMore}
                    onLoadMore={() => void logic.loadVideoList(true)}
                  />
                </Match>
              </Switch>
            </Match>

            <Match when={logic.activeTab() === 'lists'}>
              <div class="flex min-h-0 flex-1 overflow-hidden">
                <aside
                  class="flex w-64 shrink-0 flex-col overflow-hidden border-r border-base-300 bg-base-100">
                  <div
                    class="flex shrink-0 items-center justify-between border-b border-base-200 px-3 py-2">
                    <span class="text-xs font-bold text-base-content/70">列表</span>
                    <button
                      class="flex h-7 w-7 shrink-0 items-center justify-center rounded transition-colors hover:bg-base-200 disabled:cursor-not-allowed"
                      onClick={() => void logic.loadSeasonsSeriesAll()}
                      disabled={logic.ssLoading()}
                      title="刷新"
                    >
                      <IconRefresh
                        class={`h-3.5 w-3.5 text-base-content/50 ${logic.ssLoading() ? 'animate-spin' : ''}`}
                      />
                    </button>
                  </div>
                  <Switch>
                    <Match when={logic.ssLoading()}>
                      <div class="flex flex-1 items-center justify-center py-12">
                        <span class="loading loading-spinner loading-sm text-primary"></span>
                      </div>
                    </Match>
                    <Match when={!!logic.ssError()}>
                      <div class="p-3">
                        <DetailError message={logic.ssError()}
                                     onRetry={() => void logic.loadSeasonsSeriesAll()}/>
                      </div>
                    </Match>
                    <Match when={logic.listSidebarItems().length === 0}>
                      <EmptyState title="暂无合集/系列" compact/>
                    </Match>
                    <Match when={true}>
                      <SidebarList
                        list={() => logic.listSidebarItems() as any}
                        selectedId={() => logic.selectedListItem()?.id ?? null}
                        onSelect={logic.handleSelectListItem as any}
                        icon={<IconBook class="h-3 w-3"/>}
                      />
                    </Match>
                  </Switch>
                </aside>

                <div class="flex min-w-0 flex-1 flex-col overflow-hidden">
                  <Show
                    when={!!logic.selectedListItem()}
                    fallback={<EmptyState title="选择一个合集/系列查看内容"
                                          description="右侧将展示视频列表"/>}
                  >
                    <Switch>
                      <Match when={logic.listDetailLoading()}>
                        <DetailLoading/>
                      </Match>
                      <Match when={!!logic.listDetailError()}>
                        <DetailError message={logic.listDetailError()}
                                     onRetry={logic.retryListDetail}/>
                      </Match>
                      <Match when={!logic.listDetailLoading() && logic.listCards().length === 0}>
                        <EmptyState title="暂无视频"
                                    description="该合集/系列暂无可用视频或接口返回为空"/>
                      </Match>
                      <Match when={true}>
                        <VideoListSection
                          title={`${logic.selectedListItem()!.subtitle}: ${logic.selectedListItem()!.title}`}
                          mediaCount={logic.listTotal() || logic.selectedListItem()!.count || logic.listCards().length}
                          medias={() => logic.listCards()}
                          selectionResetKey={() => `${params().mid}-l-${logic.listDetailEpoch()}`}
                          showToast={showToast}
                          hasMore={logic.hasMoreListVideos}
                          loadingMore={logic.listLoadingMore}
                          onLoadMore={() => void logic.handleLoadMoreList()}
                        />
                      </Match>
                    </Switch>
                  </Show>
                </div>
              </div>
            </Match>
          </Switch>
        </main>
      </section>
      <Toast message={message()} type={type()}/>
    </UpCommonLayout>
  );
}

// ── 以下为该页专用：接口请求、分页、Tab；与 VideoListSection / VideoCardGrid 分工不同（组件不管拉数据） ──

type UpTab = 'videos' | 'lists';

type VideoListResp = {
  list?: {
    vlist?: Array<{
      aid: number;
      bvid: string;
      title: string;
      pic: string;
      length: string;
      created: number;
      play?: number;
      meta?: { stat?: { danmaku?: number } };
    }>
  };
  page: { pn: number; ps: number; count: number };
};

type SeriesArchivesResp = {
  page: { num: number; size: number; total: number };
  archives?: Array<{
    aid: number;
    bvid: string;
    title: string;
    pic: string;
    duration: number;
    pubdate: number;
    stat?: any;
  }>;
};

function normalizeBiliCover(url?: string): string {
  const u = (url ?? '').trim();
  if (!u) return '';
  if (u.startsWith('//')) return `https:${u}`;
  return u;
}

type ListKind = 'season' | 'series';

interface ListsSidebarItem extends SidebarListItem {
  kind: ListKind;
  raw: any;
  cover?: string;
}

function createUpDetailLogic(
  getMid: () => string,
  showToast: (message: string, type?: "error" | "success" | "info" | "warning") => void,
) {
  const [info] = createResource(
    getMid,
    async (mid): Promise<model.UserInfoData | null> => {
      try {
        return await Info(mid);
      } catch (error) {
        showToast(error instanceof Error ? error.message : String(error), 'error');
        return null;
      }
    },
  );

  const [activeTab, setActiveTab] = createSignal<UpTab>('videos');

  const VIDEO_PAGE_SIZE = 40;
  const [videoLoading, setVideoLoading] = createSignal(false);
  const [videoError, setVideoError] = createSignal('');
  const [videoCards, setVideoCards] = createSignal<MediaCardItem[]>([]);
  const [videoPage, setVideoPage] = createSignal(1);
  const [videoTotal, setVideoTotal] = createSignal(0);
  const [videoLoadingMore, setVideoLoadingMore] = createSignal(false);
  /** 非 append 拉取投稿列表时递增，用于 VideoListSection 清空勾选 */
  const [videoListEpoch, setVideoListEpoch] = createSignal(0);
  let videoReqSeq = 0;

  const hasMoreVideos = createMemo(() => {
    const total = videoTotal();
    const loaded = videoCards().length;
    return total > 0 && loaded < total;
  });

  function currentUpperName(): string {
    return info()?.name ?? '未知';
  }

  createEffect(() => {
    const data = info();
    if (!data?.name) {
      return;
    }

    const upperName = data.name;
    setVideoCards(prev => prev.map(c => (c.upperName === upperName ? c : {...c, upperName})));
    setListCards(prev => prev.map(c => (c.upperName === upperName ? c : {...c, upperName})));
  });

  const mapVlistToCards = (vlist: VideoListResp["list"] extends { vlist?: infer V } ? V : any): MediaCardItem[] => {
    const upperName = currentUpperName();
    return (vlist ?? []).map((v: any) => ({
      id: Number(v.aid) || 0,
      title: v.title ?? '',
      cover: normalizeBiliCover(v.pic),
      duration: parseBilibiliLengthToSeconds(v.length ?? ''),
      bvid: v.bvid ?? '',
      upperName,
      play: v.play,
      danmaku: v.meta?.stat?.danmaku,
      pubtime: v.created,
      // 与「来源」摘要、下载目录逻辑一致；实际目录在下载页对「全部投稿」仍优先用 upperName
      sourceListName: upperName,
      sourceListKind: '全部投稿',
    }));
  };

  const loadVideoList = async (append = false) => {
    const mid = getMid();
    const targetPage = append ? videoPage() + 1 : 1;
    if (append) setVideoLoadingMore(true);
    else {
      setVideoListEpoch(n => n + 1);
      setVideoLoading(true);
      setVideoError('');
      setVideoCards([]);
      setVideoPage(1);
      setVideoTotal(0);
    }

    const seq = ++videoReqSeq;
    try {
      const data = await VideoList(Number(mid), VIDEO_PAGE_SIZE, targetPage) as VideoListResp;
      if (seq !== videoReqSeq) return;
      const cards = mapVlistToCards(data.list?.vlist);
      const total = Number(data.page.count) || 0;
      setVideoTotal(total);
      if (append) setVideoCards(prev => [...prev, ...cards]);
      else setVideoCards(cards);
      setVideoPage(targetPage);
    } catch (error) {
      if (seq !== videoReqSeq) return;
      const msg = error instanceof Error ? error.message : String(error);
      if (append) showToast(`加载更多失败: ${msg}`, 'warning');
      else setVideoError(msg);
    } finally {
      if (seq === videoReqSeq) {
        if (append) setVideoLoadingMore(false);
        else setVideoLoading(false);
      }
    }
  };

  const SS_PAGE_SIZE = 20;
  const [ssLoading, setSsLoading] = createSignal(false);
  const [ssError, setSsError] = createSignal('');
  const [seasons, setSeasons] = createSignal<model.SeasonsItem[]>([]);
  const [series, setSeries] = createSignal<model.SeriesItem[]>([]);
  const [ssLoadedOnce, setSsLoadedOnce] = createSignal(false);

  const [selectedListItem, setSelectedListItem] = createSignal<ListsSidebarItem | null>(null);
  const [listDetailLoading, setListDetailLoading] = createSignal(false);
  const [listDetailError, setListDetailError] = createSignal('');
  const [listCards, setListCards] = createSignal<MediaCardItem[]>([]);
  const [listPage, setListPage] = createSignal(1);
  const [listTotal, setListTotal] = createSignal(0);
  const [listLoadingMore, setListLoadingMore] = createSignal(false);
  /** 非 append 拉取合集/系列详情时递增，用于 VideoListSection 清空勾选 */
  const [listDetailEpoch, setListDetailEpoch] = createSignal(0);
  let listReqSeq = 0;

  const listSidebarItems = createMemo((): ListsSidebarItem[] => {
    const sItems = seasons().map((s): ListsSidebarItem => ({
      id: s.meta?.season_id ?? 0,
      title: s.meta?.title ?? s.meta?.name ?? '未命名合集',
      count: s.meta?.total ?? 0,
      subtitle: '合集',
      kind: 'season',
      raw: s,
      cover: s.meta?.cover,
    }));
    const rItems = series().map((s): ListsSidebarItem => ({
      id: s.meta?.series_id ?? 0,
      title: s.meta?.name ?? '未命名系列',
      count: s.meta?.total ?? 0,
      subtitle: '系列',
      kind: 'series',
      raw: s,
      cover: s.meta?.cover,
    }));
    return [...sItems, ...rItems].filter(i => i.id > 0);
  });

  const hasMoreListVideos = createMemo(() => {
    const total = listTotal();
    const loaded = listCards().length;
    return total > 0 && loaded < total;
  });

  const mapArchivesToCards = (
    archives: any[] | undefined,
    upperName: string,
    listName: string,
    listKind: string,
  ): MediaCardItem[] => {
    const name = listName.trim();
    const kind = listKind.trim();
    return (archives ?? []).map((a: any) => ({
      id: Number(a.aid) || 0,
      title: a.title ?? '',
      cover: normalizeBiliCover(a.pic),
      duration: Number(a.duration) || 0,
      bvid: a.bvid ?? '',
      upperName,
      play: a.stat?.view,
      danmaku: a.stat?.danmaku,
      pubtime: a.pubdate ?? a.pubDate,
      sourceListName: name,
      sourceListKind: kind,
    }));
  };

  const loadListDetail = async (item: ListsSidebarItem, append: boolean) => {
    const mid = getMid();
    const targetPage = append ? listPage() + 1 : 1;
    setSelectedListItem(item);
    if (append) setListLoadingMore(true);
    else {
      setListDetailEpoch(n => n + 1);
      setListDetailLoading(true);
      setListDetailError('');
      setListCards([]);
      setListPage(1);
      setListTotal(item.count ?? 0);
    }

    const seq = ++listReqSeq;
    try {
      const upperName = currentUpperName();
      if (item.kind === 'season') {
        const data = await SeasonsArchivesList(mid, 20, targetPage, item.id) as unknown as model.SeasonsArchivesData;
        if (seq !== listReqSeq) return;
        const cards = mapArchivesToCards(data.archives, upperName, item.title, item.subtitle);
        const total = Number(data.meta?.total ?? data.page?.Total ?? item.count ?? 0) || 0;
        setListTotal(total);
        if (append) setListCards(prev => [...prev, ...cards]); else setListCards(cards);
        setListPage(targetPage);
      } else {
        const data = await SeriesList(mid, 20, targetPage, item.id) as SeriesArchivesResp;
        if (seq !== listReqSeq) return;
        const cards = mapArchivesToCards(data.archives as any[], upperName, item.title, item.subtitle);
        const total = Number((data as any).page?.total ?? (data as any).page?.Total ?? item.count ?? 0) || 0;
        setListTotal(total);
        if (append) setListCards(prev => [...prev, ...cards]); else setListCards(cards);
        setListPage(targetPage);
      }
    } catch (error) {
      if (seq !== listReqSeq) return;
      const msg = error instanceof Error ? error.message : String(error);
      if (append) showToast(`加载更多失败: ${msg}`, 'warning');
      else setListDetailError(msg);
    } finally {
      if (seq === listReqSeq) {
        if (append) setListLoadingMore(false);
        else setListDetailLoading(false);
      }
    }
  };

  const handleSelectListItem = (item: ListsSidebarItem) => {
    if (selectedListItem()?.kind === item.kind && selectedListItem()?.id === item.id && listCards().length > 0 && !listDetailLoading()) return;
    void loadListDetail(item, false);
  };

  const handleLoadMoreList = () => {
    const item = selectedListItem();
    if (!item || !hasMoreListVideos() || listLoadingMore()) return;
    void loadListDetail(item, true);
  };

  const retryListDetail = () => {
    const it = selectedListItem();
    if (!it) return;
    void loadListDetail(it, false);
  };

  const loadSeasonsSeriesAll = async () => {
    const mid = getMid();
    setSsLoading(true);
    setSsError('');
    setSeasons([]);
    setSeries([]);
    setSelectedListItem(null);
    setListCards([]);
    setListPage(1);
    setListTotal(0);

    try {
      const allSeasons: model.SeasonsItem[] = [];
      const allSeries: model.SeriesItem[] = [];

      let pageNum = 1;
      let total = 0;
      let fetched = 0;
      const maxPages = 50;

      while (pageNum <= maxPages) {
        const data = await SeasonsSeriesList(mid, SS_PAGE_SIZE, pageNum);
        const list = data.items_lists;
        const seasonsList = list?.seasons_list ?? [];
        const seriesList = list?.series_list ?? [];
        allSeasons.push(...seasonsList);
        allSeries.push(...seriesList);

        total = list?.page?.total ?? total;
        fetched = allSeasons.length + allSeries.length;
        if (total > 0 && fetched >= total) break;

        if (seasonsList.length === 0 && seriesList.length === 0) break;
        pageNum += 1;
      }

      setSeasons(allSeasons);
      setSeries(allSeries);
      setSsLoadedOnce(true);

      const first = [...allSeasons.map(s => ({
        id: s.meta?.season_id ?? 0,
        title: s.meta?.title ?? s.meta?.name ?? '未命名合集',
        count: s.meta?.total ?? 0,
        subtitle: '合集',
        kind: 'season' as const,
        raw: s,
        cover: s.meta?.cover,
      })), ...allSeries.map(s => ({
        id: s.meta?.series_id ?? 0,
        title: s.meta?.name ?? '未命名系列',
        count: s.meta?.total ?? 0,
        subtitle: '系列',
        kind: 'series' as const,
        raw: s,
        cover: s.meta?.cover,
      }))].find(i => i.id > 0) ?? null;

      if (first) void loadListDetail(first, false);
    } catch (error) {
      setSsError(error instanceof Error ? error.message : String(error));
    } finally {
      setSsLoading(false);
    }
  };

  const init = () => {
    void loadVideoList(false);
  };

  return {
    info,
    infoLoading: () => info.loading,
    activeTab,
    setActiveTab,
    videoLoading,
    videoError,
    videoCards,
    videoTotal,
    videoLoadingMore,
    hasMoreVideos,
    loadVideoList,
    ssLoadedOnce,
    ssLoading,
    ssError,
    listSidebarItems,
    selectedListItem,
    listDetailLoading,
    listDetailError,
    listCards,
    listTotal,
    listLoadingMore,
    hasMoreListVideos,
    loadSeasonsSeriesAll,
    handleSelectListItem,
    handleLoadMoreList,
    videoListEpoch,
    listDetailEpoch,
    init,
    retryListDetail,
  };
}
