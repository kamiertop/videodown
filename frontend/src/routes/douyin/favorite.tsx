import {createFileRoute} from '@tanstack/solid-router'
import {
  createEffect,
  createResource,
  createSignal,
  For,
  type JSXElement,
  Match,
  onCleanup,
  onMount,
  Switch,
} from "solid-js";
import {FavoriteVideo} from "../../../wailsjs/go/api/Douyin";
import {model} from "../../../wailsjs/go/models";
import FavoriteCollectionPanel from "../../components/douyin/FavoriteCollectionPanel.tsx";
import MixPanel from "../../components/douyin/MixPanel.tsx";
import VideoContentPanel from "../../components/douyin/VideoContentPanel.tsx";
import Toast from "../../components/Toast.tsx";
import {useToast} from "../../hooks/useToast.ts";

export const Route = createFileRoute('/douyin/favorite')({
  component: DouyinFavoritePage,
})

type FavoriteTab = 'collection' | 'video' | 'mix';

function tabLabel(tab: FavoriteTab): string {
  const labels: Record<FavoriteTab, string> = {
    collection: "收藏夹",
    video: "视频",
    mix: "合集",
  };

  return labels[tab];
}

function DouyinFavoritePage(): JSXElement {
  const [activeTab, setActiveTab] = createSignal<FavoriteTab>('collection');
  const [videoLoadingMore, setVideoLoadingMore] = createSignal(false);
  // 收藏视频接口只在用户真的切到“视频”tab 后再请求；收藏夹/合集 tab 有自己的延迟加载逻辑。
  const [videoEnabled, setVideoEnabled] = createSignal(false);
  const {message, type, showToast} = useToast();

  // 视频列表仅在切到“视频”标签后再触发，避免无意义请求。
  const [videoResult, {refetch: refetchVideos, mutate: mutateVideos}] = createResource(
    () => videoEnabled(),
    async (enabled) => {
      if (!enabled) return undefined;
      return FavoriteVideo(10, 0);
    },
  );

  const videos = () => videoResult()?.aweme_list ?? [];

  function switchTab(event: KeyboardEvent): void {
    // 收藏页只有三个平级 tab，用左右方向键做轻量切换。
    if (event.key === 'ArrowLeft') {
      event.preventDefault();
      setActiveTab((prev) => {
        if (prev === 'video') return 'collection';
        if (prev === 'mix') return 'video';
        return 'collection';
      });
      return;
    }

    if (event.key === 'ArrowRight') {
      event.preventDefault();
      setActiveTab((prev) => {
        if (prev === 'collection') return 'video';
        if (prev === 'video') return 'mix';
        return 'mix';
      });
    }
  }

  onMount(() => {
    window.addEventListener("keydown", switchTab);
  });
  onCleanup(() => {
    window.removeEventListener("keydown", switchTab);
  });

  createEffect(() => {
    if (activeTab() === 'video' && !videoEnabled()) {
      setVideoEnabled(true);
    }
  });

  function videoHasMore(): boolean {
    return Number(videoResult()?.has_more ?? 0) > 0;
  }

  async function loadMoreVideos(): Promise<void> {
    const current = videoResult();
    if (!current || Number(current.has_more ?? 0) <= 0 || videoLoadingMore()) {
      return;
    }

    setVideoLoadingMore(true);
    try {
      const next = await FavoriteVideo(20, current.cursor ?? videos().length);
      // createResource 的 mutate 只替换当前资源值；分页时手动把新旧 aweme_list 拼起来。
      mutateVideos(model.FavoriteVideoResponse.createFrom({
        ...next,
        aweme_list: [...videos(), ...(next.aweme_list ?? [])],
        uid: current.uid ?? next.uid,
        sec_uid: current.sec_uid ?? next.sec_uid,
      }));
    } catch (error) {
      showToast(error instanceof Error ? error.message : String(error), 'warning');
    } finally {
      setVideoLoadingMore(false);
    }
  }

  return (
    <section class="flex h-full min-h-0 flex-col p-2">
      <header class="shrink-0">
        <nav
          class="mb-2 flex flex-row justify-between rounded-2xl border border-base-300/90 bg-linear-to-b from-base-100 via-base-100 to-base-200/20 p-1.5 shadow-sm"
          role="tablist"
          aria-label="收藏分类"
        >
          <For each={['collection', 'video', 'mix'] as const}>
            {(tab) => (
              <button
                class={`min-h-6 flex-1 rounded-xl px-2 text-center text-xs font-semibold sm:text-sm ${
                  activeTab() === tab
                    ? "bg-primary/12 text-primary shadow-sm ring-1 ring-primary/25"
                    : "text-base-content/50 hover:bg-base-200/80 hover:text-base-content"
                } ${tab === 'mix' ? "opacity-90" : ""}`}
                type="button"
                role="tab"
                aria-selected={activeTab() === tab}
                onClick={() => setActiveTab(tab)}
              >
                {tabLabel(tab)}
              </button>
            )}
          </For>
        </nav>
      </header>

      {/* 外层只负责标签和边框，具体视频网格统一交给 VideoGrid 渲染。 */}
      <div class="min-h-0 flex-1 overflow-hidden rounded-xl border border-base-300 bg-base-100 shadow-sm">
        <div class="flex h-full min-h-0 flex-col">
          <Switch>
            <Match when={activeTab() === 'collection'}>
              {/* 收藏夹 tab：左侧收藏夹列表 + 右侧点击后加载视频。 */}
              <FavoriteCollectionPanel active={activeTab() === 'collection'} showToast={showToast}/>
            </Match>

            <Match when={activeTab() === 'video'}>
              {/* 收藏视频没有左侧集合列表，直接复用统一视频内容面板。 */}
              <VideoContentPanel
                kind="favorite-video"
                loading={videoResult.loading}
                error={videoResult.error ? String(videoResult.error) : ""}
                onRetry={() => void refetchVideos()}
                items={videos()}
                sourceName="收藏视频"
                fallbackAuthor="未知作者"
                showToast={showToast}
                hasMore={videoHasMore()}
                loadingMore={videoLoadingMore()}
                onLoadMore={() => void loadMoreVideos()}
              />
            </Match>
            <Match when={activeTab() === 'mix'}>
              {/* 收藏合集 tab 与用户合集共用 MixPanel，只是没有 userSecUserId。 */}
              <MixPanel active={activeTab() === 'mix'} showToast={showToast}/>
            </Match>
          </Switch>
        </div>
      </div>
      <Toast message={message()} type={type()}/>
    </section>
  );
}
