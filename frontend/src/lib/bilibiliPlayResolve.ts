import {VideoDetailConciseBvid, VideoPlayURL} from "../../wailsjs/go/api/BiliBili";
import {model} from "../../wailsjs/go/models";

export function bvidCacheKey(bvid: string | undefined): string | null {
  const t = bvid?.trim();
  if (!t) return null;
  return t.toUpperCase();
}

/** 列表解析缓存 / 进度：多 P 稿件用 BV+cid，否则仅 BV。 */
export function bilibiliPlayResolveKey(item: { bvid?: string; cid?: number | null }): string | null {
  const bv = bvidCacheKey(item.bvid);
  if (!bv) return null;
  const c = item.cid;
  if (c != null && Number.isFinite(c) && c > 0) {
    return `${bv}:${Math.trunc(c)}`;
  }
  return bv;
}

export function streamBaseUrl(v: { baseUrl?: string; base_url?: string }): string {
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

export interface VideoQualityOption {
  qn: number;
  label: string;
  title: string;
}

export interface VideoAccessInfo {
  isChargeableSeason: boolean;
  isUpowerExclusive: boolean;
  isUpowerPlay: boolean;
  isUpowerPreview: boolean;
  isUpowerExclusiveWithQa: boolean;
  isPaid: boolean;
  labels: string[];
}

export class BilibiliPlayResolveError extends Error {
  accessInfo: VideoAccessInfo | undefined;

  constructor(message: string, accessInfo?: VideoAccessInfo) {
    super(message);
    this.name = "BilibiliPlayResolveError";
    this.accessInfo = accessInfo;
  }
}

function videoAccessInfoFromView(view: model.VideoDetailView): VideoAccessInfo {
  const isPaid =
    (view.rights?.pay ?? 0) > 0 ||
    (view.rights?.ugc_pay ?? 0) > 0 ||
    (view.rights?.arc_pay ?? 0) > 0;
  const labels: string[] = [];

  if (view.is_upower_exclusive) labels.push("充电专属");
  if (view.is_upower_play) labels.push("支持试看");
  if (view.is_upower_preview) labels.push("试看内容");
  if (view.is_upower_exclusive_with_qa) labels.push("充电专属问答");
  if (view.is_chargeable_season) labels.push("充电合集");
  if (isPaid) labels.push("付费内容");

  return {
    isChargeableSeason: view.is_chargeable_season,
    isUpowerExclusive: view.is_upower_exclusive,
    isUpowerPlay: view.is_upower_play,
    isUpowerPreview: view.is_upower_preview,
    isUpowerExclusiveWithQa: view.is_upower_exclusive_with_qa,
    isPaid,
    labels,
  };
}

function shouldSkipPlayUrl(accessInfo: VideoAccessInfo): boolean {
  // 充电专属且当前账号没有播放/试看权限时，playurl 通常不会返回可下载 DASH 流。
  // 在详情阶段直接标记失败，避免用户选中多个视频时多打无效的播放地址请求。
  return accessInfo.isUpowerExclusive && !accessInfo.isUpowerPlay && !accessInfo.isUpowerPreview;
}

function errorMessage(e: unknown): string {
  return e instanceof Error ? e.message : String(e);
}

function supportFormatForQn(play: model.VideoURLData, qn: number): model.SupportFormat | undefined {
  return play.support_formats?.find((s: model.SupportFormat) => s.quality === qn);
}

function dashVideoQualityTitle(play: model.VideoURLData, qn: number): string {
  const lines: string[] = [];
  const fmt = supportFormatForQn(play, qn);
  if (fmt) {
    lines.push(supportFormatDetailTitle(fmt));
  } else {
    lines.push(`quality (qn): ${qn}`);
  }

  const videos = (play.dash?.video ?? []).filter((v) => v.id === qn && streamBaseUrl(v));
  if (videos.length > 0) {
    const codecs = [...new Set(videos.map((v) => codecLabel(v.codecid)))];
    lines.push(`dash streams: ${videos.length}`);
    lines.push(`codecs: ${codecs.join(", ")}`);
  }

  return lines.join("\n");
}

/** 展示用：只列出当前 play.dash.video 中实际带流地址的画质，切换时无需重拉 playurl。 */
export function sortedDashVideoQualities(play: model.VideoURLData): VideoQualityOption[] {
  const qns = new Set<number>();
  for (const video of play.dash?.video ?? []) {
    if (Number.isFinite(video.id) && streamBaseUrl(video)) {
      qns.add(video.id);
    }
  }

  return [...qns]
    .sort((a, b) => b - a)
    .map((qn) => {
      const fmt = supportFormatForQn(play, qn);
      return {
        qn,
        label: fmt ? supportFormatSelectLabel(fmt) : qualityDescription(play, qn),
        title: dashVideoQualityTitle(play, qn),
      };
    });
}

export interface ResolvedPlayInfo {
  aid: number;
  cid: number;
  bvid: string;
  partCount: number;
  /** 当前档位：初始来自 play.quality，后续只在当前 play.dash.video 内切换 */
  selectedQn: number;
  play: model.VideoURLData;
  bestVideo: model.VideoItem;
  bestAudio: model.AudioItem | undefined;
  apiQualityLabel: string;
  summaryLine: string;
  accessInfo: VideoAccessInfo;
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
  ctx: { aid: number; cid: number; bvid: string; partCount: number },
  play: model.VideoURLData,
  accessInfo: VideoAccessInfo,
  opts?: { preferredAudioId?: number },
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
  // 如果接口降级了 play.quality，但 dash.video 已经带了实际可用流，用真实选中的流档位作为初始值。
  const selectedQn = Number.isFinite(bestVideo.id) ? bestVideo.id : play.quality;
  const apiQualityLabel = qualityDescription(play, selectedQn);
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
    accessInfo,
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

/**
 * 详情 + 一次 playurl；画质与音质切换只复用返回的 dash 字段，不重新请求。
 */
export async function resolveBilibiliPlayUrl(
  bvidRaw: string,
  opts?: { cid?: number | null },
): Promise<ResolvedPlayInfo> {
  const trimmed = bvidRaw.trim();
  if (!trimmed) {
    throw new Error("BV 号为空");
  }

  const detail = await VideoDetailConciseBvid(trimmed);
  const view = detail.view;
  const accessInfo = videoAccessInfoFromView(view);
  const aid = Number(view.aid);
  const cidOpt = opts?.cid;
  const cid =
    cidOpt != null && Number.isFinite(cidOpt) && cidOpt > 0 ? Math.trunc(cidOpt) : Number(view.cid);
  if (!Number.isFinite(aid) || !Number.isFinite(cid)) {
    throw new BilibiliPlayResolveError("稿件 aid/cid 无效", accessInfo);
  }
  const bvid = view.bvid?.trim() || trimmed;
  const partCount = Math.max(1, Number(view.videos || view.pages?.length || 1));

  if (shouldSkipPlayUrl(accessInfo)) {
    throw new BilibiliPlayResolveError("充电专属视频，当前账号不可直接下载", accessInfo);
  }

  try {
    const play = await VideoPlayURL(aid, bvid, cid, 80);
    return buildResolvedPlayInfo({aid, cid, bvid, partCount}, play, accessInfo);
  } catch (e) {
    throw new BilibiliPlayResolveError(errorMessage(e), accessInfo);
  }
}
