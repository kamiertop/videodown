import {Route, Router, type RouteSectionProps} from "@solidjs/router";
import {JSXElement} from "solid-js";
import Header from "./components/Header.tsx";
import Settings from "./pages/Settings.tsx";
import About from "./pages/About.tsx";
import Home from "./pages/Home.tsx";
import Bilibili from "./pages/Bilibili.tsx";
import Douyin from "./pages/Douyin.tsx";

function Layout(props: RouteSectionProps): JSXElement {
    return (
        <div class="h-screen flex flex-col">
            <Header/>
            <main class="flex-1">
                {props.children}
            </main>
        </div>
    )
}

export default function App() :JSXElement{
    return (
        <Router root={Layout}>
            <Route path={"/"} component={Home}/>
            <Route path={"/settings"} component={Settings}/>
            <Route path={"/about"} component={About}/>
            <Route path={"/bilibili"} component={Bilibili}/>
            <Route path={"/douyin"} component={Douyin}/>
        </Router>
    )
}