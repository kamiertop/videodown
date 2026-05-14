import {VideoDetailConciseBvid} from "../../../wailsjs/go/api/BiliBili";
import {model} from "../../../wailsjs/go/models";
import type {MediaCardItem} from "../model";

type VideoDetailView = model.VideoDetailView;

// 列表页不为了展示分 P 标记额外查详情；只有用户真正加入下载队列时才展开分 P。
function normalizeBiliCover(url: string | undefined): string {
  if (!url) return "";
  if (url.startsWith("//")) return `https:${url}`;
  return url;
}

function uniqueMediaCards(items: MediaCardItem[]): MediaCardItem[] {
  const seen = new Set<string>();
  return items.filter((item) => {
    const bvid = item.bvid?.trim().toUpperCase();
    const cid = item.cid;
    const key = bvid && cid != null && cid > 0 ? `${bvid}:${cid}` : bvid || String(item.id);
    if (!key || seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}

function partListTitle(mainTitle: string, partName: string, pageIndex: number): string {
  const part = partName?.trim();
  if (part) return `${mainTitle} · P${pageIndex} ${part}`;
  return `${mainTitle} · P${pageIndex}`;
}

function pagesToMediaCards(view: VideoDetailView, item: MediaCardItem): MediaCardItem[] {
  const pages = view.pages ?? [];
  const bvid = view.bvid || item.bvid || "";
  const mainTitle = view.title?.trim() || item.title || "未命名稿件";
  const isUpArchive = item.sourceListKind === "全部投稿";

  // UP 投稿列表只知道整稿件。展开为分 P 后，目录来源改成“分P”，后端会落到 UP/稿件名/P标题。
  return uniqueMediaCards(
      pages.map((p) => ({
        ...item,
        id: Number(p.cid) || p.page || item.id,
        title: partListTitle(mainTitle, p.part, p.page),
        cover: normalizeBiliCover(p.first_frame || view.pic || item.cover),
        duration: Number(p.duration) || item.duration,
        bvid,
        cid: Number(p.cid) || undefined,
        link: bvid ? `https://www.bilibili.com/video/${bvid}?p=${p.page}` : item.link,
        upperName: view.owner?.name || item.upperName,
        play: view.stat?.view ?? item.play,
        danmaku: view.stat?.danmaku ?? item.danmaku,
        pubtime: view.pubdate ?? item.pubtime,
        sourceListName: isUpArchive ? mainTitle : item.sourceListName,
        sourceListKind: isUpArchive ? "分P" : item.sourceListKind,
      })),
  );
}

export async function expandBilibiliDownloadItem(item: MediaCardItem): Promise<MediaCardItem[]> {
  const bvid = item.bvid?.trim();
  if (!bvid) return [];

  const detail = await VideoDetailConciseBvid(bvid);
  const view = detail.view;
  if ((view.pages?.length ?? 0) <= 1) return [item];

  return pagesToMediaCards(view, item);
}

export async function expandBilibiliDownloadItems(items: MediaCardItem[]): Promise<MediaCardItem[]> {
  const result: MediaCardItem[][] = [];
  const concurrency = 4;
  let index = 0;

  // 批量下载可能包含很多投稿；限制详情请求并发，同时用二维数组保留原顺序。
  async function worker(): Promise<void> {
    for (; ;) {
      const currentIndex = index;
      const current = items[index];
      index += 1;
      if (!current) return;
      result[currentIndex] = await expandBilibiliDownloadItem(current);
    }
  }

  await Promise.all(Array.from({length: Math.min(concurrency, items.length)}, () => worker()));
  return uniqueMediaCards(result.flat());
}
