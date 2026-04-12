import {createSignal, Show, type JSXElement} from "solid-js";
import type {model} from "../../../wailsjs/go/models.ts";
import bilibiliAvatarFallback from "../../assets/bilibili_256_256.svg";

interface BiliBiliProfileCardProps {
    loading?: boolean;
    loggingOut?: boolean;
    onLogout?: () => void | Promise<void>;
    profile: model.MyInfoProfile | undefined;
}

export default function ProfileCard(props: BiliBiliProfileCardProps): JSXElement {
    const [avatarLoadFailed, setAvatarLoadFailed] = createSignal(false);

    function isVip(): boolean {
        // 会员有效，并且会员类型大于0（0是无会员，1是月度会员，2是年度及以上
        return (props.profile?.vip?.status ?? 0) === 1 && (props.profile?.vip?.type ?? 0) > 0;
    }

    function vipLabel(): string {
        if (!isVip()) {
            return '普通用户';
        }

        const role: number = props.profile?.vip?.role ?? 0;

        if (role >= 15) return '百年大会员';
        if (role >= 7) return '十年大会员';
        if (role >= 3) return '年度大会员';

        return '月度大会员';
    }

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
                                <p class="mt-1 text-sm font-medium text-base-content/80">{vipLabel()}</p>
                                <p class="mt-1 text-xs text-base-content/60">
                                    {isVip()
                                        ? '可解锁更高下载清晰度（具体以稿件与账号权限为准）'
                                        : '当前账号可选清晰度可能受限，开通大会员可解锁更高画质'}
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
