import {createSignal} from "solid-js";
import type {BilibiliMyInfo} from "./shared.ts";

const [accountInfo, setAccountInfo] = createSignal<BilibiliMyInfo | null>(null);
const [accountInfoLoading, setAccountInfoLoading] = createSignal(false);

export function useBilibiliAccountStore() {
    return {
        accountInfo,
        setAccountInfo,
        accountInfoLoading,
        setAccountInfoLoading
    };
}

