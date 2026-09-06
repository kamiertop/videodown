import {PlayHistory} from '@bindings/github.com/kamiertop/videodown/bilibili/api/bilibili';
import * as model from '@bindings/github.com/kamiertop/videodown/bilibili/model/models';
import {createFileRoute} from '@tanstack/solid-router'
import {createSignal, type JSXElement, Show} from 'solid-js';
import VideoListSection from '../../components/bilibili/VideoListSection';
import DetailError from '../../components/DetailError';
import type {MediaCardItem} from '../../lib/model';

export const Route = createFileRoute('/bilibili/play-history')({component: PlayHistoryPage});

function PlayHistoryPage(): JSXElement {
  const [items, setItems] = createSignal<model.PlayHistoryItem[]>([]);
  const [cursor, setCursor] = createSignal<model.PlayHistoryCursor>({max: 0, view_at: 0, business: '', ps: 20});
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal('');
  const [hasMore, setHasMore] = createSignal(true);

  async function load(more = false): Promise<void> {
    if (loading() || (more && !hasMore())) return;
    setLoading(true);
    setError('');
    try {
      const current = more ? cursor() : {max: 0, view_at: 0};
      const data = await PlayHistory(current.max, current.view_at);
      setItems(more ? [...items(), ...(data.list ?? [])] : (data.list ?? []));
      setCursor(data.cursor);
      setHasMore((data.list?.length ?? 0) > 0 && data.cursor.max !== current.max);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setLoading(false);
    }
  }

  void load();

  function title(item: model.PlayHistoryItem): string {
    return item.title || item.long_title || '未命名视频';
  }

  function link(item: model.PlayHistoryItem): string {
    return item.history.bvid ? `https://www.bilibili.com/video/${item.history.bvid}` : '';
  }

  function media(item: model.PlayHistoryItem, index: number): MediaCardItem {
    return {
      id: index + 1, title: title(item), cover: item.cover || '', duration: item.duration || 0,
      bvid: item.history.bvid || '', link: link(item), upperName: item.author_name || '未知 UP 主',
      play: 0, danmaku: 0, pubtime: item.view_at, sourceListName: 'B站播放历史'
    };
  }

  return <div class="flex h-full min-h-0 flex-col p-4">
    <Show when={!error()} fallback={<DetailError message={error()} onRetry={() => void load()}/>}>
      <Show when={items().length > 0} fallback={<div
          class="flex min-h-0 flex-1 items-center justify-center text-sm text-base-content/50">暂无播放历史</div>}>
        <div class="flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-base-300 bg-base-100">
          <VideoListSection title="播放历史" mediaCount={items().length} medias={() => items().map(media)}
                            selectionResetKey={() => cursor().max} hasMore={hasMore} loadingMore={loading}
                            onLoadMore={() => void load(true)}/>
        </div>
      </Show>
    </Show>
  </div>;
}
