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
import {extractBvid} from "../../lib/format";
import type {MediaCardItem} from "../../lib/model.ts";

type VideoDetailView = Awaited<ReturnType<typeof VideoDetailConciseBvid>>["view"];

interface ParsedVideoGroup {
    kind: "合集" | "系列";
    title: string;
    current: MediaCardItem;
    items: MediaCardItem[];
}

function normalizeBiliCover(url: string | undefined): string {
    if (!url) return "";
    if (url.startsWith("//")) return `https:${url}`;
    return url;
}

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
        const key = item.bvid?.trim() || String(item.id);
        if (!key || seen.has(key)) return false;
        seen.add(key);
        return true;
    });
}

function ugcSeasonToCards(view: VideoDetailView, current: MediaCardItem): MediaCardItem[] {
    const season = view.ugc_season;
    const listName = season?.title?.trim() || current.sourceListName;
    const episodes = season?.sections?.flatMap((section) => section.episodes ?? []) ?? [];

    return uniqueMediaCards(episodes.map((episode) => ({
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
    })));
}

function findGroupForParsedVideo(view: VideoDetailView, current: MediaCardItem): ParsedVideoGroup | null {
    const seasonCards = ugcSeasonToCards(view, current);
    if (seasonCards.length > 1) {
        return {
            kind: "合集",
            title: view.ugc_season?.title?.trim() || "未命名合集",
            current,
            items: seasonCards,
        };
    }

    return null;
}

export const Route = createFileRoute("/bilibili/download")({
    component: DownLoad,
});

function DownLoad(): JSXElement {
    const [videoURL, setVideoURL] = createSignal<string>("");
    const [parsing, setParsing] = createSignal(false);
    const [parsedGroup, setParsedGroup] = createSignal<ParsedVideoGroup | null>(null);
    const [selectedGroupIds, setSelectedGroupIds] = createSignal<number[]>([]);
    const {message, type, showToast} = useToast();
    const queue = useBilibiliDownloadQueue(showToast);
    const selectedGroupSet = createMemo(() => new Set(selectedGroupIds()));

    const allGroupSelected = createMemo(() => {
        const group = parsedGroup();
        if (!group || group.items.length === 0) return false;
        const selected = selectedGroupSet();
        return group.items.every((item) => selected.has(item.id));
    });

    function addParsedVideos(items: MediaCardItem[], successMessage: string): void {
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
            const detail = await VideoDetailConciseBvid(bvid);
            const card = detailToMediaCard(detail.view);
            if (!card.bvid) {
                showToast("解析成功，但详情中没有 BV 号", "error");
                return;
            }

            const group = findGroupForParsedVideo(detail.view, card);
            if (group) {
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
                        <DownloadVideoCard
                            canDownload={queue.canDownload(item) && (!queue.downloading() || queue.isDownloading(item))}
                            downloading={queue.isDownloading(item)}
                            item={item}
                            entry={queue.entryForBvid(item.bvid)}
                            progress={queue.progressFor(item)}
                            onPickQn={(qn) => queue.handlePickQn(item.bvid, qn)}
                            onPickAudio={(audioId) => queue.handlePickAudio(item.bvid, audioId)}
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
