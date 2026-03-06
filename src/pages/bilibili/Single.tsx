import {createSignal, JSXElement, Show} from "solid-js";


export default function BilibiliVideo(): JSXElement {
    const [link, setLink] = createSignal("");
    const [submitError, setSubmitError] = createSignal<string | null>(null);
    const [submitMessage, setSubmitMessage] = createSignal<string | null>(null);

    function resetFeedback() {
        setSubmitError(null);
        setSubmitMessage(null);
    }

    function handleDownload(event: SubmitEvent) {
        event.preventDefault();
        setSubmitError(null);

        if (!link().trim()) {
            setSubmitMessage(null);
            setSubmitError("请先输入 B 站视频链接");
            return;
        }

        setSubmitMessage("");
    }

    return (
        <div class="space-y-4">
            <form class="space-y-4" onSubmit={handleDownload}>
                <label class="block space-y-2">

                    <input
                        type="url"
                        value={link()}
                        onInput={(event) => {
                            setLink(event.currentTarget.value);
                            setSubmitError(null);
                        }}
                        placeholder="https://www.bilibili.com/video/BV..."
                        class="w-full rounded-2xl border border-neutral-300 bg-white px-4 py-3 outline-none transition focus:border-sky-500 focus:ring-2 focus:ring-sky-100"
                    />
                </label>

                <div class="flex justify-end gap-3">
                    <button
                        type="button"
                        onClick={() => {
                            setLink("");
                            resetFeedback();
                        }}
                        class="rounded-2xl border border-neutral-300 bg-white px-5 py-3 text-sm font-medium text-neutral-700 transition hover:bg-neutral-50"
                    >
                        清空
                    </button>
                    <button
                        type="submit"
                        class="rounded-2xl bg-sky-600 px-6 py-3 text-sm font-medium text-white transition hover:bg-sky-700"
                    >
                        解析
                    </button>
                </div>

                <Show when={submitError()}>
                    <div class="rounded-2xl bg-red-50 px-4 py-3 text-sm text-red-600">
                        {submitError()}
                    </div>
                </Show>

                <Show when={submitMessage()}>
                    <div class="rounded-2xl bg-green-50 px-4 py-3 text-sm text-green-700">
                        {submitMessage()}
                    </div>
                </Show>
            </form>
        </div>
    );
}
