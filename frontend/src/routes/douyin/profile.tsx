import {createFileRoute} from '@tanstack/solid-router'
import {createSignal, type JSXElement, Match, onMount, Show, Switch} from "solid-js";
import {GetCookie, Profile, SetCookie} from "../../../wailsjs/go/api/Douyin";
import {model} from "../../../wailsjs/go/models";
import DetailError from "../../components/DetailError.tsx";
import DetailLoading from "../../components/DetailLoading.tsx";
import IconRefresh from "../../components/icons/IconRefresh.tsx";
import Toast from "../../components/Toast.tsx";
import {useToast} from "../../hooks/useToast.ts";
import {formatCount} from "../../lib/format.ts";

export const Route = createFileRoute('/douyin/profile')({
  component: RouteComponent,
})

function avatarUrl(user: model.MyInfoResponse | undefined): string {
  return user?.user.avatar_medium?.url_list?.[0] || user?.user.avatar_larger?.url_list?.[0] || user?.user.avatar_thumb?.url_list?.[0] || "";
}

function RouteComponent(): JSXElement {
  const [cookie, setCookie] = createSignal<string>('');
  const [profile, setProfile] = createSignal<model.MyInfoResponse | null>(null);
  const [loadingProfile, setLoadingProfile] = createSignal(false);
  const [profileError, setProfileError] = createSignal("");
  const {message, type, showToast} = useToast();

  async function loadProfile(showSuccess = false): Promise<void> {
    setLoadingProfile(true);
    setProfileError("");
    try {
      const info = await Profile();
      setProfile(info);
      window.dispatchEvent(new CustomEvent("douyin-profile-updated", {
        detail: {
          nickname: info.user?.nickname,
          avatar: avatarUrl(info),
        },
      }));

      if (showSuccess) showToast("账号信息已刷新", "success");
    } catch (error) {
      const msg = error instanceof Error ? error.message : String(error);
      setProfileError(msg);
      showToast(msg, "error");
    } finally {
      setLoadingProfile(false);
    }
  }

  onMount(async () => {
    try {
      const c: string = await GetCookie();
      if (c == "") {
        showToast('Cookie为空，请设置Cookie', 'warning');
      } else {
        setCookie(c);
        await loadProfile();
      }
    } catch (error) {
      const errMsg: string = error as string;
      if (errMsg.includes("未设置")) {
        showToast(errMsg, 'warning');
      } else {
        showToast(errMsg, 'error');
      }
    }
  })

  async function saveCookie(c: string): Promise<void> {
    try {
      setCookie(c)
      await SetCookie(c);
      setProfile(null);
      window.dispatchEvent(new CustomEvent("douyin-profile-updated", {detail: null}));
    } catch (e) {
      showToast("保存Cookie失败：" + e, 'error');
    }
  }

  return (
    <section class="flex h-full min-h-0 flex-col gap-3 overflow-hidden bg-base-200/40 px-2 py-3">
      <main class="min-h-0 flex-1 overflow-auto rounded-lg border border-base-300 bg-base-100 p-4">
        <div class="mx-auto flex max-w-5xl flex-col gap-4">
          <section class="overflow-hidden rounded-lg border border-base-300 bg-base-100">
            <Switch>
              <Match when={loadingProfile()}>
                <div class="h-56">
                  <DetailLoading/>
                </div>
              </Match>
              <Match when={profileError()}>
                {(error) => <DetailError message={error()} onRetry={() => void loadProfile(true)}/>}
              </Match>
              <Match when={profile()}>
                {(account) => (
                  <div>
                    <div class="relative">
                      <div
                        class="flex items-center gap-4 bg-linear-to-r from-primary/10 via-secondary/10 to-base-100 px-5 py-5 pr-16">
                        <div class="h-20 w-20 shrink-0 overflow-hidden rounded-full bg-base-200 ring-4 ring-base-100">
                          <Show when={avatarUrl(account())} fallback={<div class="h-full w-full bg-base-200"/>}>
                            <img
                              src={avatarUrl(account())}
                              alt=""
                              class="h-full w-full object-cover"
                              referrerPolicy="no-referrer"
                            />
                          </Show>
                        </div>
                        <div class="min-w-0 flex-1">
                          <div class="flex flex-wrap items-center gap-2">
                            <h3
                              class="truncate text-xl font-black text-base-content">{account().user.nickname || "抖音用户"}</h3>
                            <Show when={account()}>
                              <span class="badge badge-outline">抖音号 {account().user.unique_id}</span>
                              <span
                                class="badge badge-outline truncate">OwnerSecUid：{profile()?.user.sec_uid}</span>
                            </Show>
                          </div>
                          <p class="mt-2 line-clamp-2 text-sm text-base-content/65">
                            {account().user.signature || "暂无签名"}
                          </p>
                        </div>
                      </div>
                      <button
                        class="btn btn-ghost btn-circle btn-sm absolute right-3 top-3"
                        type="button"
                        title={loadingProfile() ? "刷新中" : "刷新账号信息"}
                        aria-label={loadingProfile() ? "刷新中" : "刷新账号信息"}
                        disabled={loadingProfile() || !cookie().trim()}
                        onClick={() => void loadProfile(true)}
                      >
                        <IconRefresh
                          class={`h-4 w-4 text-base-content/60 transition-transform ${loadingProfile() ? "animate-spin" : ""}`}
                        />
                      </button>
                    </div>
                    <div class="grid grid-cols-3 gap-px bg-base-200 text-center text-sm">
                      <div class="bg-base-100 px-4 py-3">
                        <p class="text-xs text-base-content/50">作品</p>
                        <p class="mt-1 font-bold tabular-nums">{formatCount(account().user.aweme_count)}</p>
                      </div>
                      <div class="bg-base-100 px-4 py-3">
                        <p class="text-xs text-base-content/50">粉丝</p>
                        <p class="mt-1 font-bold tabular-nums">{formatCount(account().user.follower_count)}</p>
                      </div>
                      <div class="bg-base-100 px-4 py-3">
                        <p class="text-xs text-base-content/50">关注</p>
                        <p class="mt-1 font-bold tabular-nums">{formatCount(account().user.following_count)}</p>
                      </div>
                    </div>
                  </div>
                )}
              </Match>
              <Match when={true}>
                <div class="flex h-56 flex-col items-center justify-center gap-2 text-sm text-base-content/55">
                  <p>尚未加载账号信息</p>
                  <button
                    class="btn btn-primary btn-sm"
                    type="button"
                    disabled={!cookie().trim()}
                    onClick={() => void loadProfile(true)}
                  >
                    获取账号信息
                  </button>
                </div>
              </Match>
            </Switch>
          </section>

          <div class="rounded-lg bg-base-200/80 px-4 py-3 text-xl font-bold">
            复制浏览器完整 Cookie，至少包含 sessionid、ttwid、UIFID、s_v_web_id
          </div>
          <textarea
            class="textarea min-h-72 w-full border-base-300 bg-base-100 font-mono text-xs leading-6"
            value={cookie()}
            onInput={(e) => void saveCookie(e.currentTarget.value)}
            placeholder="粘贴抖音 Cookie 至此"
          />
        </div>
      </main>
      <Toast message={message()} type={type()}></Toast>
    </section>
  )
}
