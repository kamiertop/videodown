import {invoke} from "@tauri-apps/api/core";
import {createEffect, createMemo, createSignal, JSXElement, onMount, Show} from "solid-js";
import bilibiliAvatarFallback from "../../assets/bilibili_256_256.svg";
import {useBilibiliAccountStore} from "./account.ts";
import type {BilibiliMyInfo} from "./shared.ts";

function formatBirthday(value?: number): string {
    if (!value) {
        return "未公开";
    }

    const date = new Date(value * 1000);
    if (Number.isNaN(date.getTime())) {
        return "未公开";
    }

    return date.toLocaleDateString("zh-CN");
}

export default function BilibiliAvatar(): JSXElement {
    const {
        accountInfo,
        setAccountInfo,
        accountInfoLoading,
        setAccountInfoLoading
    } = useBilibiliAccountStore();
    const [loading, setLoading] = createSignal(!accountInfo()?.profile && accountInfoLoading());
    const [error, setError] = createSignal<string | null>(null);
    const [fallbackInfo, setFallbackInfo] = createSignal<BilibiliMyInfo | null>(null);
    const [avatarLoadFailed, setAvatarLoadFailed] = createSignal(false);

    async function loadAccountInfoFallback() {
        setLoading(true);
        setAccountInfoLoading(true);
        try {
            const response = await invoke<BilibiliMyInfo>("my_info");
            setAccountInfo(response);
            setFallbackInfo(response);
            setAvatarLoadFailed(false);
            if (!response?.profile) {
                setError("出错了, 快提一个issue吧");
            }
        } catch (err) {
            setError(err instanceof Error ? err.message : String(err));
        } finally {
            setAccountInfoLoading(false);
            setLoading(false);
        }
    }

    createEffect(() => {
        if (accountInfo()?.profile) {
            setFallbackInfo(null);
            setError(null);
            setLoading(false);
            setAvatarLoadFailed(false);
            return;
        }

        if (accountInfoLoading()) {
            setLoading(true);
        }
    });

    onMount(() => {
        if (!accountInfo()?.profile && !accountInfoLoading()) {
            void loadAccountInfoFallback();
        }
    });

    const info = createMemo(() => accountInfo() ?? fallbackInfo());
    const profile = createMemo(() => info()?.profile);
    const displayName = createMemo(() => profile()?.name?.trim() || "Bilibili 用户");
    const displayFace = createMemo(() => {
        const embeddedFace = profile()?.face_data_url?.trim();
        const remoteFace = profile()?.face?.trim();
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
    const displaySign = createMemo(() => profile()?.sign?.trim() || "这个人很神秘，什么都没有写。");

    return (
        <div class="rounded-3xl border border-neutral-200 bg-neutral-50 p-6 space-y-6">
            <Show when={loading()}>
                <div class="rounded-2xl border border-neutral-200 bg-white px-4 py-6 text-sm text-neutral-500">
                    正在加载当前账号信息...
                </div>
            </Show>

            <Show when={error()}>
                <div class="rounded-2xl bg-red-50 px-4 py-3 text-sm text-red-600">
                    {error()}
                </div>
            </Show>

            <Show when={!loading() && info()}>
                <div class="space-y-6">
                    <div class="flex flex-col gap-4 md:flex-row md:items-center">
                        <img
                            src={displayFace()}
                            alt={displayName()}
                            referrerPolicy="no-referrer"
                            onError={() => setAvatarLoadFailed(true)}
                            class="h-24 w-24 rounded-full border border-neutral-200 bg-white object-cover shadow-sm"
                        />
                        <div class="space-y-2">
                            <div>
                                <h2 class="text-2xl font-semibold text-neutral-900">{displayName()}</h2>
                                <p class="text-sm text-neutral-500">UID：{profile()?.mid ?? "未获取"}</p>
                            </div>
                            <p class="text-sm text-neutral-600">{displaySign()}</p>
                        </div>
                    </div>

                    <div class="grid grid-cols-1 gap-4 text-sm md:grid-cols-2 xl:grid-cols-4">
                        <div class="rounded-2xl border border-neutral-200 bg-white p-4">
                            <p class="text-neutral-500">等级</p>
                            <p class="mt-2 text-lg font-semibold text-neutral-900">Lv.{profile()?.level ?? "--"}</p>
                        </div>
                        <div class="rounded-2xl border border-neutral-200 bg-white p-4">
                            <p class="text-neutral-500">性别</p>
                            <p class="mt-2 text-lg font-semibold text-neutral-900">{profile()?.sex || "未公开"}</p>
                        </div>
                        <div class="rounded-2xl border border-neutral-200 bg-white p-4">
                            <p class="text-neutral-500">生日</p>
                            <p class="mt-2 text-lg font-semibold text-neutral-900">{formatBirthday(profile()?.birthday)}</p>
                        </div>
                        <div class="rounded-2xl border border-neutral-200 bg-white p-4">
                            <p class="text-neutral-500">硬币</p>
                            <p class="mt-2 text-lg font-semibold text-neutral-900">{info()?.coins ?? "--"}</p>
                        </div>
                    </div>

                    <div class="grid grid-cols-1 gap-4 text-sm md:grid-cols-2">
                        <div class="rounded-2xl border border-neutral-200 bg-white p-4">
                            <p class="text-neutral-500">昵称</p>
                            <p class="mt-2 font-semibold text-neutral-900">{displayName()}</p>
                        </div>
                        <div class="rounded-2xl border border-neutral-200 bg-white p-4">
                            <p class="text-neutral-500">账号状态</p>
                            <p class="mt-2 font-semibold text-green-600">已登录</p>
                        </div>
                    </div>
                </div>
            </Show>
        </div>
    );
}
