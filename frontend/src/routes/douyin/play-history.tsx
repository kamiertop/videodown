import {createFileRoute, useNavigate} from '@tanstack/solid-router'
import {createEffect, createMemo, createSignal, For, type JSXElement, Match, Switch, untrack} from "solid-js";
import {createStore} from "solid-js/store";
import {History as FetchHistory} from "@bindings/github.com/kamiertop/videodown/douyin/api/douyin";
import * as model from "@bindings/github.com/kamiertop/videodown/douyin/model/models";
import DetailError from "../../components/DetailError.tsx";
import EmptyState from "../../components/EmptyState.tsx";
import Toast from "../../components/Toast.tsx";
import VideoGrid, {type DouyinVideoCardItem} from "../../components/douyin/VideoGrid.tsx";
import {useToast} from "../../hooks/useToast.ts";
import {
  douyinDownloadAssets,
  douyinImageURLs,
  douyinMediaBadge,
  douyinMusicURL,
} from "../../lib/douyin/media.ts";
import {addDouyinVideos, type DouyinDownloadItem, type DouyinVideoOption} from "../../lib/douyin/store.ts";
import {formatCount, formatDate, formatDuration} from "../../lib/format.ts";

type HistoryItem = model.HistoryItem;

type HistoryStatus = -1 | 0 | 1;

interface FilterOption<T extends number> {
  label: string;
  value: T;
}

const STATUS_OPTIONS: FilterOption<HistoryStatus>[] = [
  {label: "全部", value: -1},
  {label: "未看完", value: 0},
  {label: "已看完", value: 1},
];

const CATEGORY_OPTIONS: FilterOption<number>[] = [
  {label: "全部分类", value: 0},
  {label: "二次元", value: 1},
  {label: "音乐", value: 2},
  {label: "体育", value: 3},
  {label: "电影", value: 4},
  {label: "游戏", value: 5},
];

const DURATION_OPTIONS: FilterOption<number>[] = [
  {label: "全部时长", value: 0},
  {label: "1分钟内", value: 1},
  {label: "1-3分钟", value: 2},
  {label: "3-10分钟", value: 3},
  {label: "10分钟以上", value: 4},
];

export const Route = createFileRoute('/douyin/play-history')({
  component: DouyinPlayHistoryPage,
})

function normalizeDouyinDuration(value?: number): number {
  if (!value || value <= 0) return 0;
  return value >= 1000 ? Math.floor(value / 1000) : value;
}

function firstURL(playAddr: model.PlayInfo | undefined): string {
  return playAddr?.url_list?.[0] ?? "";
}

function historyKey(item: HistoryItem, index: number): string {
  return item.aweme_id || `history-${index}`;
}

function historyTitle(item: HistoryItem): string {
  return item.item_title || item.desc || item.caption || `作品 ${item.aweme_id || ""}`.trim();
}

function historyCover(item: HistoryItem): string {
  return item.video?.cover?.url_list?.[0]
      ?? item.video?.origin_cover?.url_list?.[0]
      ?? item.video?.dynamic_cover?.url_list?.[0]
      ?? "";
}

function coverCandidates(item: HistoryItem): model.Cover[] {
  const covers = [
    item.video?.dynamic_cover,
    item.video?.raw_cover,
    item.video?.cover,
    item.video?.origin_cover,
  ].filter((cover): cover is model.Cover => !!cover && (cover.url_list?.length ?? 0) > 0);

  const seen = new Set<string>();
  return covers.filter((cover) => {
    const key = cover.uri || cover.url_list?.[0];
    if (!key || seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}

function codecLabel(item: model.BitRateItem): string {
  if (item.is_h265 === 1) return "H.265";
  if (item.is_bytevc1 === 1) return "ByteVC1";
  return "H.264";
}

function videoOptionLabel(name: string, width: number | undefined, height: number | undefined, fps: number | undefined, dataSize: number, codec: string): string {
  const resolution = width && height ? `${width}×${height}` : "未知分辨率";
  const frameRate = fps ? ` · ${fps}fps` : "";
  return `${name} · ${resolution}${frameRate} · ${formatDataSize(dataSize)} · ${codec}`;
}

function optionID(prefix: string, url: string, index: number): string {
  return `${prefix}-${index}-${url.slice(0, 32)}`;
}

function formatDataSize(value: number | undefined): string {
  const size = Number(value) || 0;
  if (size <= 0) return "未知大小";
  const units = ["B", "KB", "MB", "GB"];
  let n = size;
  let index = 0;
  while (n >= 1024 && index < units.length - 1) {
    n /= 1024;
    index += 1;
  }
  return `${n >= 10 || index === 0 ? n.toFixed(0) : n.toFixed(1)} ${units[index]}`;
}

function videoOptions(item: HistoryItem): DouyinVideoOption[] {
  const seen = new Set<string>();
  const options: DouyinVideoOption[] = [];

  for (const [index, entry] of (item.video?.bit_rate ?? []).entries()) {
    const url = firstURL(entry.play_addr);
    if (!url || seen.has(url)) continue;
    seen.add(url);
    const gearName = entry.gear_name || `bit_rate_${index + 1}`;
    const codec = codecLabel(entry);
    const dataSize = Number(entry.play_addr?.data_size ?? 0);
    options.push({
      id: optionID("bitrate", url, index),
      label: videoOptionLabel(gearName, entry.play_addr?.width, entry.play_addr?.height, entry.FPS, dataSize, codec),
      gearName,
      dataSize,
      width: entry.play_addr?.width,
      height: entry.play_addr?.height,
      fps: entry.FPS,
      bitRate: entry.bit_rate,
      codec,
      url,
    });
  }

  const fallbacks: Array<{ name: string; codec: string; playAddr?: model.PlayInfo }> = [
    {name: "play_addr_h264", codec: "H.264", playAddr: item.video?.play_addr_h264},
    {name: "play_addr_265", codec: "H.265", playAddr: item.video?.play_addr_265},
    {name: "play_addr", codec: "默认", playAddr: item.video?.play_addr},
  ];

  for (const [index, fallback] of fallbacks.entries()) {
    const url = firstURL(fallback.playAddr);
    if (!url || seen.has(url)) continue;
    seen.add(url);
    const dataSize = Number(fallback.playAddr?.data_size ?? 0);
    options.push({
      id: optionID(fallback.name, url, index),
      label: videoOptionLabel(fallback.name, fallback.playAddr?.width, fallback.playAddr?.height, undefined, dataSize, fallback.codec),
      gearName: fallback.name,
      dataSize,
      width: fallback.playAddr?.width,
      height: fallback.playAddr?.height,
      codec: fallback.codec,
      url,
    });
  }

  return options;
}

function defaultVideoOption(options: DouyinVideoOption[]): DouyinVideoOption | undefined {
  // 默认在所有编码格式中选择最高画质。
  return [...options]
      .sort((a, b) => ((b.width ?? 0) * (b.height ?? 0) - (a.width ?? 0) * (a.height ?? 0)) || (b.fps ?? 0) - (a.fps ?? 0) || (b.bitRate ?? 0) - (a.bitRate ?? 0) || b.dataSize - a.dataSize)[0];
}

function toDownloadItem(item: HistoryItem): DouyinDownloadItem {
  const awemeId = item.aweme_id;
  const title = historyTitle(item);
  const cover = historyCover(item);
  const duration = normalizeDouyinDuration(item.video?.duration ?? 0);
  const options = videoOptions(item);
  const selected = defaultVideoOption(options);
  const mediaBadge = douyinMediaBadge(item);

  return {
    awemeId,
    sourceKind: "播放历史",
    sourceName: "抖音播放历史",
    title,
    cover,
    coverCandidates: coverCandidates(item),
    duration,
    authorName: item.author?.nickname || item.author?.uid || "未知作者",
    publishTime: item.create_time ?? 0,
    diggCount: item.statistics?.digg_count ?? 0,
    collectCount: item.statistics?.collect_count ?? 0,
    link: awemeId ? `https://www.douyin.com/video/${awemeId}` : undefined,
    videoURL: selected?.url,
    videoOptions: options,
    selectedVideoOptionId: selected?.id,
    imageURLs: douyinImageURLs(item),
    assets: mediaBadge ? douyinDownloadAssets(item) : undefined,
    musicURL: mediaBadge ? douyinMusicURL(item) : undefined,
    mediaBadge,
  };
}

function toCardItem(item: HistoryItem, index: number): DouyinVideoCardItem {
  const duration = normalizeDouyinDuration(item.video?.duration ?? 0);
  const mediaBadge = douyinMediaBadge(item);

  return {
    id: historyKey(item, index),
    cover: historyCover(item),
    title: historyTitle(item),
    author: item.author?.nickname || item.author?.uid || "未知作者",
    publishText: formatDate(item.create_time ?? 0),
    durationText: formatDuration(duration),
    downloadItem: toDownloadItem(item),
    mediaBadge,
    isTop: item.is_top === 1,
  };
}

function DouyinPlayHistoryPage(): JSXElement {
  const navigate = useNavigate();
  const {message, type, showToast} = useToast();
  const [items, setItems] = createSignal<HistoryItem[]>([]);
  const [cursor, setCursor] = createSignal(0);
  const [hasMore, setHasMore] = createSignal(false);
  const [loading, setLoading] = createSignal(false);
  const [loadingMore, setLoadingMore] = createSignal(false);
  const [error, setError] = createSignal("");
  const [status, setStatus] = createSignal<HistoryStatus>(-1);
  const [category, setCategory] = createSignal(0);
  const [duration, setDuration] = createSignal(0);
  const [allSelected, setAllSelected] = createSignal(false);
  const [selectedMap, setSelectedMap] = createStore<Record<string, true>>({});
  let requestSeq = 0;

  async function loadFirst(): Promise<void> {
    const seq = ++requestSeq;
    setLoading(true);
    setError("");
    untrack(clearSelection);
    try {
      const data = await FetchHistory(0, status(), category(), duration());
      if (seq !== requestSeq) return;
      setItems(data.aweme_list ?? []);
      setCursor(Number(data.max_cursor ?? 0));
      setHasMore(Number(data.has_more ?? 0) > 0);
    } catch (err) {
      if (seq !== requestSeq) return;
      setError(err instanceof Error ? err.message : String(err));
      setItems([]);
      setHasMore(false);
    } finally {
      if (seq === requestSeq) setLoading(false);
    }
  }

  async function loadMore(): Promise<void> {
    if (!hasMore() || loadingMore()) return;
    const seq = ++requestSeq;
    setLoadingMore(true);
    try {
      const data = await FetchHistory(cursor(), status(), category(), duration());
      if (seq !== requestSeq) return;
      const nextItems = data.aweme_list ?? [];
      setItems((current) => [...current, ...nextItems]);
      setCursor(Number(data.max_cursor ?? cursor()));
      setHasMore(Number(data.has_more ?? 0) > 0);
    } catch (err) {
      if (seq !== requestSeq) return;
      showToast(err instanceof Error ? err.message : String(err), "warning");
    } finally {
      if (seq === requestSeq) setLoadingMore(false);
    }
  }

  createEffect(() => {
    status();
    category();
    duration();
    void loadFirst();
  });

  function clearSelection(): void {
    setAllSelected(false);
    for (const id of Object.keys(selectedMap)) setSelectedMap(id, undefined!);
  }

  const cardItems = createMemo(() => items().map(toCardItem));
  const selectedKeys = createMemo(() => Object.keys(selectedMap));

  createEffect(() => {
    const validIds = new Set(cardItems().map((item) => item.id));
    for (const id of Object.keys(selectedMap)) {
      if (!validIds.has(id)) setSelectedMap(id, undefined!);
    }
  });

  const selectedCount = createMemo(() => {
    if (allSelected()) return Math.max(0, cardItems().length - selectedKeys().length);
    return selectedKeys().length;
  });

  function allVideosSelected(): boolean {
    return cardItems().length > 0 && selectedCount() === cardItems().length;
  }

  function isVideoSelected(id: string): boolean {
    if (allSelected()) return !selectedMap[id];
    return selectedMap[id];
  }

  function toggleSelect(id: string): void {
    if (selectedMap[id]) setSelectedMap(id, undefined!);
    else setSelectedMap(id, true);
  }

  function toggleSelectAll(): void {
    if (allVideosSelected()) {
      clearSelection();
      return;
    }
    setAllSelected(true);
    for (const id of Object.keys(selectedMap)) setSelectedMap(id, undefined!);
  }

  function videoCardClass(id: string): string {
    if (allSelected()) {
      return selectedMap[id]
          ? "border-2 border-base-300 bg-base-100 opacity-70"
          : "border-2 border-transparent bg-base-100 ring-1 ring-base-300";
    }
    return selectedMap[id]
        ? "border-2 border-success bg-success/5 shadow-sm shadow-success/15"
        : "border-2 border-transparent bg-base-100 ring-1 ring-base-300";
  }

  async function enqueueAndGoDownload(targets: DouyinVideoCardItem[]): Promise<void> {
    const downloadItems = targets.map((item) => item.downloadItem).filter((item) => item.videoURL);
    if (downloadItems.length === 0) {
      showToast("当前选择没有可用下载地址", "warning");
      return;
    }
    addDouyinVideos(downloadItems);
    clearSelection();
    await navigate({to: "/douyin/download"});
  }

  const selectedItems = createMemo(() => cardItems().filter((item) => isVideoSelected(item.id)));
  const hasActiveFilters = createMemo(() => status() !== -1 || category() !== 0 || duration() !== 0);

  function resetFilters(): void {
    if (!hasActiveFilters()) return;
    setStatus(-1);
    setCategory(0);
    setDuration(0);
  }

  function HistoryFilters(): JSXElement {
    return (
        <>
          <select class="select select-sm w-28" value={status()}
                  onChange={(event) => setStatus(Number(event.currentTarget.value) as HistoryStatus)}>
            <For each={STATUS_OPTIONS}>
              {(option) => <option value={option.value}>{option.label}</option>}
            </For>
          </select>
          <select class="select select-sm w-32" value={category()}
                  onChange={(event) => setCategory(Number(event.currentTarget.value))}>
            <For each={CATEGORY_OPTIONS}>
              {(option) => <option value={option.value}>{option.label}</option>}
            </For>
          </select>
          <select class="select select-sm w-36" value={duration()}
                  onChange={(event) => setDuration(Number(event.currentTarget.value))}>
            <For each={DURATION_OPTIONS}>
              {(option) => <option value={option.value}>{option.label}</option>}
            </For>
          </select>
          <button
              class="btn btn-ghost btn-sm"
              type="button"
              onClick={resetFilters}
              disabled={!hasActiveFilters() || loading()}
          >
            重置
          </button>
        </>
    );
  }

  return (
      <div class="flex h-full min-h-0 flex-col p-4">
        <section class="min-h-0 flex-1 overflow-hidden rounded-lg border border-base-300 bg-base-100">
          <Switch>
            <Match when={loading()}>
              <div class="flex h-full items-center justify-center">
                <span class="loading loading-spinner loading-md text-primary"/>
              </div>
            </Match>
            <Match when={error()}>
              <DetailError message={error()} onRetry={() => void loadFirst()}/>
            </Match>
            <Match when={items().length === 0}>
              <div class="flex h-full min-h-0 flex-1 items-center justify-center p-6">
                <EmptyState title="暂无播放历史" description="请先确认抖音账号已登录，或稍后重试。"/>
              </div>
            </Match>
            <Match when={cardItems().length > 0}>
              <VideoGrid
                  title={`播放历史 · ${formatCount(cardItems().length)} 个`}
                  items={cardItems()}
                  selectedCount={selectedCount()}
                  allSelected={allVideosSelected()}
                  selectedClass={videoCardClass}
                  onToggleItem={toggleSelect}
                  onToggleAll={toggleSelectAll}
                  onClearSelection={clearSelection}
                  onDownloadSelected={() => void enqueueAndGoDownload(selectedItems())}
                  onDownloadAll={() => void enqueueAndGoDownload(cardItems())}
                  toolbarMiddle={<HistoryFilters/>}
                  refreshing={loading()}
                  onRefresh={() => void loadFirst()}
                  hasMore={hasMore()}
                  loadingMore={loadingMore()}
                  onLoadMore={() => void loadMore()}
              />
            </Match>
          </Switch>
        </section>

        <Toast message={message()} type={type()}/>
      </div>
  );
}
