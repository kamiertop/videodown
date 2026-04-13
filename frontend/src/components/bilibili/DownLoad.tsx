import {createSignal, For, type JSXElement} from "solid-js";


interface VideoListProps {
    author: string;
    cover: string;
    title: string;
    duration: string;
}

const [videos, setVideos] = createSignal<VideoListProps[]>([]);


export function addVideo(selectedVideos: VideoListProps[]): void {
    setVideos([...videos(), ...selectedVideos]);
}

export default function DownLoad(): JSXElement {
    return (
        <div>
            <For each={videos()}>
                {
                    item => <Card {...item}/>
                }
            </For>
        </div>
    )
}

function Card(props: VideoListProps): JSXElement {
    return (
        <div class={"flex flex-row card"}>
            <div>
                <img src={props.cover} alt={props.title}/>
            </div>
            <div>
                
            </div>
        </div>
    )
}
