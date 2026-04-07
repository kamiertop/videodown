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
