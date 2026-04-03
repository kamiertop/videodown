import {Show, type JSXElement} from "solid-js";

interface StatusToastProps {
    message?: string;
    tone?: "info" | "success" | "warning";
}

const toneClassMap = {
    info: "alert-info",
    success: "alert-success",
    warning: "alert-warning",
} as const;

export default function StatusToast(props: StatusToastProps): JSXElement {
    const toneClass = () => toneClassMap[props.tone || "info"];

    return (
        <Show when={props.message}>
            <div class="toast toast-top toast-end z-40 top-16 md:top-20">
                <div class={`alert ${toneClass()} shadow-lg`}>
                    <span>{props.message}</span>
                </div>
            </div>
        </Show>
    )
}
