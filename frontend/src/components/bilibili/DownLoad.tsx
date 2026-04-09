import {For, type JSXElement} from "solid-js";

interface DownLoadProps {
    author: string;
    cover: string;
    title: string;
    duration: string;
}

export default function DownLoad(props: DownLoadProps[]): JSXElement {
    return (
        <div>
            <For each={props}>
                {item => (
                    <div>
                        <div>{item.author}</div>
                    </div>
                )}
            </For>
        </div>
    )
}
