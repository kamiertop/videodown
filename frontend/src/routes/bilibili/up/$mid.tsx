import {createFileRoute, Link} from '@tanstack/solid-router'
import {createSignal, Match, onCleanup, onMount, Switch, type JSXElement} from "solid-js";
import {Info} from "../../../../wailsjs/go/api/BiliBili";
import {model} from "../../../../wailsjs/go/models";
import ErrorToast from "../../../components/ErrorToast";
import UpCommonLayout from "../../../components/UpCommonLayout";


export const Route = createFileRoute('/bilibili/up/$mid')({
    component: UpDetail,
})

function UpDetail(): JSXElement {
    const params = Route.useParams();
    const [loading, setLoading] = createSignal(true);
    const [errorText, setErrorText] = createSignal('');
    const [info, setInfo] = createSignal<model.UserInfoData | null>(null);

    let errorToastTimer: number | undefined;
    const showErrorToast = (message: string) => {
        setErrorText(message);
        if (errorToastTimer !== undefined) {
            window.clearTimeout(errorToastTimer);
        }
        errorToastTimer = window.setTimeout(() => {
            setErrorText('');
            errorToastTimer = undefined;
        }, 3000)
    }

    const loadInfo = async (midOrSpaceUrl: string) => {
        setLoading(true);
        setErrorText('');
        try {
            const data = await Info(midOrSpaceUrl);
            setInfo(data);
        } catch (error) {
            showErrorToast(error instanceof Error ? error.message : String(error));
        } finally {
            setLoading(false);
        }
    };

    onMount(() => {
        void loadInfo(params().mid);
    })
    onCleanup(() => {
        if (errorToastTimer !== undefined) {
            window.clearTimeout(errorToastTimer);
        }
    })

    return (
        <UpCommonLayout
            headerLeft={
                <>
                    <Link
                        to="/bilibili/up"
                        class="btn btn-ghost btn-sm gap-1"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24"
                             stroke="currentColor" stroke-width="2">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7"/>
                        </svg>
                        返回
                    </Link>
                    <div class="h-5 w-px bg-base-300"></div>
                    <h2 class="text-sm font-bold text-base-content">UP主详情</h2>
                    <span class="rounded-full bg-base-200 px-2 py-0.5 text-xs tabular-nums text-base-content/60">
                        mid: {params().mid}
                    </span>
                </>
            }
            headerRight={
                <Switch>
                    <Match when={loading()}>
                        <div class="flex items-center gap-2">
                            <span class="loading loading-spinner loading-xs text-primary"></span>
                            <span class="text-xs text-base-content/50">获取UP主信息...</span>
                        </div>
                    </Match>
                    <Match when={!loading() && info()}>
                        <div class="flex min-w-0 items-center gap-2">
                            <div class="h-8 w-8 shrink-0 overflow-hidden rounded-full bg-base-200 ring-2 ring-base-200">
                                <img
                                    src={info()!.face}
                                    alt={info()!.name}
                                    referrerPolicy="no-referrer"
                                    class="h-full w-full object-cover"
                                />
                            </div>
                            <div class="min-w-0">
                                <div class="flex min-w-0 items-center gap-2">
                                    <span class="max-w-[16rem] truncate text-sm font-black text-base-content">
                                        {info()!.name}
                                    </span>
                                    <span class="badge badge-outline badge-sm">Lv.{info()!.level}</span>
                                    <span
                                        class={`badge badge-sm ${info()!.is_followed ? 'badge-primary' : 'badge-ghost'}`}
                                    >
                                        {info()!.is_followed ? '已关注' : '未关注'}
                                    </span>
                                </div>
                            </div>
                        </div>
                    </Match>
                </Switch>
            }
        >
            {/* 内容区 - 这里将用于投稿/合集/系列等 */}
            <div class="h-full p-4">
                <Switch>
                    <Match when={loading()}>
                        <div class="flex h-full items-center justify-center">
                            <span class="loading loading-spinner loading-md text-primary"></span>
                        </div>
                    </Match>
                    <Match when={!loading()}>
                        <div class="flex h-full items-center justify-center">
                            <div class="text-center">
                                <svg xmlns="http://www.w3.org/2000/svg"
                                     class="mx-auto h-16 w-16 text-base-content/20" fill="none"
                                     viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
                                    <path stroke-linecap="round" stroke-linejoin="round"
                                          d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                                </svg>
                                <p class="mt-3 text-sm font-semibold text-base-content/60">UP主投稿视频</p>
                                <p class="mt-1 max-w-xs text-xs text-base-content/40">
                                    待接入后端投稿视频 / 合集 / 系列 API 后，这里将展示对应内容，支持批量选择下载
                                </p>
                            </div>
                        </div>
                    </Match>
                </Switch>
            </div>

            <ErrorToast message={errorText()}/>
        </UpCommonLayout>
    );
}
