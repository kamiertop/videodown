import {createFileRoute} from '@tanstack/solid-router'
import {createEffect, createMemo, createSignal, For, Match, onCleanup, onMount, Show, Switch, type JSXElement} from "solid-js";
import {FavoritesList} from "../../../wailsjs/go/api/BiliBili";
import type {model} from "../../../wailsjs/go/models";
import ErrorToast from "../../components/ErrorToast";
import StatusToast from "../../components/StatusToast";

export const Route = createFileRoute('/bilibili/favorite')({
    component: Favorite,
})

/** 与行 DOM 一致；少用圆角/阴影/ring，滚动才顺滑 */
const ROW_HEIGHT = 48;
const OVERSCAN = 6;

function FavoriteScrollBody(props: { list: () => readonly model.FavoriteItem[]; listEpoch: () => number }): JSXElement {
    const [startIndex, setStartIndex] = createSignal(0);
    const [viewportHeight, setViewportHeight] = createSignal(0);
    let viewportEl: HTMLDivElement | undefined;
    let resizeObserver: ResizeObserver | undefined;
    let scrollRaf = 0;

    const fullList = createMemo(() => props.list());

    const visibleRows = createMemo(() => Math.ceil(viewportHeight() / ROW_HEIGHT) + OVERSCAN * 2);
    const endIndex = createMemo(() => Math.min(fullList().length, startIndex() + visibleRows()));
    const visibleList = createMemo(() => fullList().slice(startIndex(), endIndex()));
    const topPlaceholderHeight = createMemo(() => startIndex() * ROW_HEIGHT);
    const bottomPlaceholderHeight = createMemo(() => Math.max(0, (fullList().length - endIndex()) * ROW_HEIGHT));

    const applyScrollTop = (scrollTop: number) => {
        const next = Math.max(0, Math.floor(scrollTop / ROW_HEIGHT) - OVERSCAN);
        if (next !== startIndex()) {
            setStartIndex(next);
        }
    };

    createEffect(() => {
        props.listEpoch();
        queueMicrotask(() => {
            const el = viewportEl;
            if (!el) {
                return;
            }
            el.scrollTop = 0;
            setStartIndex(0);
        });
    });

    onMount(() => {
        const el = viewportEl;
        if (!el) {
            return;
        }

        setViewportHeight(el.clientHeight);

        resizeObserver = new ResizeObserver(() => {
            if (!viewportEl) {
                return;
            }
            setViewportHeight(viewportEl.clientHeight);
            applyScrollTop(viewportEl.scrollTop);
        });
        resizeObserver.observe(el);

        const onScroll = () => {
            if (scrollRaf) {
                return;
            }
            scrollRaf = requestAnimationFrame(() => {
                scrollRaf = 0;
                if (!viewportEl) {
                    return;
                }
                applyScrollTop(viewportEl.scrollTop);
            });
        };

        el.addEventListener('scroll', onScroll, {passive: true});

        onCleanup(() => {
            resizeObserver?.disconnect();
            resizeObserver = undefined;
            el.removeEventListener('scroll', onScroll);
            if (scrollRaf) {
                cancelAnimationFrame(scrollRaf);
                scrollRaf = 0;
            }
        });
    });

    return (
        <div
            class="min-h-0 flex-1 overflow-y-auto overscroll-y-contain px-3"
            ref={(el) => (viewportEl = el)}
        >
            <div style={{height: `${topPlaceholderHeight()}px`}} aria-hidden="true"></div>
            <For each={visibleList()}>
                {(item) => (
                    <div class="flex h-12 items-center gap-3 border-b border-base-200 px-1 text-sm">
                        <span class="min-w-0 flex-1 truncate font-semibold text-base-content">
                            {item.title}
                        </span>
                        <span class="shrink-0 tabular-nums text-xs text-primary">
                            {item.media_count} 个视频
                        </span>
                    </div>
                )}
            </For>
            <div style={{height: `${bottomPlaceholderHeight()}px`}} aria-hidden="true"></div>
        </div>
    );
}

function Favorite(): JSXElement {
    const [loading, setLoading] = createSignal(true);
    const [favorites, setFavorites] = createSignal<model.FavoritesData | null>(null);
    const [errorText, setErrorText] = createSignal('');
    const [statusText, setStatusText] = createSignal('');
    const [listEpoch, setListEpoch] = createSignal(0);

    const listAccessor = createMemo(() => favorites()?.list || []);

    const loadFavorites = async () => {
        setLoading(true);
        setErrorText('');
        try {
            const data = await FavoritesList();
            setFavorites(data);
            setListEpoch((n) => n + 1);
            setStatusText(`已加载 ${data.list?.length || 0} 个收藏夹`);
        } catch (error) {
            setErrorText(error instanceof Error ? error.message : String(error));
        } finally {
            setLoading(false);
        }
    }

    onMount(() => {
        void loadFavorites();
    })

    return (
        <section class="h-full overflow-hidden bg-base-200/40 px-4 py-4 md:px-6 md:py-5">
            <div class="flex h-full min-h-0 flex-col gap-3">
                <div class="flex shrink-0 items-center justify-between gap-3 rounded-xl border border-base-300 bg-base-100/90 px-4 py-3">
                    <div class="min-w-0">
                        <h1 class="text-lg font-bold text-base-content md:text-xl">我的收藏夹</h1>
                        <p class="text-sm text-base-content/70">
                            共 {favorites()?.count ?? 0} 个收藏夹
                        </p>
                    </div>
                    <button
                        class="btn btn-primary btn-sm shrink-0"
                        onClick={() => void loadFavorites()}
                        disabled={loading()}
                    >
                        刷新
                    </button>
                </div>

                <div class="flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-base-300 bg-base-100/90">
                    <Switch>
                        <Match when={loading()}>
                            <div class="flex h-full min-h-[12rem] items-center justify-center">
                                <span class="loading loading-spinner loading-lg text-primary"></span>
                            </div>
                        </Match>

                        <Match when={!loading() && (favorites()?.list?.length || 0) === 0}>
                            <div class="flex h-full min-h-[12rem] items-center justify-center text-base-content/60">
                                暂无收藏夹数据
                            </div>
                        </Match>

                        <Match when={!loading() && (favorites()?.list?.length || 0) > 0}>
                            <FavoriteScrollBody list={listAccessor} listEpoch={listEpoch}/>
                        </Match>
                    </Switch>
                </div>
            </div>

            <Show when={statusText() && !loading()}>
                <StatusToast message={statusText()} tone="success"/>
            </Show>
            <ErrorToast message={errorText()}/>
        </section>
    )
}
