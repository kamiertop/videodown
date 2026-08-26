import {createSignal} from "solid-js";
import * as model from "@bindings/github.com/kamiertop/videodown/douyin/model/models";

const FOLLOW_CACHE_KEY = "douyin:user:follow-cache";

function loadFollowCache(): model.FollowResponse | null {
  if (typeof window === "undefined") return null;
  const raw = window.sessionStorage.getItem(FOLLOW_CACHE_KEY);
  if (!raw) return null;
  try {
    return (JSON.parse(raw));
  } catch {
    return null;
  }
}

function saveFollowCache(data: model.FollowResponse | null): void {
  if (typeof window === "undefined") return;
  if (!data) {
    window.sessionStorage.removeItem(FOLLOW_CACHE_KEY);
    return;
  }
  try {
    window.sessionStorage.setItem(FOLLOW_CACHE_KEY, JSON.stringify(data));
  } catch {
    // 缓存失败不影响主流程。
  }
}

const [followData, setFollowData] = createSignal<model.FollowResponse | null>(loadFollowCache());
const [followSearchInput, setFollowSearchInput] = createSignal("");
const [followSearchResults, setFollowSearchResults] = createSignal<model.SearchFollowResponse[] | null>(null);
const [followSearching, setFollowSearching] = createSignal(false);
const [followLoadingMore, setFollowLoadingMoreSignal] = createSignal(false);

export {
  followData,
  followSearchInput,
  setFollowSearchInput,
  followSearchResults,
  setFollowSearchResults,
  followSearching,
  setFollowSearching,
  followLoadingMore,
  setFollowLoadingMoreSignal as setFollowLoadingMore,
};

export function updateFollowData(data: model.FollowResponse | null): void {
  setFollowData(data);
  saveFollowCache(data);
}

export function appendFollowData(next: model.FollowResponse, total: number, currentItems: readonly model.FollowItem[]): void {
  const merged = ({
    ...next,
    followings: [...currentItems, ...(next.followings ?? [])],
    total: total ?? next.total,
  });
  updateFollowData(merged);
}

export function resetFollowSearch(): void {
  setFollowSearchInput("");
  setFollowSearchResults(null);
  setFollowSearching(false);
}
