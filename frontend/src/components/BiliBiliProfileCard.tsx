import {createSignal, Show, type JSXElement} from "solid-js";
import type {model} from "../../wailsjs/go/models";
import bilibiliAvatarFallback from "../assets/bilibili_256_256.svg";

interface BiliBiliProfileCardProps {
    loading?: boolean;
    loggingOut?: boolean;
    onLogout?: () => void | Promise<void>;
    profile: model.MyInfoProfile | null;
}

export default function BiliBiliProfileCard(props: BiliBiliProfileCardProps): JSXElement {
    const [avatarLoadFailed, setAvatarLoadFailed] = createSignal(false);

    const avatar = () => {
        if (avatarLoadFailed() || !props.profile?.face?.trim()) {
            return bilibiliAvatarFallback;
        }
        return props.profile.face;
    }

    return (
        <div class="mx-auto flex h-full max-w-4xl items-center justify-center">
            <div class="card w-full max-w-2xl border border-base-300 bg-base-100 shadow-xl">
                <div class="card-body gap-5">
                    <div class="flex items-center gap-4">
                        <img
                            src={avatar()}
                            alt={props.profile?.name || 'Bilibili 用户头像'}
                            referrerPolicy="no-referrer"
                            onError={() => setAvatarLoadFailed(true)}
                            class="h-20 w-20 rounded-3xl border border-base-300 bg-base-200 object-cover"
                        />
                        <div class="min-w-0 space-y-1">
                            <p class="text-sm text-base-content/60">当前登录账号</p>
                            <h1 class="truncate text-2xl font-black text-base-content">
                                {props.profile?.name || 'Bilibili 用户'}
                            </h1>
                        </div>
                    </div>

                    <Show when={props.loading}>
                        <div class="flex items-center gap-2 text-sm text-base-content/60">
                            <span class="loading loading-spinner loading-sm text-primary"></span>
                            正在加载个人信息...
                        </div>
                    </Show>

                    <Show when={!props.loading && props.profile}>
                        <div class="grid gap-3 sm:grid-cols-3">
                            <div class="rounded-2xl bg-base-200/70 px-4 py-3">
                                <p class="text-xs uppercase tracking-wide text-base-content/50">签名</p>
                                <p class="mt-1 text-sm leading-6 text-base-content/80">
                                    {props.profile?.sign?.trim() || '这个账号还没有填写个性签名。'}
                                </p>
                            </div>
                            <div class="rounded-2xl bg-base-200/70 px-4 py-3">
                                <p class="text-xs uppercase tracking-wide text-base-content/50">会员状态</p>
                                <p class="mt-1 text-sm text-base-content/80">
                                    {props.profile?.vip?.status ? '大会员已开通' : '普通用户'}
                                </p>
                            </div>
                            <div class="rounded-2xl bg-base-200/70 px-4 py-3">
                                <p class="text-xs uppercase tracking-wide text-base-content/50">UID</p>
                                <p class="mt-1 text-sm text-base-content/80">
                                    {props.profile?.mid || '-'}
                                </p>
                            </div>
                        </div>

                        <div class="flex justify-end">
                            <button
                                type="button"
                                class="btn btn-outline btn-error"
                                disabled={props.loggingOut}
                                onClick={() => void props.onLogout?.()}
                            >
                                <Show
                                    when={props.loggingOut}
                                    fallback="退出登录"
                                >
                                    <span class="loading loading-spinner loading-xs"></span>
                                    正在退出...
                                </Show>
                            </button>
                        </div>
                    </Show>
                </div>
            </div>
        </div>
    )
}
