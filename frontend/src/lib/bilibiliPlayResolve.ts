import {VideoDetailConciseBvid, VideoPlayURL} from "../../wailsjs/go/api/BiliBili";
import {model} from "../../wailsjs/go/models";

export function bvidCacheKey(bvid: string | undefined): string | null {
    const t = bvid?.trim();
    if (!t) return null;
    return t.toUpperCase();
}

function streamBaseUrl(v: { baseUrl?: string; base_url?: string }): string {
    return (v.baseUrl || v.base_url || "").trim();
}

/** 在同一批流里选“最佳视频”：优先码率，其次分辨率 */
function pickBestVideo(videos: model.VideoItem[] | undefined): model.VideoItem | undefined {
    const valid = (videos ?? []).filter((v) => streamBaseUrl(v));
    if (valid.length === 0) return undefined;
    return valid.reduce((a, b) => {
        if (b.bandwidth !== a.bandwidth) return b.bandwidth > a.bandwidth ? b : a;
        return b.width * b.height > a.width * a.height ? b : a;
    });
}

/** 先按 qn 过滤，再在该档位内挑最佳编码流 */
function pickBestVideoAtQn(videos: model.VideoItem[] | undefined, qn: number): model.VideoItem | undefined {
    const hits = (videos ?? []).filter((v) => v.id === qn);
    return pickBestVideo(hits);
}

/** 音频默认选最高码率 */
function pickBestAudio(audios: model.AudioItem[] | undefined): model.AudioItem | undefined {
    const valid = (audios ?? []).filter((a) => streamBaseUrl(a));
    if (valid.length === 0) return undefined;
    return valid.reduce((a, b) => (b.bandwidth > a.bandwidth ? b : a));
}

/** 按音轨 id 精确命中（用于“切音轨”） */
function pickAudioById(audios: model.AudioItem[] | undefined, id: number): model.AudioItem | undefined {
    return (audios ?? []).find((a) => a.id === id && streamBaseUrl(a));
}

/** 展示用：音轨按码率从高到低 */
export function sortedAudioTracks(audios: model.AudioItem[] | undefined): model.AudioItem[] {
    const valid = (audios ?? []).filter((a) => streamBaseUrl(a));
    return [...valid].sort((a, b) => b.bandwidth - a.bandwidth);
}

/** 下拉选项：码率 + codecs */
export function audioSelectLabel(a: model.AudioItem): string {
    const kbps = Math.round(a.bandwidth / 1000);
    const codec = a.codecs?.trim() || `codecid ${a.codecid}`;
    return `${kbps} kbps · ${codec}`;
}

export function audioDetailTitle(a: model.AudioItem): string {
    const lines: string[] = [];
    lines.push(`id: ${a.id}`);
    lines.push(`bandwidth: ${a.bandwidth} bps (~${Math.round(a.bandwidth / 1000)} kbps)`);
    if (a.codecs?.trim()) lines.push(`codecs: ${a.codecs.trim()}`);
    lines.push(`codecid: ${a.codecid}`);
    if (a.mimeType?.trim() || a.mime_type?.trim()) {
        lines.push(`mime: ${(a.mimeType || a.mime_type).trim()}`);
    }
    return lines.join("\n");
}

export function codecLabel(codecid: number): string {
    switch (codecid) {
        case 7:
            return "AVC";
        case 12:
            return "HEVC";
        case 13:
            return "AV1";
        default:
            return `codec ${codecid}`;
    }
}

function qualityDescription(play: model.VideoURLData, qn: number): string {
    const hit = play.support_formats?.find((s: model.SupportFormat) => s.quality === qn);
    return hit?.display_desc?.trim() || hit?.new_description?.trim() || `qn ${qn}`;
}

/** 展示用：按 qn 从高到低排列 support_formats */
export function sortedSupportFormats(formats: model.SupportFormat[] | undefined): model.SupportFormat[] {
    if (!formats?.length) return [];
    return [...formats].sort((a, b) => b.quality - a.quality);
}

/** 下拉选项一行文案：标题 · 封装 · 角标 */
export function supportFormatSelectLabel(fmt: model.SupportFormat): string {
    const main = fmt.display_desc?.trim() || fmt.new_description?.trim() || `qn ${fmt.quality}`;
    const parts: string[] = [main];
    if (fmt.format?.trim()) parts.push(fmt.format.trim());
    if (fmt.superscript?.trim()) parts.push(fmt.superscript.trim());
    return parts.join(" · ");
}

/** option 的 title，悬停查看完整字段 */
export function supportFormatDetailTitle(fmt: model.SupportFormat): string {
    const lines: string[] = [];
    if (fmt.display_desc?.trim()) lines.push(`display_desc: ${fmt.display_desc.trim()}`);
    if (fmt.new_description?.trim()) lines.push(`new_description: ${fmt.new_description.trim()}`);
    if (fmt.format?.trim()) lines.push(`format: ${fmt.format.trim()}`);
    if (fmt.codecs?.length) lines.push(`codecs: ${fmt.codecs.join(", ")}`);
    if (fmt.superscript?.trim()) lines.push(`superscript: ${fmt.superscript.trim()}`);
    lines.push(`quality (qn): ${fmt.quality}`);
    return lines.join("\n");
}

export interface ResolvedPlayInfo {
    aid: number;
    cid: number;
    bvid: string;
    partCount: number;
    /** 当前档位（与 play.quality 一致），由用户点击 support_formats 或默认最高档 */
    selectedQn: number;
    play: model.VideoURLData;
    bestVideo: model.VideoItem;
    bestAudio: model.AudioItem | undefined;
    apiQualityLabel: string;
    summaryLine: string;
}

function buildSummaryLine(
    play: model.VideoURLData,
    bestVideo: model.VideoItem,
    bestAudio: model.AudioItem | undefined,
    partCount: number,
    qn: number,
): string {
    const apiQualityLabel = qualityDescription(play, qn);
    const res = `${bestVideo.width}×${bestVideo.height}`;
    const partHint = partCount > 1 ? ` · 首P（共 ${partCount}P）` : "";
    const mbps = (bestVideo.bandwidth / 1e6).toFixed(2);
    let line = `当前流 · ${apiQualityLabel} · ${res} · ${codecLabel(bestVideo.codecid)} · ~${mbps} Mbps${partHint}`;
    if (bestAudio) {
        const kbps = Math.round(bestAudio.bandwidth / 1000);
        line += ` · 音频 ~${kbps} kbps`;
    }
    return line;
}

function buildResolvedPlayInfo(
    ctx: {aid: number; cid: number; bvid: string; partCount: number},
    play: model.VideoURLData,
    opts?: {preferredAudioId?: number},
): ResolvedPlayInfo {
    // 有些响应会返回多编码同档位，这里优先用 play.quality 对应档位内的“最佳视频流”。
    const bestVideo =
        pickBestVideoAtQn(play.dash?.video, play.quality) ?? pickBestVideo(play.dash?.video);
    if (!bestVideo) {
        throw new Error("未获取到 DASH 视频流（可能为特殊稿件或需登录）");
    }
    let bestAudio = pickBestAudio(play.dash?.audio);
    const prefId = opts?.preferredAudioId;
    if (prefId != null) {
        // 切清晰度后尽量保留用户之前选过的音轨。
        const pref = pickAudioById(play.dash?.audio, prefId);
        if (pref) bestAudio = pref;
    }
    const selectedQn = play.quality;
    const apiQualityLabel = qualityDescription(play, play.quality);
    const summaryLine = buildSummaryLine(play, bestVideo, bestAudio, ctx.partCount, selectedQn);

    return {
        aid: ctx.aid,
        cid: ctx.cid,
        bvid: ctx.bvid,
        partCount: ctx.partCount,
        selectedQn,
        play,
        bestVideo,
        bestAudio,
        apiQualityLabel,
        summaryLine,
    };
}

/**
 * 本地切换档位：仅复用当前 play.dash 中已存在的流地址，不发起新请求。
 * 若目标 qn 在当前 dash 中不可用，返回 null 让上层决定是否回退到重新请求。
 */
export function switchResolvedPlayAtQn(current: ResolvedPlayInfo, qn: number): ResolvedPlayInfo | null {
    if (qn === current.selectedQn) return current;

    // 只在本地 dash.video 中找目标 qn，不发请求。
    const bestVideo = pickBestVideoAtQn(current.play.dash?.video, qn);
    if (!bestVideo) return null;

    return {
        ...current,
        selectedQn: qn,
        bestVideo,
        apiQualityLabel: qualityDescription(current.play, qn),
        summaryLine: buildSummaryLine(current.play, bestVideo, current.bestAudio, current.partCount, qn),
    };
}

/** 本地切换音轨：仅复用当前 play.dash.audio */
export function switchResolvedAudio(current: ResolvedPlayInfo, audioId: number): ResolvedPlayInfo | null {
    // 只切换音轨，不改清晰度。
    const hit = pickAudioById(current.play.dash?.audio, audioId);
    if (!hit) return null;
    return {
        ...current,
        bestAudio: hit,
        summaryLine: buildSummaryLine(
            current.play,
            current.bestVideo,
            hit,
            current.partCount,
            current.selectedQn,
        ),
    };
}

function defaultQnFromProbe(probe: model.VideoURLData): number {
    if (probe.accept_quality?.length) {
        return Math.max(...probe.accept_quality);
    }
    return probe.quality;
}

/**
 * 详情 + 探测 + 默认最高可用档位 playurl；档位与编码仅通过 support_formats 展示与点击切换。
 */
export async function resolveBilibiliPlayUrl(bvidRaw: string): Promise<ResolvedPlayInfo> {
    const trimmed = bvidRaw.trim();
    if (!trimmed) {
        throw new Error("BV 号为空");
    }

    const detail = await VideoDetailConciseBvid(trimmed);
    const view = detail.view;
    const aid = Number(view.aid);
    const cid = Number(view.cid);
    if (!Number.isFinite(aid) || !Number.isFinite(cid)) {
        throw new Error("稿件 aid/cid 无效");
    }
    const bvid = view.bvid?.trim() || trimmed;
    const partCount = Math.max(1, Number(view.videos || view.pages?.length || 1));

    // 先探测可用档位，再拉默认最高可用档位，避免直接请求不可用 qn。
    const probe = await VideoPlayURL(aid, bvid, cid, 80);
    const defaultQn = defaultQnFromProbe(probe);

    const play = await VideoPlayURL(aid, bvid, cid, defaultQn);
    return buildResolvedPlayInfo({aid, cid, bvid, partCount}, play);
}

/** 用户点击某条 support_formats 后按对应 qn 重新拉取 playurl */
export async function refetchBilibiliPlayAtQn(input: {
    aid: number;
    cid: number;
    bvid: string;
    partCount: number;
    qn: number;
    /** 若新响应里仍有同 id 音轨则沿用 */
    preferredAudioId?: number;
}): Promise<ResolvedPlayInfo> {
    // 真正重拉 playurl 的入口：仅在本地切换失败或需要强制刷新时调用。
    const play = await VideoPlayURL(input.aid, input.bvid, input.cid, input.qn);
    return buildResolvedPlayInfo(
        {
            aid: input.aid,
            cid: input.cid,
            bvid: input.bvid,
            partCount: input.partCount,
        },
        play,
        input.preferredAudioId != null ? {preferredAudioId: input.preferredAudioId} : undefined,
    );
}
