import {createFileRoute} from '@tanstack/solid-router'
import {createSignal, type JSXElement, onCleanup, onMount} from "solid-js";
import {useToast} from "../../hooks/useToast";
import Toast from "../../components/Toast";
import CollectionPanel from "../../components/bilibili/favorite/CollectionPanel";
import FavoritePanel from "../../components/bilibili/favorite/FavoritePanel";

export const Route = createFileRoute('/bilibili/favorite')({
    component: Favorite,
})

function Favorite(): JSXElement {
    const [activeTab, setActiveTab] = createSignal<'favorite' | 'collection'>('favorite');
    const {message, type, showToast} = useToast();

    // 方向键切换标签页：收藏夹 ←，合集 →
    function switchTab(event: KeyboardEvent): void {
        if (event.key === 'ArrowLeft') {
            event.preventDefault();
            setActiveTab('favorite');
            return;
        }

        if (event.key === 'ArrowRight') {
            event.preventDefault();
            setActiveTab('collection');
        }
    }

    onMount(() => {
        window.addEventListener("keydown", switchTab);
    })
    onCleanup(() => {
        window.removeEventListener("keydown", switchTab)
    })

    return (
        <section class="flex h-full min-h-0 flex-col gap-3 overflow-hidden bg-base-200/40 p-3">
            <div class="flex shrink-0 rounded-xl border border-base-300 bg-base-100 p-1">
                <button
                    class={`flex-1 rounded-lg py-2.5 text-center text-sm font-bold transition-colors ${
                        activeTab() === 'favorite'
                            ? 'bg-success/10 text-success'
                            : 'text-base-content/60 hover:text-base-content'
                    }`}
                    onClick={() => setActiveTab('favorite')}
                >
                    收藏夹
                </button>
                <button
                    class={`flex-1 rounded-lg py-2.5 text-center text-sm font-bold transition-colors ${
                        activeTab() === 'collection'
                            ? 'bg-success/10 text-success'
                            : 'text-base-content/60 hover:text-base-content'
                    }`}
                    onClick={() => setActiveTab('collection')}
                >
                    合集
                </button>
            </div>
            <div class="min-h-0 flex-1">
                <FavoritePanel
                    active={activeTab() === 'favorite'}
                    showToast={showToast}
                />
                <CollectionPanel
                    active={activeTab() === 'collection'}
                    showToast={showToast}
                />
            </div>

            <Toast message={message()} type={type()}/>
        </section>
    );
}
