import {JSXElement, onMount} from "solid-js";

import BilibiliQrcode from "../components/BilibiliQrcode.tsx";
import {invoke} from "@tauri-apps/api/core";


export default function Bilibili(): JSXElement {
    onMount(async () => {
        const loggedIn = await invoke<boolean>("is_logged_in");
        if (!loggedIn) {  // 如果未登录，则显示二维码登录
            return <BilibiliQrcode/>
        }
    })

    return (
        <></>
    )
}
