import {createSignal, onCleanup} from "solid-js";

interface UseToastReturn {
    message: () => string;
    type: () => "error" | "success" | "info" | "warning";
    showToast: (message: string, type?: "error" | "success" | "info" | "warning") => void;
}

/**
 * Toast 提示 Hook
 * @param duration - 自动消失时间(毫秒),默认 3000ms
 */
export function useToast(duration = 2000): UseToastReturn {
    const [message, setMessage] = createSignal<string>('');
    const [type, setType] = createSignal<"error" | "success" | "info" | "warning">('info');

    let timer: number | undefined;

    const showToast = (msg: string, toastType: "error" | "success" | "info" | "warning" = "info") => {
        setMessage(msg);
        setType(toastType);

        if (timer !== undefined) {
            window.clearTimeout(timer);
        }

        timer = window.setTimeout(() => {
            setMessage('');
            timer = undefined;
        }, duration);
    };

    // 组件卸载时清理定时器
    onCleanup(() => {
        if (timer !== undefined) {
            window.clearTimeout(timer);
        }
    });

    return {
        message,
        type,
        showToast,
    };
}
