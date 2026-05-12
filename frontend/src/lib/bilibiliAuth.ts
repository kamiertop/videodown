import {IsLoggedIn, MyInfo, RefreshCookie} from "../../wailsjs/go/api/BiliBili";
import type {model} from "../../wailsjs/go/models";

let cachedLoggedIn: boolean | null = null;
let inflight: Promise<boolean> | null = null;
let cachedMyInfo: model.MyInfoProfile | null = null;
let myInfoInflight: Promise<model.MyInfoProfile> | null = null;
let refreshCookieInflight: Promise<model.CookieRefreshData> | null = null;
let refreshCookieChecked = false;

function clearBilibiliAuthCache(): void {
    cachedLoggedIn = null;
    inflight = null;
    cachedMyInfo = null;
    myInfoInflight = null;
    refreshCookieInflight = null;
    refreshCookieChecked = false;
}

function handleBilibiliAuthChanged(event: Event): void {
    const detail = (event as CustomEvent<{ loggedIn?: boolean }>).detail;
    if (detail?.loggedIn === false) {
        clearBilibiliAuthCache();
        return;
    }

    cachedLoggedIn = detail?.loggedIn ?? null;
    inflight = null;
    cachedMyInfo = null;
    refreshCookieChecked = false;
}

if (typeof window !== "undefined") {
    window.addEventListener("bilibili-auth-changed", handleBilibiliAuthChanged);
}

/** 合并短时间内的多次 IsLoggedIn 调用为一次后端请求（例如侧栏与账号页同时挂载）。 */
export async function getLoggedInDeduped(): Promise<boolean> {
    if (cachedLoggedIn !== null) {
        return cachedLoggedIn;
    }
    if (inflight) {
        return inflight;
    }
    inflight = IsLoggedIn().then((v) => {
        cachedLoggedIn = v;
        inflight = null;
        return v;
    });
    return inflight;
}

/** 合并侧栏头像与账号页同时触发的 MyInfo 调用，避免重复请求 B 站 myinfo 接口。 */
export async function getMyInfoDeduped(): Promise<model.MyInfoProfile> {
    if (cachedMyInfo) {
        return cachedMyInfo;
    }
    if (myInfoInflight) {
        return myInfoInflight;
    }
    myInfoInflight = MyInfo().then((profile) => {
        cachedMyInfo = profile;
        myInfoInflight = null;
        return profile;
    }).catch((error: unknown) => {
        myInfoInflight = null;
        throw error;
    });
    return myInfoInflight;
}

/** 进入 B 站功能页时检查并刷新 Cookie；每次会话只检查一次，并合并并发调用。 */
export async function refreshBilibiliCookieDeduped(): Promise<model.CookieRefreshData | undefined> {
    if (refreshCookieChecked) {
        return undefined;
    }
    if (refreshCookieInflight) {
        return refreshCookieInflight;
    }

    const loggedIn = await getLoggedInDeduped();
    if (!loggedIn) {
        refreshCookieChecked = true;
        return undefined;
    }

    refreshCookieChecked = true;
    refreshCookieInflight = RefreshCookie().then((data) => {
        refreshCookieInflight = null;
        if (data.refresh) {
            cachedMyInfo = null;
        }
        return data;
    }).catch((error: unknown) => {
        refreshCookieInflight = null;
        refreshCookieChecked = true;
        throw error;
    });
    return refreshCookieInflight;
}
