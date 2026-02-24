import { createSignal, JSXElement, onMount } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { invoke } from "@tauri-apps/api/core";

export default function Settings(): JSXElement {
    const navigate = useNavigate();
    const [storage, setStorage] = createSignal<string>("");

    function changeStorage() {
        const newStorage = prompt("请输入新的存储位置");
        if (newStorage) {
            setStorage(newStorage);
        }
    }

    onMount(async () => {
        try {
            const loggedIn = await invoke<boolean>("is_logged_in");
            if (!loggedIn) {
                navigate("/bilibili", { replace: true });
                return;
            }
            const dir = await invoke<string>("get_storage");
            setStorage(dir);
        } catch (error) {
            // TODO
        }
    });
    
    return (
        <div class="px-6 py-4 space-y-6">
            <section class="space-y-2">
                <h2 class="text-lg font-semibold">通用</h2>
                <ul class="space-y-2">
                    <li class="flex items-center justify-between rounded-md border border-neutral-200 px-4 py-3">
                        <span>存储位置</span>
                        <span>{storage()}</span>
                        <button class="text-sm text-blue-600" onClick={changeStorage}>更改</button>
                    </li>
                    <li class="flex items-center justify-between rounded-md border border-neutral-200 px-4 py-3">
                        <span>命名规则</span>
                        <button class="text-sm text-blue-600">编辑</button>
                    </li>
                </ul>
            </section>

            <section class="space-y-2">
                <h2 class="text-lg font-semibold">账户</h2>
                <ul class="space-y-2">
                    <li class="flex items-center justify-between rounded-md border border-neutral-200 px-4 py-3">
                        <span>退出登录</span>
                        <button class="text-sm text-red-600">退出</button>
                    </li>
                </ul>
            </section>
        </div>
    );
}
