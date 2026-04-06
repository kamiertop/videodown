import {createFileRoute, useNavigate} from '@tanstack/solid-router'
import {createSignal, For, Match, onCleanup, onMount, Show, Switch, type JSXElement} from "solid-js";
import {FollowList, Info} from "../../../../wailsjs/go/api/BiliBili";
import {model} from "../../../../wailsjs/go/models";
import UpCommonLayout from "../../../components/UpCommonLayout";

const PAGE_SIZE = 30;

export const Route = createFileRoute('/bilibili/up/')({
    component: UpIndex,
})

function UpIndex(): JSXElement {
    const navigate = useNavigate();
    const [loading, setLoading] = createSignal(true);
    const [parsing, setParsing] = createSignal(false);
    const [followData, setFollowData] = createSignal<model.FollowData | null>(null);
    const [page, setPage] = createSignal(1);
    const [errorText, setErrorText] = createSignal('');
    const [searchInput, setSearchInput] = createSignal('');
    const [parsedInfo, setParsedInfo] = createSignal<model.UserInfoData | null>(null);

    let errorToastTimer: number | undefined;
    const showErrorToast = (message: string) => {
        setErrorText(message);
        if (errorToastTimer !== undefined) {
            window.clearTimeout(errorToastTimer);
        }
        errorToastTimer = window.setTimeout(() => {
            setErrorText('');
            errorToastTimer = undefined;
        }, 1500)
    }

    const totalPages = () => {
        const total = followData()?.total ?? 0;
        return Math.max(1, Math.ceil(total / PAGE_SIZE));
    };

    const loadFollows = async (pn: number) => {
        setLoading(true);
        setErrorText('');
        try {
            const data = await FollowList(pn, PAGE_SIZE);
            setFollowData(data);
            setPage(pn);
        } catch (error) {
            showErrorToast(error instanceof Error ? error.message : String(error));
        } finally {
            setLoading(false);
        }
    };

    const goToUp = (mid: number) => {
        void navigate({to: '/bilibili/up/$mid', params: {mid: String(mid)}});
    };

    const handleSearch = async () => {
        const raw = searchInput().trim();
        if (!raw) return;

        if (parsing()) return;
        setParsing(true);
        setErrorText('');
        setParsedInfo(null);
        try {
            const info = await Info(raw);
            setParsedInfo(info);
        } catch (error) {
            showErrorToast(error instanceof Error ? error.message : String(error));
        } finally {
            setParsing(false);
        }
    };

    const handleKeyDown = (e: KeyboardEvent) => {
        if (e.key === 'Enter') void handleSearch();
    };

    onMount(() => {
        void loadFollows(1);
    });
    onCleanup(() => {
        if (errorToastTimer !== undefined) {
            window.clearTimeout(errorToastTimer);
        }
    })

    return (
        <UpCommonLayout
            headerLeft={
                <div class="flex items-center gap-2">
                    <h2 class="text-sm font-bold text-base-content">我的关注</h2>
                    <Show when={followData()}>
                        <span
                            class="rounded-full bg-primary/10 px-2 py-0.5 text-xs font-bold tabular-nums text-primary">
                            {followData()!.total}
                        </span>
                    </Show>
                    <button
                        class="flex h-6 w-6 items-center justify-center rounded transition-colors hover:bg-base-200 disabled:cursor-not-allowed"
                        onClick={() => void loadFollows(page())}
                        disabled={loading()}
                        title="刷新"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg"
                             class={`h-3.5 w-3.5 text-base-content/50 ${loading() ? 'animate-spin' : ''}`}
                             fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                            <path stroke-linecap="round" stroke-linejoin="round"
                                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
                        </svg>
                    </button>
                </div>
            }
            headerRight={
                <div class="relative flex items-center gap-2">
                    <Show when={errorText()}>
                        <div
                            class="pointer-events-none absolute right-full top-1/2 z-50 mr-2 -translate-y-1/2"
                            role="alert"
                        >
                            <div
                                class="w-max min-w-64 max-w-sm whitespace-normal break-words rounded-lg border border-error/30 bg-error/10 px-3 py-2 text-xs font-semibold leading-relaxed text-error shadow-md"
                            >
                                {errorText()}
                            </div>
                        </div>
                    </Show>
                    <input
                        type="text"
                        class="input input-sm input-bordered w-72"
                        placeholder="输入 UP主 mid / 空间链接（支持不带 https）"
                        value={searchInput()}
                        onInput={(e) => setSearchInput(e.currentTarget.value)}
                        onKeyDown={handleKeyDown}
                    />
                    <button class="btn btn-primary btn-sm" onClick={() => void handleSearch()} disabled={parsing()}>
                        <Show when={!parsing()} fallback={<span class="loading loading-spinner loading-xs"></span>}>
                            解析
                        </Show>
                    </button>
                </div>
            }
            headerBelow={
                parsedInfo() ? (
                    <div class="rounded-xl border border-base-300 bg-base-100 px-4 py-3">
                        <div
                            class="flex flex-wrap items-center gap-3 rounded-xl border border-base-300 bg-base-200/30 p-3">
                            <div class="h-10 w-10 overflow-hidden rounded-full bg-base-200 ring-2 ring-base-200">
                                <img
                                    src={parsedInfo()!.face}
                                    alt={parsedInfo()!.name}
                                    referrerPolicy="no-referrer"
                                    class="h-full w-full object-cover"
                                />
                            </div>
                            <div class="min-w-0 flex-1">
                                <div class="flex flex-wrap items-center gap-2">
                                    <span
                                        class="truncate text-sm font-bold text-base-content">{parsedInfo()!.name}</span>
                                    <span class="badge badge-outline badge-sm">mid: {parsedInfo()!.mid}</span>
                                    <span
                                        class={`badge badge-sm ${parsedInfo()!.is_followed ? 'badge-primary' : 'badge-ghost'}`}>
                                        {parsedInfo()!.is_followed ? '已关注' : '未关注'}
                                    </span>
                                    <span class="badge badge-outline badge-sm">Lv.{parsedInfo()!.level}</span>
                                </div>
                                <div class="mt-0.5 line-clamp-1 text-xs text-base-content/60">
                                    {parsedInfo()!.sign?.trim() ? parsedInfo()!.sign : '这个人很懒，什么也没有写~'}
                                </div>
                            </div>
                            <div class="flex items-center gap-2">
                                <button class="btn btn-outline btn-sm" onClick={() => setParsedInfo(null)}>
                                    清除
                                </button>
                                <button class="btn btn-primary btn-sm" onClick={() => goToUp(parsedInfo()!.mid)}>
                                    进入详情
                                </button>
                            </div>
                        </div>
                    </div>
                ) : undefined
            }
        >
            {/* 关注列表 */}
            <div class="min-h-0 flex-1 overflow-auto rounded-xl border border-base-300 bg-base-100">
                <Switch>
                    <Match when={loading()}>
                        <div class="flex h-full items-center justify-center">
                            <span class="loading loading-spinner loading-md text-primary"></span>
                        </div>
                    </Match>
                    <Match when={!followData()?.list?.length}>
                        <div class="flex h-full items-center justify-center text-base-content/40">
                            <div class="text-center">
                                <svg xmlns="http://www.w3.org/2000/svg" class="mx-auto h-14 w-14" fill="none"
                                     viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
                                    <path stroke-linecap="round" stroke-linejoin="round"
                                          d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"/>
                                </svg>
                                <p class="mt-3 text-sm font-semibold text-base-content/60">暂无关注</p>
                                <p class="mt-1 text-xs text-base-content/50">可以通过上方搜索框直接输入 UP主 mid
                                    解析</p>
                            </div>
                        </div>
                    </Match>
                    <Match when={followData()?.list?.length}>
                        <div class="p-4">
                            <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
                                <For each={followData()!.list}>
                                    {(up) => {
                                        return (
                                            <button
                                                class="group flex flex-col items-center gap-2 rounded-xl border border-base-300 bg-base-100 p-4 transition-all duration-150 hover:-translate-y-px hover:border-primary/40 hover:shadow-md active:scale-[0.98]"
                                                onClick={() => goToUp(up.mid)}
                                            >
                                                <div class="relative">
                                                    <div
                                                        class="h-16 w-16 overflow-hidden rounded-full bg-base-200 ring-2 ring-base-200 transition group-hover:ring-primary/30">
                                                        <img
                                                            src={up.face}
                                                            alt={up.uname}
                                                            referrerPolicy="no-referrer"
                                                            class="h-full w-full object-cover"
                                                            loading="lazy"
                                                        />
                                                    </div>
                                                </div>
                                                <span
                                                    class="max-w-full truncate text-sm font-semibold text-base-content group-hover:text-primary">
                                                    {up.uname}
                                                </span>
                                                <span
                                                    class="line-clamp-2 max-w-full text-center text-xs leading-relaxed text-base-content/50">
                                                    {up.sign || '这个人很懒，什么也没有写~'}
                                                </span>
                                            </button>
                                        );
                                    }}
                                </For>
                            </div>

                            {/* 翻页 */}
                            <Show when={totalPages() > 1}>
                                <div class="mt-4 flex items-center justify-center gap-2">
                                    <button
                                        class="btn btn-outline btn-sm"
                                        disabled={page() <= 1}
                                        onClick={() => void loadFollows(page() - 1)}
                                    >
                                        上一页
                                    </button>
                                    <span class="text-sm tabular-nums text-base-content/70">
                                        {page()} / {totalPages()}
                                    </span>
                                    <button
                                        class="btn btn-outline btn-sm"
                                        disabled={page() >= totalPages()}
                                        onClick={() => void loadFollows(page() + 1)}
                                    >
                                        下一页
                                    </button>
                                </div>
                            </Show>
                        </div>
                    </Match>
                </Switch>
            </div>

        </UpCommonLayout>
    );
}
