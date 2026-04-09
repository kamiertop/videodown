import {createFileRoute} from '@tanstack/solid-router'
import {
    createMemo,
    createSignal,
    For,
    onMount,
    Show,
    type JSXElement,
} from "solid-js";
import {
    VideoDetailConcise,
    VideoDetailConciseBvid,
    VideoPlayURL,
} from "../../../wailsjs/go/api/BiliBili";
import type {model} from "../../../wailsjs/go/models";
import {ClipboardSetText} from "../../../wailsjs/runtime";
import ErrorToast from "../../components/ErrorToast";
import StatusToast from "../../components/StatusToast";

export const Route = createFileRoute('/bilibili/download')({
    validateSearch: (raw: Record<string, unknown>): { bvids?: string } => {
        const v = raw.bvids;
        if (typeof v === 'string' && v.trim()) return {bvids: v.trim()};
        return {};
    },
    component: DownLoad,
})

const BV_RE = /BV1\w{9}/i;

type StreamFormat = 'dash' | 'mp4' | 'flv';

interface DownloadItem {
    id: string;
    bvid: string;
    aid: number;
    title: string;
    pic: string;
    pages: { cid: number; label: string }[];
    cid: number | null;
    qn: number;
    streamFormat: StreamFormat;
    detailLoading: boolean;
    detailError: string;
    optionsLoading: boolean;
    acceptQuality: number[];
    acceptDescription: string[];
    urlLoading: boolean;
    playData: model.VideoURLData | null;
    playError: string;
}

const FORMAT_OPTIONS: { value: StreamFormat; label: string }[] = [
    {value: 'dash', label: 'DASH（音视频分离）'},
    {value: 'mp4', label: 'MP4（渐进）'},
    {value: 'flv', label: 'FLV'},
];

function streamFormatToFnval(f: StreamFormat): number {
    switch (f) {
        case 'dash':
            return 4048;
        case 'mp4':
            return 80;
        case 'flv':
            return 0;
        default:
            return 4048;
    }
}

function parseVideoInput(raw: string): { kind: 'aid'; aid: number } | { kind: 'bvid'; bvid: string } | null {
    const s = raw.trim();
    if (!s) return null;
    if (/^\d+$/.test(s)) {
        const aid = parseInt(s, 10);
        if (Number.isFinite(aid) && aid > 0) return {kind: 'aid', aid};
        return null;
    }
    const m = s.match(BV_RE);
    if (m) return {kind: 'bvid', bvid: m[0]};
    return null;
}

function pageOptions(view: model.VideoDetailView): { cid: number; label: string }[] {
    const pages = view.pages;
    if (pages?.length) {
        return pages.map(p => ({
            cid: p.cid,
            label: `P${p.page} ${p.part || '未命名'}`.trim(),
        }));
    }
    if (view.cid) {
        return [{cid: view.cid, label: 'P1 正片'}];
    }
    return [];
}

function pickStreamUrl(v: model.VideoItem | model.AudioItem): string {
    const u = v.baseUrl?.trim() || v.base_url?.trim();
    return u || '';
}

function formatBitrate(bps: number): string {
    if (!bps || bps <= 0) return '—';
    if (bps >= 1_000_000) return `${(bps / 1_000_000).toFixed(2)} Mbps`;
    if (bps >= 1000) return `${Math.round(bps / 1000)} kbps`;
    return `${bps} bps`;
}

function codecLabel(codecid: number): string {
    switch (codecid) {
        case 7:
            return 'AVC';
        case 12:
            return 'HEVC';
        case 13:
            return 'AV1';
        default:
            return codecid ? `codec ${codecid}` : '';
    }
}

function splitBvidsParam(s: string): string[] {
    return [...new Set(s.split(',').map(x => x.trim()).filter(Boolean))];
}

function qualityRows(item: DownloadItem): { q: number; label: string }[] {
    const qs = item.acceptQuality;
    const ds = item.acceptDescription;
    if (qs?.length) {
        return qs.map((q, i) => ({q, label: ds[i]?.trim() ? ds[i] : `qn ${q}`}));
    }
    return [{q: item.qn, label: `qn ${item.qn}（默认）`}];
}

function emptyItem(id: string, bvid: string): DownloadItem {
    return {
        id,
        bvid,
        aid: 0,
        title: '',
        pic: '',
        pages: [],
        cid: null,
        qn: 80,
        streamFormat: 'dash',
        detailLoading: true,
        detailError: '',
        optionsLoading: false,
        acceptQuality: [],
        acceptDescription: [],
        urlLoading: false,
        playData: null,
        playError: '',
    };
}

function StreamPanel(props: {
    play: model.VideoURLData;
    onCopy: (url: string) => void;
}): JSXElement {
    const sortedVideos = () => {
        const list = props.play.dash?.video ?? [];
        return [...list].sort((a, b) => (b.height || 0) - (a.height || 0));
    };
    const audioList = () => props.play.dash?.audio ?? [];
    return (
        <div class="mt-3 space-y-2 border-t border-base-300/60 pt-3">
            <div class="flex flex-wrap gap-2 text-xs">
                <span class="badge badge-outline badge-sm">容器 {props.play.format || '—'}</span>
                <span class="badge badge-outline badge-sm">qn {props.play.quality ?? '—'}</span>
            </div>
            <div class="overflow-hidden rounded-lg border border-base-300/80">
                <div
                    class="grid grid-cols-[minmax(0,1fr)_auto] gap-2 border-b border-base-300/80 bg-base-300/40 px-2 py-1.5 text-[11px] font-medium text-base-content/70">
                    <span>流</span>
                    <span class="w-16 shrink-0 text-right">操作</span>
                </div>
                <div class="max-h-64 divide-y divide-base-300/60 overflow-y-auto">
                    <For each={sortedVideos()}>
                        {item => {
                            const url = pickStreamUrl(item);
                            const res = item.width > 0 && item.height > 0
                                ? `${item.width}×${item.height}`
                                : '—';
                            const codec = codecLabel(item.codecid);
                            return (
                                <div
                                    class="grid grid-cols-1 gap-1 px-2 py-2 text-xs sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center">
                                    <div class="min-w-0 space-y-0.5">
                                        <div class="flex flex-wrap items-center gap-1.5">
                                            <span class="badge badge-primary badge-xs">视频</span>
                                            <span class="font-medium">{res}</span>
                                            {codec ? <span class="text-base-content/60">{codec}</span> : null}
                                        </div>
                                        <div class="break-all font-mono text-[10px] text-base-content/50">
                                            {item.mimeType || item.mime_type || '—'} · {formatBitrate(item.bandwidth)}
                                        </div>
                                        <div class="truncate font-mono text-[9px] text-base-content/40" title={url}>
                                            {url}
                                        </div>
                                    </div>
                                    <div class="flex justify-end">
                                        <button type="button" class="btn btn-ghost btn-xs"
                                                onClick={() => props.onCopy(url)}>
                                            复制
                                        </button>
                                    </div>
                                </div>
                            );
                        }}
                    </For>
                    <For each={audioList()}>
                        {item => {
                            const url = pickStreamUrl(item);
                            return (
                                <div
                                    class="grid grid-cols-1 gap-1 px-2 py-2 text-xs sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center">
                                    <div class="min-w-0 space-y-0.5">
                                        <div class="flex flex-wrap items-center gap-1.5">
                                            <span class="badge badge-secondary badge-xs">音频</span>
                                            <span class="text-base-content/70">id {item.id}</span>
                                        </div>
                                        <div class="break-all font-mono text-[10px] text-base-content/50">
                                            {item.mimeType || item.mime_type || '—'} · {formatBitrate(item.bandwidth)}
                                        </div>
                                        <div class="truncate font-mono text-[9px] text-base-content/40" title={url}>
                                            {url}
                                        </div>
                                    </div>
                                    <div class="flex justify-end">
                                        <button type="button" class="btn btn-ghost btn-xs"
                                                onClick={() => props.onCopy(url)}>
                                            复制
                                        </button>
                                    </div>
                                </div>
                            );
                        }}
                    </For>
                </div>
            </div>
            <Show when={!sortedVideos().length && !audioList().length}>
                <p class="text-xs text-base-content/50">当前格式下未返回可拆分的 DASH
                    音视频流，可尝试更换「格式」后重新获取。</p>
            </Show>
        </div>
    );
}

function DownLoad(): JSXElement {
    const search = Route.useSearch();
    const [input, setInput] = createSignal('');
    const [items, setItems] = createSignal<DownloadItem[]>([]);
    const [errorText, setErrorText] = createSignal('');
    const [statusText, setStatusText] = createSignal('');

    let errorToastTimer: number | undefined;
    let statusToastTimer: number | undefined;

    const showErrorToast = (message: string) => {
        setErrorText(message);
        if (errorToastTimer !== undefined) window.clearTimeout(errorToastTimer);
        errorToastTimer = window.setTimeout(() => {
            setErrorText('');
            errorToastTimer = undefined;
        }, 3500);
    };

    const showStatusToast = (message: string) => {
        setStatusText(message);
        if (statusToastTimer !== undefined) window.clearTimeout(statusToastTimer);
        statusToastTimer = window.setTimeout(() => {
            setStatusText('');
            statusToastTimer = undefined;
        }, 2200);
    };

    const mergeItem = (id: string, patch: Partial<DownloadItem>) => {
        setItems(prev => prev.map(it => (it.id === id ? {...it, ...patch} : it)));
    };

    const copyUrl = async (url: string) => {
        if (!url) {
            showErrorToast('无可用地址');
            return;
        }
        try {
            await ClipboardSetText(url);
            showStatusToast('已复制链接');
        } catch (e) {
            showErrorToast(e instanceof Error ? e.message : String(e));
        }
    };

    const prefetchOptions = async (id: string) => {
        const it = items().find(i => i.id === id);
        if (!it || it.cid == null) return;
        mergeItem(id, {optionsLoading: true, playError: '', playData: null});
        try {
            const fn = streamFormatToFnval(it.streamFormat);
            const data = await VideoPlayURL(it.aid, it.bvid, it.cid, 80, fn);
            const qs = data.accept_quality ?? [];
            const desc = data.accept_description ?? [];
            const qn = qs.length ? qs[qs.length - 1] : 80;
            mergeItem(id, {
                acceptQuality: qs,
                acceptDescription: desc,
                qn,
                optionsLoading: false,
            });
        } catch (e) {
            mergeItem(id, {
                optionsLoading: false,
                playError: e instanceof Error ? e.message : String(e),
            });
        }
    };

    const fetchPlayForItem = async (id: string) => {
        const it = items().find(i => i.id === id);
        if (!it || it.cid == null) {
            showErrorToast('请先等待详情加载并选择分 P');
            return;
        }
        mergeItem(id, {urlLoading: true, playError: ''});
        try {
            const fn = streamFormatToFnval(it.streamFormat);
            const data = await VideoPlayURL(it.aid, it.bvid, it.cid, it.qn, fn);
            mergeItem(id, {playData: data, urlLoading: false});
            showStatusToast('已获取播放地址');
        } catch (e) {
            mergeItem(id, {
                playData: null,
                urlLoading: false,
                playError: e instanceof Error ? e.message : String(e),
            });
        }
    };

    const loadDetailForBvid = async (bvid: string) => {
        try {
            const data = await VideoDetailConciseBvid(bvid);
            const view = data.view;
            const pages = pageOptions(view);
            const cid = pages[0]?.cid ?? null;
            mergeItem(bvid, {
                aid: view.aid,
                bvid: view.bvid || bvid,
                title: view.title ?? '',
                pic: view.pic ?? '',
                pages,
                cid,
                detailLoading: false,
                detailError: '',
            });
            if (cid != null) await prefetchOptions(bvid);
            else mergeItem(bvid, {detailLoading: false});
        } catch (e) {
            mergeItem(bvid, {
                detailLoading: false,
                detailError: e instanceof Error ? e.message : String(e),
            });
        }
    };

    const loadDetailForAid = async (aid: number, tempId: string) => {
        try {
            const data = await VideoDetailConcise(aid);
            const view = data.view;
            const pages = pageOptions(view);
            const cid = pages[0]?.cid ?? null;
            const key = view.bvid || tempId;
            setItems(prev => prev.map(it => {
                if (it.id !== tempId) return it;
                return {
                    ...it,
                    id: key,
                    aid: view.aid,
                    bvid: view.bvid,
                    title: view.title ?? '',
                    pic: view.pic ?? '',
                    pages,
                    cid,
                    detailLoading: false,
                    detailError: '',
                };
            }));
            if (cid != null) await prefetchOptions(key);
        } catch (e) {
            mergeItem(tempId, {
                detailLoading: false,
                detailError: e instanceof Error ? e.message : String(e),
            });
        }
    };

    const bootstrapFromBvids = (list: string[]) => {
        setItems(list.map(bv => emptyItem(bv, bv)));
        for (const bv of list) void loadDetailForBvid(bv);
    };

    const addFromInput = async () => {
        const parsed = parseVideoInput(input());
        if (!parsed) {
            showErrorToast('请输入有效的 BV 链接、BV 号或纯数字 av 号（aid）');
            return;
        }
        setInput('');
        if (parsed.kind === 'bvid') {
            const bv = parsed.bvid;
            if (items().some(i => i.bvid === bv)) {
                showStatusToast('列表中已有该稿件');
                return;
            }
            setItems(prev => [...prev, emptyItem(bv, bv)]);
            await loadDetailForBvid(bv);
            showStatusToast('已加入列表');
            return;
        }
        const aid = parsed.aid;
        const probeId = `aid-${aid}`;
        if (items().some(i => i.aid === aid)) {
            showStatusToast('列表中已有该稿件');
            return;
        }
        setItems(prev => [...prev, {...emptyItem(probeId, ''), aid, detailLoading: true}]);
        await loadDetailForAid(aid, probeId);
        showStatusToast('已加入列表');
    };

    onMount(() => {
        const q = search().bvids;
        if (!q) return;
        const list = splitBvidsParam(q);
        if (list.length === 0) return;
        bootstrapFromBvids(list);
    });

    const hasRows = createMemo(() => items().length > 0);

    return (
        <div class="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto overflow-x-hidden p-6">
            <ErrorToast message={errorText()}/>
            <StatusToast message={statusText()}/>

            <header class="flex flex-col gap-1">
                <h1 class="text-xl font-semibold tracking-tight text-base-content">视频下载</h1>
                <p class="text-sm text-base-content/60">
                    从 UP 主 / 收藏跳转时会自动列出稿件。每行可选择分 P、清晰度与封装格式，再获取播放地址。
                </p>
            </header>

            <div class="flex flex-col gap-2 sm:flex-row sm:items-center">
                <input
                    type="text"
                    class="input input-bordered input-sm flex-1 font-mono text-sm"
                    placeholder="粘贴 BV 链接 / BV 号 / aid 添加到列表"
                    value={input()}
                    onInput={e => setInput(e.currentTarget.value)}
                    onKeyDown={e => {
                        if (e.key === 'Enter') void addFromInput();
                    }}
                />
                <button type="button" class="btn btn-primary btn-sm shrink-0" onClick={() => void addFromInput()}>
                    添加到列表
                </button>
            </div>

            <Show when={!hasRows()}>
                <p class="rounded-lg border border-dashed border-base-300 bg-base-200/40 px-4 py-8 text-center text-sm text-base-content/50">
                    列表为空。从其他页面「下载」跳转，或在上框输入 BV / aid 添加。
                </p>
            </Show>

            <div class="flex flex-col gap-4">
                <For each={items()}>
                    {item => (
                        <div class="rounded-xl border border-base-300/80 bg-base-200/40 p-4">
                            <div class="flex flex-col gap-3 lg:flex-row lg:items-start">
                                <Show when={item.pic?.trim()}>
                                    <img
                                        src={item.pic}
                                        alt=""
                                        class="h-24 w-40 shrink-0 rounded-lg object-cover ring-1 ring-base-300"
                                    />
                                </Show>
                                <div class="min-w-0 flex-1 space-y-2">
                                    <div>
                                        <h2 class="text-sm font-semibold leading-snug text-base-content line-clamp-2">
                                            {item.detailLoading ? '加载中…' : (item.title || '（无标题）')}
                                        </h2>
                                        <div
                                            class="mt-1 flex flex-wrap gap-x-3 gap-y-0.5 text-[11px] text-base-content/60">
                                            <span>BV {item.bvid || '—'}</span>
                                            <span>AID {item.aid || '—'}</span>
                                        </div>
                                        <Show when={item.detailError}>
                                            <p class="mt-1 text-xs text-error">{item.detailError}</p>
                                        </Show>
                                        <Show when={item.playError}>
                                            <p class="mt-1 text-xs text-error">{item.playError}</p>
                                        </Show>
                                    </div>

                                    <div class="flex flex-col gap-2 xl:flex-row xl:flex-wrap xl:items-end">
                                        <label class="form-control min-w-0 flex-1 sm:max-w-55">
                                            <span class="label py-0"><span
                                                class="label-text text-[11px]">分 P</span></span>
                                            <select
                                                class="select select-bordered select-sm"
                                                disabled={item.detailLoading || !item.pages.length}
                                                value={item.cid ?? ''}
                                                onChange={e => {
                                                    const v = parseInt(e.currentTarget.value, 10);
                                                    const cid = Number.isFinite(v) ? v : null;
                                                    mergeItem(item.id, {
                                                        cid,
                                                        playData: null,
                                                        acceptQuality: [],
                                                        acceptDescription: [],
                                                    });
                                                    if (cid != null) void prefetchOptions(item.id);
                                                }}
                                            >
                                                <For each={item.pages}>
                                                    {p => (
                                                        <option value={p.cid}>{p.label}（cid {p.cid}）</option>
                                                    )}
                                                </For>
                                            </select>
                                        </label>

                                        <label class="form-control min-w-0 flex-1 sm:max-w-50">
                                            <span class="label py-0"><span class="label-text text-[11px]">清晰度</span></span>
                                            <select
                                                class="select select-bordered select-sm"
                                                disabled={item.detailLoading || item.optionsLoading || !qualityRows(item).length}
                                                value={item.qn}
                                                onChange={e => {
                                                    const v = parseInt(e.currentTarget.value, 10);
                                                    mergeItem(item.id, {qn: v, playData: null});
                                                }}
                                            >
                                                <For each={qualityRows(item)}>
                                                    {row => (
                                                        <option value={row.q}>{row.label}</option>
                                                    )}
                                                </For>
                                            </select>
                                        </label>

                                        <label class="form-control min-w-0 flex-1 sm:max-w-55">
                                            <span class="label py-0"><span
                                                class="label-text text-[11px]">格式</span></span>
                                            <select
                                                class="select select-bordered select-sm"
                                                disabled={item.detailLoading || item.cid == null}
                                                value={item.streamFormat}
                                                onChange={e => {
                                                    const v = e.currentTarget.value as StreamFormat;
                                                    mergeItem(item.id, {
                                                        streamFormat: v,
                                                        playData: null,
                                                        acceptQuality: [],
                                                        acceptDescription: [],
                                                    });
                                                    void prefetchOptions(item.id);
                                                }}
                                            >
                                                <For each={FORMAT_OPTIONS}>
                                                    {opt => (
                                                        <option value={opt.value}>{opt.label}</option>
                                                    )}
                                                </For>
                                            </select>
                                        </label>

                                        <button
                                            type="button"
                                            class="btn btn-secondary btn-sm shrink-0"
                                            disabled={
                                                item.detailLoading
                                                || item.cid == null
                                                || item.urlLoading
                                                || item.optionsLoading
                                            }
                                            onClick={() => void fetchPlayForItem(item.id)}
                                        >
                                            {(item.urlLoading || item.optionsLoading)
                                                ? <span class="loading loading-spinner loading-xs"/>
                                                : null}
                                            {item.optionsLoading
                                                ? '准备清晰度…'
                                                : item.urlLoading
                                                    ? '获取中…'
                                                    : '获取播放地址'}
                                        </button>
                                    </div>

                                    <Show when={item.playData}>
                                        {get => <StreamPanel play={get()} onCopy={u => void copyUrl(u)}/>}
                                    </Show>
                                </div>
                            </div>
                        </div>
                    )}
                </For>
            </div>
        </div>
    );
}
