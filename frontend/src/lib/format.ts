/** B 站稿件时长字符串 `mm:ss` 或 `hh:mm:ss` → 秒 */
export function parseBilibiliLengthToSeconds(length: string): number {
    const parts = length.trim().split(':').map((p) => parseInt(p, 10));
    if (parts.some((n) => Number.isNaN(n))) return 0;
    if (parts.length === 2) return parts[0] * 60 + parts[1];
    if (parts.length === 3) return parts[0] * 3600 + parts[1] * 60 + parts[2];
    return 0;
}

export function formatDuration(totalSeconds: number): string {
    const safe = Number.isFinite(totalSeconds) ? Math.max(0, Math.floor(totalSeconds)) : 0;
    const hours = Math.floor(safe / 3600);
    const minutes = Math.floor((safe % 3600) / 60);
    const seconds = safe % 60;
    if (hours > 0) {
        return `${hours}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
    }
    return `${minutes}:${String(seconds).padStart(2, '0')}`;
}

export function formatCount(value?: number): string {
    if (!value || value <= 0) return '0';
    if (value >= 10000) return `${(value / 10000).toFixed(1)}万`;
    return String(value);
}

export function formatDate(timestamp?: number): string {
    if (!timestamp) return '';
    const d = new Date(timestamp * 1000);
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    return `${y}-${m}-${day}`;
}

/**
 * 从 B 站视频链接中提取 BV 号
 * 支持的格式:
 * - BV 号本身: BV1xx411c7mD
 * - 完整 URL: https://www.bilibili.com/video/BV1xx411c7mD
 * - 短链接: https://b23.tv/BV1xx411c7mD
 * - AV 号: av12345678 (需要转换)
 */
export function extractBvid(input: string): string | null {
    if (!input || !input.trim()) {
        return null;
    }

    const trimmed = input.trim();

    // 直接匹配 BV 号 (BV 后跟 10 位字母数字)
    const bvPattern: RegExp = /BV([A-Za-z0-9]{10})/;
    const match = trimmed.match(bvPattern);
    if (match) {
        return `BV${match[1]}`;
    }

    // 匹配 AV 号 (av 后跟数字)
    const avPattern = /(?:av|AV)(\d+)/;
    const avMatch = trimmed.match(avPattern);
    if (avMatch) {
        // TODO: 如果需要支持 AV 转 BV,可以在这里调用后端 API 转换
        // 目前先返回 null,提示不支持
        console.warn('AV 号需要转换为 BV 号,暂不支持');
        return null;
    }

    return null;
}

/**
 * 从 B 站视频链接中解析分 P 序号（`?p=` / `&p=`，或 `page=`），缺省为 1。
 */
export function extractBilibiliPartIndex(input: string): number {
    const s = input.trim();
    const m = /(?:^|[?&])p=(\d+)/i.exec(s) || /(?:^|[?&])page=(\d+)/i.exec(s);
    if (!m) return 1;
    const n = parseInt(m[1], 10);
    if (!Number.isFinite(n) || n < 1) return 1;
    return n;
}
