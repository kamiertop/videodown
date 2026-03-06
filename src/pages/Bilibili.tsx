import {A, type RouteSectionProps, useLocation, useNavigate} from "@solidjs/router";
import {invoke} from "@tauri-apps/api/core";
import {createMemo, createSignal, For, Match, onMount, Show, Switch} from "solid-js";

import BilibiliQrcode from "../components/BilibiliQrcode.tsx";
import bilibiliAvatarFallback from "../assets/bilibili_256_256.svg";
import {useBilibiliAccountStore} from "./bilibili/account.ts";
import {bilibiliMenuItems, BilibiliMyInfo} from "./bilibili/shared.ts";

const avatarPageMeta = {
    label: "当前账号",
    description: "查看账号头像、昵称和登录状态。"
};

export default function Bilibili(props: RouteSectionProps) {
    const navigate = useNavigate();
    const location = useLocation();
    const [isCheckingLogin, setIsCheckingLogin] = createSignal(true);
    const [loggedIn, setLoggedIn] = createSignal(false);
    const [submitError, setSubmitError] = createSignal<string | null>(null);
    const [avatarLoadFailed, setAvatarLoadFailed] = createSignal(false);
    const {
        accountInfo,
        setAccountInfo,
        accountInfoLoading,
        setAccountInfoLoading
    } = useBilibiliAccountStore();

    async function refreshAccountInfo() {
        setAccountInfoLoading(true);
        try {
            const response = await invoke<BilibiliMyInfo>("my_info");
            setAccountInfo(response);
            setAvatarLoadFailed(false);
        } catch (error) {
            setSubmitError(error instanceof Error ? error.message : String(error));
        } finally {
            setAccountInfoLoading(false);
        }
    }

    onMount(async () => {
        try {
            const status = await invoke<boolean>("is_logged_in");
            setLoggedIn(status);
            if (status) {
                await refreshAccountInfo();
                if (location.pathname === "/bilibili") {
                    navigate("/bilibili/video", {replace: true});
                }
            } else {
                setAccountInfo(null);
            }
        } catch (error) {
            setSubmitError(error instanceof Error ? error.message : String(error));
        } finally {
            setIsCheckingLogin(false);
        }
    });

    const activeMenuItem = createMemo(() => {
        if (location.pathname === "/bilibili/avatar") {
            return avatarPageMeta;
        }
        return bilibiliMenuItems.find((item) => location.pathname === item.path) ?? bilibiliMenuItems[0];
    });

    const accountProfile = createMemo(() => accountInfo()?.profile);
    const accountDisplayName = createMemo(() => accountProfile()?.name?.trim() || "Bilibili 用户");
    const accountDisplayFace = createMemo(() => {
        const embeddedFace = accountProfile()?.face_data_url?.trim();
        const remoteFace = accountProfile()?.face?.trim();
        if (avatarLoadFailed()) {
            return bilibiliAvatarFallback;
        }
        if (embeddedFace) {
            return embeddedFace;
        }
        if (remoteFace) {
            return remoteFace;
        }
        return bilibiliAvatarFallback;
    });
    const accountStatusText = createMemo(() => {
        if (accountInfoLoading()) {
            return "正在加载账号信息";
        }
        if (accountProfile()) {
            return "点击查看账号页";
        }
        return "账号信息暂不可用";
    });

    function handleLoginSuccess() {
        setLoggedIn(true);
        setSubmitError(null);
        void refreshAccountInfo();
        navigate("/bilibili/video", {replace: true});
    }

    return (
        <div class="h-full bg-neutral-50 text-neutral-900">
            <Switch>
                <Match when={isCheckingLogin()}>
                    <div class="h-full flex items-center justify-center text-sm text-neutral-500">
                        正在检查 B站 登录状态...
                    </div>
                </Match>

                <Match when={!loggedIn()}>
                    <div class="h-full flex flex-col items-center justify-center gap-4 px-4">
                        <div class="text-center space-y-2">
                            <h1 class="text-2xl font-bold">先登录 B站 账号</h1>
                            <p class="text-sm text-neutral-500">登录完成后，会自动进入下载工作台。</p>
                        </div>
                        <BilibiliQrcode onLoginSuccess={handleLoginSuccess}/>
                        <Show when={submitError()}>
                            <p class="text-sm text-red-600">{submitError()}</p>
                        </Show>
                    </div>
                </Match>

                <Match when={loggedIn()}>
                    <div class="flex h-full overflow-hidden border border-neutral-200 bg-white shadow-sm">
                        <aside class="w-64 border-r border-neutral-200 bg-neutral-50/80 p-4 flex flex-col gap-4">
                            <A
                                href="/bilibili/avatar"
                                class={`block rounded-2xl border bg-white p-4 shadow-sm transition ${
                                    location.pathname === "/bilibili/avatar"
                                        ? "border-sky-500 bg-sky-50 text-sky-700"
                                        : "border-neutral-200 hover:border-neutral-300 hover:bg-neutral-50"
                                }`}
                            >
                                <div class="flex items-center gap-3">
                                    <img
                                        src={accountDisplayFace()}
                                        alt={accountDisplayName()}
                                        referrerPolicy="no-referrer"
                                        onError={() => setAvatarLoadFailed(true)}
                                        class="h-14 w-14 rounded-2xl border border-neutral-200 bg-white object-cover"
                                    />
                                    <div class="min-w-0">
                                        <p class="text-sm text-neutral-500">当前账号</p>
                                        <p class="truncate font-semibold text-neutral-900">{accountDisplayName()}</p>
                                        <p class={`mt-1 text-xs ${accountInfoLoading() ? "text-neutral-500" : "text-green-600"}`}>
                                            {accountStatusText()}
                                        </p>
                                    </div>
                                </div>
                            </A>

                            <nav class="flex flex-col gap-2">
                                <For each={bilibiliMenuItems}>
                                    {(item) => {
                                        const isActive = () => location.pathname === item.path;
                                        return (
                                            <A
                                                href={item.path}
                                                class={`block w-full rounded-2xl border px-4 py-3 text-left transition ${
                                                    isActive()
                                                        ? "border-sky-500 bg-sky-50 text-sky-700 shadow-sm"
                                                        : "border-transparent bg-white text-neutral-700 hover:border-neutral-200 hover:bg-neutral-50"
                                                }`}
                                            >
                                                <p class="text-sm font-semibold">{item.label}</p>
                                                <p class="mt-1 text-xs text-neutral-500">{item.description}</p>
                                            </A>
                                        );
                                    }}
                                </For>
                            </nav>
                        </aside>

                        <section class="flex-1 overflow-y-auto p-6 md:p-8">
                            <div class="max-w-4xl space-y-6">
                                <h1 class="text-3xl font-bold text-neutral-900">{activeMenuItem().label}</h1>
                                <p class="text-sm text-neutral-500">{activeMenuItem().description}</p>
                            
                                {props.children}

                                <Show when={submitError()}>
                                    <div class="rounded-2xl bg-red-50 px-4 py-3 text-sm text-red-600">
                                        {submitError()}
                                    </div>
                                </Show>
                            </div>
                        </section>
                    </div>
                </Match>
            </Switch>
        </div>
    )
}
