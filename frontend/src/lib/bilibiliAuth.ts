import {IsLoggedIn} from "../../wailsjs/go/api/BiliBili";

let cachedLoggedIn: boolean | null = null;
let inflight: Promise<boolean> | null = null;

if (typeof window !== "undefined") {
    window.addEventListener("bilibili-auth-changed", () => {
        cachedLoggedIn = null;
        inflight = null;
    });
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
