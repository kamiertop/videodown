import {createFileRoute} from '@tanstack/solid-router'
import {createSignal, type JSXElement, onMount} from "solid-js";
import {Dynamic} from "@bindings/github.com/kamiertop/videodown/douyin/api/douyin";
import * as model from "@bindings/github.com/kamiertop/videodown/douyin/model/models";
import VideoContentPanel from "../../components/douyin/VideoContentPanel";
import Toast from "../../components/Toast";
import {useToast} from "../../hooks/useToast";

export const Route = createFileRoute('/douyin/dynamic')({ component: DouyinDynamicPage })

function DouyinDynamicPage(): JSXElement {
  const {showToast, message, type} = useToast();
  const [items, setItems] = createSignal<model.AwemeItem[]>([]);
  const [cursor, setCursor] = createSignal(0);
  const [hasMore, setHasMore] = createSignal(false);
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal("");

  async function load(nextCursor: number, append = false): Promise<void> {
    if (loading()) return;
    setLoading(true); setError("");
    try {
      const response = await Dynamic(nextCursor);
      const next = (response.data ?? []).map((entry) => entry.aweme).filter(Boolean);
      setItems((current) => append ? [...current, ...next] : next);
      setCursor(response.cursor ?? nextCursor);
      setHasMore(Number(response.has_more) > 0);
    } catch (err) {
      const text = err instanceof Error ? err.message : String(err);
      setError(text); showToast(text, "error");
    } finally { setLoading(false); }
  }

  onMount(() => { void load(0); });

  return <div class="flex h-full min-h-0 flex-col bg-base-200/40 p-4">
    <VideoContentPanel
      kind="follow-dynamic"
      title="关注动态"
      items={items()}
      loading={loading() && items().length === 0}
      error={error()}
      onRetry={() => void load(0)}
      sourceName="关注动态"
      fallbackAuthor="关注的用户"
      showToast={showToast}
      refreshing={loading() && items().length > 0}
      onRefresh={() => void load(0)}
      hasMore={hasMore()}
      loadingMore={loading()}
      onLoadMore={() => void load(cursor(), true)}
    />
    <Toast message={message()} type={type()} />
  </div>;
}
