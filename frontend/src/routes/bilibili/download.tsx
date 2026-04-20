import {createFileRoute} from "@tanstack/solid-router";
import {createMemo, createSignal, For, Show, type JSXElement} from "solid-js";
import {VideoDetailConciseBvid} from "../../../wailsjs/go/api/BiliBili";
import DownloadInputBar from "../../components/bilibili/downloadPage/DownloadInputBar.tsx";
import DownloadSummaryBar from "../../components/bilibili/downloadPage/DownloadSummaryBar.tsx";
import DownloadVideoCard from "../../components/bilibili/downloadPage/DownloadVideoCard.tsx";
import Toast from "../../components/Toast";
import {useToast} from "../../hooks/useToast";
import {useBilibiliDownloadQueue} from "../../lib/bilibiliDownloadQueue.ts";
import {addVideos, removeVideo, videoList} from "../../lib/bilibiliStore.ts";
import {extractBilibiliPartIndex, extractBvid} from "../../lib/format";
import type {MediaCardItem} from "../../lib/model.ts";

type VideoDetailView = Awaited<ReturnType<typeof VideoDetailConciseBvid>>["view"];

// 手动输入一个 BV 后，如果详情里带 ugc_season，就先暂存在这里。
// 页面会展示这个分组，让用户决定“添加全部 / 添加部分 / 只添加当前视频”。
interface ParsedVideoGroup {
    kind: "合集" | "系列" | "分P";
    title: string;
    current: MediaCardItem;
    items: MediaCardItem[];
}

function normalizeBiliCover(url: string | undefined): string {
    if (!url) return "";
    // B 站图片经常返回 //i0.hdslb.com/...，img 标签里统一补成完整 https 地址。
    if (url.startsWith("//")) return `https:${url}`;
    return url;
}

// 后端视频详情结构很大，下载列表只需要统一的 MediaCardItem。
// 手动解析进来的条目 sourceListKind 固定为“解析结果”，后端据此不会再按来源创建子目录。
function detailToMediaCard(view: VideoDetailView): MediaCardItem {
    return {
        id: Number(view.aid) || Date.now(),
        title: view.title ?? "",
        cover: normalizeBiliCover(view.pic),
        duration: Number(view.duration) || 0,
        bvid: view.bvid ?? "",
        link: view.bvid ? `https://www.bilibili.com/video/${view.bvid}` : undefined,
        upperName: view.owner?.name || "未知",
        play: view.stat?.view,
        danmaku: view.stat?.danmaku,
        pubtime: view.pubdate,
        sourceListName: "手动解析",
        sourceListKind: "解析结果",
    };
}

function uniqueMediaCards(items: MediaCardItem[]): MediaCardItem[] {
    const seen = new Set<string>();
    return items.filter((item) => {
        const cid = item.cid;
        const key =
            cid != null && cid > 0
                ? `cid:${cid}`
                : item.bvid?.trim().toUpperCase() || String(item.id);
        if (!key || seen.has(key)) return false;
        seen.add(key);
        return true;
    });
}

function ugcSeasonToCards(view: VideoDetailView, current: MediaCardItem): MediaCardItem[] {
    const season = view.ugc_season;
    const listName = season?.title?.trim() || current.sourceListName;
    const episodes = season?.sections?.flatMap((section) => section.episodes ?? []) ?? [];

    return uniqueMediaCards(
        episodes.map((episode) => ({
            id: Number(episode.aid) || Number(episode.id) || Date.now(),
            title: episode.arc?.title || episode.title || "",
            cover: normalizeBiliCover(episode.arc?.pic || season?.cover || current.cover),
            duration: Number(episode.arc?.duration || episode.page?.duration || 0),
            bvid: episode.bvid ?? "",
            link: episode.bvid ? `https://www.bilibili.com/video/${episode.bvid}` : undefined,
            upperName: episode.arc?.author?.name || current.upperName,
            play: episode.arc?.stat?.view,
            danmaku: episode.arc?.stat?.danmaku,
            pubtime: episode.arc?.pubdate,
            sourceListName: listName,
            sourceListKind: "合集",
        })),
    );
}

function partListTitle(mainTitle: string, partName: string, pageIndex: number): string {
    const part = partName?.trim();
    if (part) return `${mainTitle} · P${pageIndex} ${part}`;
    return `${mainTitle} · P${pageIndex}`;
}

function pagesToMediaCards(view: VideoDetailView, upperName: string): MediaCardItem[] {
    const pages = view.pages ?? [];
    const bvid = view.bvid ?? "";
    const mainTitle = view.title?.trim() || "未命名稿件";
    return uniqueMediaCards(
        pages.map((p) => ({
            id: Number(p.cid) || p.page || Date.now(),
            title: partListTitle(mainTitle, p.part, p.page),
            cover: normalizeBiliCover(p.first_frame || view.pic),
            duration: Number(p.duration) || 0,
            bvid,
            cid: Number(p.cid) || undefined,
            link: bvid ? `https://www.bilibili.com/video/${bvid}?p=${p.page}` : undefined,
            upperName,
            play: view.stat?.view,
            danmaku: view.stat?.danmaku,
            pubtime: view.pubdate,
            sourceListName: mainTitle,
            sourceListKind: "分P",
        })),
    );
}

function findGroupForParsedVideo(
    view: VideoDetailView,
    current: MediaCardItem,
    rawInput: string,
): ParsedVideoGroup | null {
    // 只使用视频详情里已经返回的 ugc_season。
    // 如果没有 ugc_season，就当作普通单视频，不额外请求 UP 主页面或系列接口。
    const seasonCards = ugcSeasonToCards(view, current);
    if (seasonCards.length > 1) {
        return {
            kind: "合集",
            title: view.ugc_season?.title?.trim() || "未命名合集",
            current,
            items: seasonCards,
        };
    }

    const pageCards = pagesToMediaCards(view, current.upperName);
    if (pageCards.length > 1) {
        const wantP = extractBilibiliPartIndex(rawInput);
        const pages = view.pages ?? [];
        const hit =
            pages.find((p) => p.page === wantP) ??
            pages[0];
        const currentPart =
            pageCards.find((c) => c.cid === Number(hit?.cid)) ?? pageCards[0] ?? current;
        return {
            kind: "分P",
            title: view.title?.trim() || "多P 视频",
            current: currentPart,
            items: pageCards,
        };
    }

    return null;
}

export const Route = createFileRoute("/bilibili/download")({
    component: DownLoad,
});

function DownLoad(): JSXElement {
    // 输入框状态：只服务手动解析，不影响来自 UP/收藏/合集页面加入的 videoList。
    const [videoURL, setVideoURL] = createSignal<string>("");
    const [parsing, setParsing] = createSignal(false);
    // parsedGroup 表示“已解析到合集但还没真正加入下载列表”。
    const [parsedGroup, setParsedGroup] = createSignal<ParsedVideoGroup | null>(null);
    const [selectedGroupIds, setSelectedGroupIds] = createSignal<number[]>([]);
    const {message, type, showToast} = useToast();
    // 下载相关的副作用（监听 videoList、解析 DASH、接收进度事件）都集中在这个 hook 里。
    const queue = useBilibiliDownloadQueue(showToast);
    const selectedGroupSet = createMemo(() => new Set(selectedGroupIds()));

    const allGroupSelected = createMemo(() => {
        const group = parsedGroup();
        if (!group || group.items.length === 0) return false;
        const selected = selectedGroupSet();
        return group.items.every((item) => selected.has(item.id));
    });

    function addParsedVideos(items: MediaCardItem[], successMessage: string): void {
        // 真正加入全局下载列表后，清掉“合集选择”临时状态，避免下次解析混用旧数据。
        addVideos(items);
        setParsedGroup(null);
        setSelectedGroupIds([]);
        setVideoURL("");
        showToast(successMessage, "success");
    }

    function toggleGroupVideo(id: number): void {
        setSelectedGroupIds((prev) => {
            const next = new Set(prev);
            if (next.has(id)) next.delete(id);
            else next.add(id);
            return [...next];
        });
    }

    function toggleAllGroupVideos(): void {
        const group = parsedGroup();
        if (!group) return;
        setSelectedGroupIds(allGroupSelected() ? [] : group.items.map((item) => item.id));
    }

    async function parseVideo(): Promise<void> {
        if (parsing()) return;

        const bvid = extractBvid(videoURL());
        if (bvid === null) {
            showToast("请输入有效的 B 站视频链接或 BV 号", "error");
            return;
        }

        setParsing(true);
        try {
            // 手动解析只负责拿详情和判断是否为合集；播放地址由 useBilibiliDownloadQueue 监听 videoList 后解析。
            const detail = await VideoDetailConciseBvid(bvid);
            const card = detailToMediaCard(detail.view);
            if (!card.bvid) {
                showToast("解析成功，但详情中没有 BV 号", "error");
                return;
            }

            const group = findGroupForParsedVideo(detail.view, card, videoURL());
            if (group) {
                // 合集默认全选，但不立刻加入列表；让用户先删掉不想下载的条目。
                setParsedGroup(group);
                setSelectedGroupIds(group.items.map((item) => item.id));
                showToast(`发现${group.kind}「${group.title}」`, "info");
                return;
            }

            addParsedVideos([card], `已添加：${card.title || card.bvid}`);
        } catch (e) {
            showToast(e instanceof Error ? e.message : String(e), "error");
        } finally {
            setParsing(false);
        }
    }

    return (
        <div class="flex flex-col pt-4 pl-4 pr-4 pb-4 h-full">
            <DownloadInputBar
                value={videoURL()}
                onInput={setVideoURL}
                onParse={() => void parseVideo()}
                onClear={() => setVideoURL("")}
            />
            <Show when={parsedGroup()}>
                {(group) => (
                    // 这里展示的是“待确认的合集”，不是已进入下载队列的视频列表。
                    <section class="mt-3 rounded-lg border border-base-300 bg-base-100 p-3">
                        <div class="flex items-center gap-2">
                            <div class="min-w-0 flex-1">
                                <p class="truncate text-sm font-semibold">
                                    发现{group().kind}：{group().title}
                                </p>
                                <p class="text-xs text-base-content/60">
                                    可以添加全部，也可以只添加勾选的视频。
                                </p>
                            </div>
                            <button class="btn btn-ghost btn-xs" type="button" onClick={toggleAllGroupVideos}>
                                {allGroupSelected() ? "取消全选" : "全选"}
                            </button>
                            <button
                                class="btn btn-outline btn-primary btn-xs"
                                type="button"
                                disabled={selectedGroupIds().length === 0}
                                onClick={() => {
                                    const selected = group().items.filter((item) => selectedGroupSet().has(item.id));
                                    addParsedVideos(selected, `已添加 ${selected.length} 个视频`);
                                }}
                            >
                                添加已选 ({selectedGroupIds().length})
                            </button>
                            <button
                                class="btn btn-primary btn-xs"
                                type="button"
                                onClick={() => addParsedVideos(group().items, `已添加全部 ${group().items.length} 个视频`)}
                            >
                                添加全部
                            </button>
                            <button
                                class="btn btn-ghost btn-xs"
                                type="button"
                                onClick={() => addParsedVideos([group().current], `已添加：${group().current.title || group().current.bvid}`)}
                            >
                                只添加当前
                            </button>
                        </div>
                        <div class="mt-3 max-h-64 overflow-y-auto rounded border border-base-200">
                            <For each={group().items}>
                                {(item) => (
                                    <label class="flex cursor-pointer items-center gap-3 border-b border-base-200 px-3 py-2 last:border-b-0">
                                        <input
                                            class="checkbox checkbox-primary checkbox-sm"
                                            type="checkbox"
                                            checked={selectedGroupSet().has(item.id)}
                                            onChange={() => toggleGroupVideo(item.id)}
                                        />
                                        <img
                                            class="h-12 w-20 rounded object-cover"
                                            src={item.cover}
                                            alt={item.title}
                                            referrerPolicy="no-referrer"
                                        />
                                        <div class="min-w-0 flex-1">
                                            <p class="truncate text-sm font-medium">{item.title}</p>
                                            <p class="text-xs text-base-content/55">{item.bvid}</p>
                                        </div>
                                    </label>
                                )}
                            </For>
                        </div>
                    </section>
                )}
            </Show>
            <DownloadSummaryBar
                count={videoList().length}
                sourceSummary={queue.listSourceSummary()}
                downloading={queue.downloading()}
                onDownload={() => void queue.startDownload()}
            />
            <section class="mt-3 flex flex-1 flex-col gap-3 overflow-y-auto pr-4">
                <For each={videoList()}>
                    {(item) => (
                        // 每张卡片通过 bvid 到 queue 里取解析状态、画质音质和下载进度。
                        <DownloadVideoCard
                            canDownload={queue.canDownload(item) && (!queue.downloading() || queue.isDownloading(item))}
                            downloading={queue.isDownloading(item)}
                            item={item}
                            entry={queue.entryForItem(item)}
                            progress={queue.progressFor(item)}
                            onPickQn={(qn) => queue.handlePickQn(item, qn)}
                            onPickAudio={(audioId) => queue.handlePickAudio(item, audioId)}
                            onRemove={() => removeVideo(item.id)}
                            onDownload={() => void queue.downloadOne(item)}
                        />
                    )}
                </For>
            </section>
            <Toast message={message()} type={type()}/>
        </div>
    );
}
