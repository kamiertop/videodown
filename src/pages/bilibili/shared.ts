export type BilibiliMenuKey = "video" | "uploader" | "series" | "favorite";

export interface BilibiliProfile {
    mid: number;
    name: string;
    sex: string;
    face: string;
    face_data_url?: string;
    sign: string;
    rank: number;
    level: number;
    birthday: number;
}

export interface BilibiliMyInfo {
    coins?: number;
    profile?: BilibiliProfile;
}

export interface BilibiliMenuItem {
    key: BilibiliMenuKey;
    path: string;
    label: string;
    description: string;
}

export const bilibiliMenuItems: BilibiliMenuItem[] = [
    {key: "video", path: "/bilibili/video", label: "单视频下载", description: "粘贴链接下载单个视频"},
    {key: "uploader", path: "/bilibili/uploader", label: "UP主视频下载", description: "按 UP 主页批量下载"},
    {key: "series", path: "/bilibili/series", label: "合集/系列下载", description: "下载合集或系列内容"},
    {key: "favorite", path: "/bilibili/favorite", label: "收藏夹视频下载", description: "按收藏夹批量下载"}
];
