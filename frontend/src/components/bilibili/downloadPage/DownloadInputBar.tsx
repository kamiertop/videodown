import type {JSXElement} from "solid-js";

interface DownloadInputBarProps {
    value: string;
    onInput: (value: string) => void;
    onParse: () => void;
    onClear: () => void;
}

// 手动解析输入区。这里只收集输入和触发事件，具体怎么解析 BV/AV/URL 放在下载页里处理。
export default function DownloadInputBar(props: DownloadInputBarProps): JSXElement {
    return (
        <section>
            <div class="flex flex-row join gap-2">
                <input
                    type="text"
                    placeholder="请输入视频链接, 支持BV号、AV号、视频URL等格式, 可按回车直接解析"
                    value={props.value}
                    onInput={(e) => props.onInput(e.currentTarget.value)}
                    onKeyDown={(e) => {
                        if (e.key === "Enter") {
                            props.onParse();
                        }
                    }}
                    class="input input-success w-full"
                />
                <button class="btn btn-outline btn-secondary" type="button" onClick={props.onParse}>
                    解析
                </button>
                <button class="btn btn-outline btn-info" type="button" onClick={props.onClear}>
                    清空
                </button>
            </div>
        </section>
    );
}
