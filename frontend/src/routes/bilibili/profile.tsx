import {createFileRoute} from '@tanstack/solid-router'
import {createResource, createSignal, Match, onCleanup, onMount, Switch, type JSXElement} from "solid-js";
import {LogOut, MyInfo, PollQRCode} from "../../../wailsjs/go/api/BiliBili";
import {getLoggedInDeduped} from "../../lib/bilibiliAuth";
import type {model} from "../../../wailsjs/go/models";
import QRCode from "../../components/bilibili/QRCode.tsx";
import Toast from "../../components/Toast";
import ProfileCard from "../../components/bilibili/ProfileCard.tsx";
import {useToast} from "../../hooks/useToast";

export const Route = createFileRoute('/bilibili/profile')({
    component: Profile,
})

function Profile(): JSXElement {
    const [checkingLogin, setCheckingLogin] = createSignal(true);
    const [loggedIn, setLoggedIn] = createSignal(false);
    const [logoutLoading, setLogoutLoading] = createSignal(false);
    const {message, type, showToast} = useToast();
    const [qrExpired, setQRExpired] = createSignal(false);
    const [profile, {mutate: mutateProfile}] = createResource(
        () => loggedIn(),
        async (isLoggedIn) => {
            if (!isLoggedIn) {
                return undefined;
            }

            try {
                return await MyInfo();
            } catch (error) {
                showToast(error instanceof Error ? error.message : String(error), 'error');
                return undefined;
            }
        },
    );

    let pollTimer: number | undefined;

    const emitAuthChanged = (loggedIn: boolean) => {
        window.dispatchEvent(new CustomEvent('bilibili-auth-changed', {
            detail: {loggedIn},
        }));
    }

    const stopPolling = () => {
        if (pollTimer !== undefined) {
            window.clearTimeout(pollTimer);
            pollTimer = undefined;
        }
    }

    const schedulePoll = (qrcodeKey: string) => {
        stopPolling();
        pollTimer = window.setTimeout(() => {
            void pollQRCodeStatus(qrcodeKey)
        }, 1200)
    }

    const handleLogout = async () => {
        if (logoutLoading()) {
            return;
        }

        setLogoutLoading(true);
        try {
            await LogOut();
            stopPolling();
            setLoggedIn(false);
            mutateProfile(undefined);
            setQRExpired(false);
            showToast('已退出登录，请重新扫码登录。', 'info');
            emitAuthChanged(false);
        } catch (error) {
            showToast(error instanceof Error ? error.message : String(error), 'error');
        } finally {
            setLogoutLoading(false);
        }
    }

    const pollQRCodeStatus = async (qrcodeKey: string) => {
        try {
            const result = await PollQRCode(qrcodeKey);

            if (result.code === 0) {
                stopPolling();
                setLoggedIn(true);
                setQRExpired(false);
                showToast('扫码已确认，登录成功。', 'success');
                emitAuthChanged(true);
                return;
            }

            if (result.code === 86090) {
                showToast('已扫码，请在 B 站 App 中确认登录。', 'info');
                schedulePoll(qrcodeKey);
                return;
            }

            if (result.code === 86038) {
                stopPolling();
                setQRExpired(true);
                showToast('二维码已失效，请刷新后重试。', 'warning');
                return;
            }

            if (result.code === 86101) {
                showToast('二维码已就绪，请使用 B 站 App 扫码。', 'info');
                schedulePoll(qrcodeKey);
                return;
            }

            showToast(result.message || `登录状态更新中（${result.code}）...`, 'info');

            schedulePoll(qrcodeKey);
        } catch (error) {
            stopPolling();
            showToast(error instanceof Error ? error.message : String(error), 'error');
        }
    }

    onMount(async () => {
        try {
            const isLoggedIn = await getLoggedInDeduped();
            setLoggedIn(isLoggedIn);
        } catch (error) {
            showToast(error instanceof Error ? error.message : String(error), 'error');
        } finally {
            setCheckingLogin(false);
        }
    })

    onCleanup(() => {
        stopPolling();
    })

    return (
        <section class="h-full overflow-hidden bg-base-200/40 px-4 py-4 md:px-6 md:py-5">
            <Switch>
                <Match when={checkingLogin()}>
                    <div class="flex h-full items-center justify-center">
                        <span class="loading loading-spinner loading-lg text-primary"></span>
                    </div>
                </Match>

                <Match when={loggedIn()}>
                    <ProfileCard
                        loading={profile.loading}
                        loggingOut={logoutLoading()}
                        onLogout={handleLogout}
                        profile={profile()}
                    />
                </Match>

                <Match when={!loggedIn()}>
                    <div class="flex h-full flex-col items-center justify-center gap-8">
                        <div class="text-center space-y-3">
                            <h1 class="text-3xl font-black leading-tight text-base-content">先完成扫码登录，再进入下载流程</h1>
                            <p class="text-sm text-base-content/60">使用 B 站 App 扫描二维码即可登录</p>
                        </div>

                        <QRCode
                            expired={qrExpired()}
                            onLoad={(data: model.QRCodeData) => {
                                showToast('二维码已就绪，请使用 B 站 App 扫码。', 'info');
                                setQRExpired(false);
                                schedulePoll(data.qrcode_key);
                            }}
                        />
                    </div>
                </Match>
            </Switch>
            <Toast message={message()} type={type()}/>
        </section>
    )
}
