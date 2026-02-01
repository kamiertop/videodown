/* @refresh reload */
import "./index.css"
import { render } from "solid-js/web";
import App from "./App";


const root: HTMLElement | null = document.getElementById("root")

if (!root) {
    throw new Error("root not found")
}

render(App, root);
