import {createFileRoute} from '@tanstack/solid-router'
import {createSignal, Match, onCleanup, onMount, Switch, type JSXElement} from "solid-js";
import {IsLoggedIn, LogOut, MyInfo, PollQRCode} from "../../../wailsjs/go/api/BiliBili";
import type {model} from "../../../wailsjs/go/models";
import BiliBiliQRCode from "../../components/BiliBiliQRCode";
import ErrorToast from "../../components/ErrorToast";
import BiliBiliProfileCard from "../../components/BiliBiliProfileCard";
import StatusToast from "../../components/StatusToast";

export const Route = createFileRoute('/bilibili/profile')({
    component: Profile,
})

function Profile(): JSXElement {
    const [checkingLogin, setCheckingLogin] = createSignal(true);
    const [loggedIn, setLoggedIn] = createSignal(false);
    const [profileLoading, setProfileLoading] = createSignal(false);
    const [logoutLoading, setLogoutLoading] = createSignal(false);
    const [profile, setProfile] = createSignal<model.MyInfoProfile | null>(null);
    const [statusText, setStatusText] = createSignal('');
    const [statusTone, setStatusTone] = createSignal<"info" | "success" | "warning">("info");
    const [errorText, setErrorText] = createSignal('');
    const [qrExpired, setQRExpired] = createSignal(false);

    let pollTimer: number | undefined;
    let errorToastTimer: number | undefined;
    let statusToastTimer: number | undefined;

    const emitAuthChanged = (loggedIn: boolean) => {
        window.dispatchEvent(new CustomEvent('bilibili-auth-changed', {
            detail: {loggedIn},
        }));
    }

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

    const showStatusToast = (message: string, tone: "info" | "success" | "warning" = "info") => {
        setStatusText(message);
        setStatusTone(tone);
        if (statusToastTimer !== undefined) {
            window.clearTimeout(statusToastTimer);
        }
        statusToastTimer = window.setTimeout(() => {
            setStatusText('');
            statusToastTimer = undefined;
        }, 2500)
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

    const loadProfile = async () => {
        setProfileLoading(true);
        try {
            const result = await MyInfo();
            setProfile(result);
        } catch (error) {
            showErrorToast(error instanceof Error ? error.message : String(error));
        } finally {
            setProfileLoading(false);
        }
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
            setProfile(null);
            setQRExpired(false);
            showStatusToast('已退出登录，请重新扫码登录。', 'info');
            emitAuthChanged(false);
        } catch (error) {
            showErrorToast(error instanceof Error ? error.message : String(error));
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
                showStatusToast('扫码已确认，登录成功。', 'success');
                await loadProfile();
                emitAuthChanged(true);
                return;
            }

            if (result.code === 86090) {
                showStatusToast('已扫码，请在 B 站 App 中确认登录。', 'info');
                schedulePoll(qrcodeKey);
                return;
            }

            if (result.code === 86038) {
                stopPolling();
                setQRExpired(true);
                showStatusToast('二维码已失效，请刷新后重试。', 'warning');
                return;
            }

            if (result.code === 86101) {
                showStatusToast('二维码已就绪，请使用 B 站 App 扫码。', 'info');
                schedulePoll(qrcodeKey);
                return;
            }

            showStatusToast(result.message || `登录状态更新中（${result.code}）...`, 'info');

            schedulePoll(qrcodeKey);
        } catch (error) {
            stopPolling();
            showErrorToast(error instanceof Error ? error.message : String(error));
        }
    }

    onMount(async () => {
        try {
            const isLoggedIn = await IsLoggedIn();
            setLoggedIn(isLoggedIn);
            if (isLoggedIn) {
                await loadProfile();
            }
        } catch (error) {
            showErrorToast(error instanceof Error ? error.message : String(error));
        } finally {
            setCheckingLogin(false);
        }
    })

    onCleanup(() => {
        stopPolling();
        if (errorToastTimer !== undefined) {
            window.clearTimeout(errorToastTimer);
        }
        if (statusToastTimer !== undefined) {
            window.clearTimeout(statusToastTimer);
        }
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
                    <BiliBiliProfileCard
                        loading={profileLoading()}
                        loggingOut={logoutLoading()}
                        onLogout={handleLogout}
                        profile={profile()}
                    />
                </Match>

                <Match when={!loggedIn()}>
                    <div class="mx-auto grid h-full max-w-5xl gap-5 lg:grid-cols-[0.95fr_1.05fr] lg:items-center">
                        <div class="space-y-3">
                            <h1 class="text-3xl font-black leading-tight text-base-content">先完成扫码登录，再进入下载流程</h1>
                        </div>

                        <div class="space-y-3">
                            <BiliBiliQRCode
                                expired={qrExpired()}
                                onLoad={(data: model.QRCodeData) => {
                                    showStatusToast('二维码已就绪，请使用 B 站 App 扫码。', 'info');
                                    setQRExpired(false);
                                    schedulePoll(data.qrcode_key);
                                }}
                            />
                        </div>
                    </div>
                </Match>
            </Switch>
            <StatusToast message={statusText()} tone={statusTone()}/>
            <ErrorToast message={errorText()}/>
        </section>
    )
}
