import {Show, type JSXElement} from "solid-js";

interface ToastProps {
    message?: string;
    type?: "error" | "success" | "info" | "warning";
}

const typeClassMap = {
    error: "alert-error",
    success: "alert-success",
    info: "alert-info",
    warning: "alert-warning",
} as const;

export default function Toast(props: ToastProps): JSXElement {
    const alertClass = () => typeClassMap[props.type || "info"];

    return (
        <Show when={props.message}>
            <div class="toast toast-top toast-end z-50 top-16 md:top-20">
                <div class={`alert ${alertClass()} shadow-lg`}>
                    <span>{props.message}</span>
                </div>
            </div>
        </Show>
    )
}
