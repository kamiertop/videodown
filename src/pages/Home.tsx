import {JSXElement} from "solid-js";
import {A} from "@solidjs/router";
import bilibili from "../assets/bilibili_256_256.svg"
import douyin from "../assets/douyin_256_256.svg"

export default function Home(): JSXElement {
    return (
        <div class="flex flex-row items-center justify-around h-full px-4">
            <A href={"/bilibili/video"}>
                <img src={bilibili} alt={"bilibili"}/>
            </A>
            <A href={"/douyin"}>
                <img src={douyin} alt={"douyin"}/>
            </A>
        </div>
    );
}