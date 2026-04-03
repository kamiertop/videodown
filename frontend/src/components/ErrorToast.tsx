import {Show, type JSXElement} from "solid-js";

interface ErrorToastProps {
    message?: string;
}

export default function ErrorToast(props: ErrorToastProps): JSXElement {
    return (
        <Show when={props.message}>
            <div class="toast toast-top toast-end z-50 top-16 md:top-20">
                <div class="alert alert-error shadow-lg">
                    <span>{props.message}</span>
                </div>
            </div>
        </Show>
    )
}
