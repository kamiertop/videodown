import {createFileRoute, useNavigate} from '@tanstack/solid-router'
import {createSignal, type JSXElement, onCleanup, onMount} from "solid-js";
import {useToast} from "../../hooks/useToast";
import Toast from "../../components/Toast";
import CollectionPanel from "../../components/bilibili/favorite/CollectionPanel";
import FavoritePanel from "../../components/bilibili/favorite/FavoritePanel";
import {type MediaCardItem} from "../../components/VideoCardGrid";

export const Route = createFileRoute('/bilibili/favorite')({
    component: Favorite,
})

type SidebarTab = 'favorite' | 'collection';

function Favorite(): JSXElement {
    const navigate = useNavigate();
    const [activeTab, setActiveTab] = createSignal<SidebarTab>('favorite');
    const [favoriteCount, setFavoriteCount] = createSignal(0);
    const [collectionCount, setCollectionCount] = createSignal(0);
    const {message, type, showToast} = useToast();

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
    const downloadMediaList = async (medias: MediaCardItem[], label: string) => {
        const bvids = [...new Set(medias.map(media => media.bvid?.trim()).filter(Boolean))];
        if (bvids.length === 0) {
            showToast('没有可下载的视频（缺少 BV 号）', 'warning');
            return;
        }
        await navigate({
            to: '/bilibili/download',
            search: {bvids: bvids.join(',')},
        });
        showToast(`${label}，已打开下载页`, 'success');
    };

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
                    <span class="ml-1 text-xs font-normal tabular-nums">{favoriteCount()}</span>
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
                    <span class="ml-1 text-xs font-normal tabular-nums">{collectionCount()}</span>
                </button>
            </div>
            <div class="min-h-0 flex-1">
                <FavoritePanel
                    active={activeTab() === 'favorite'}
                    onCountChange={setFavoriteCount}
                    onDownloadMediaList={downloadMediaList}
                    showToast={showToast}
                />
                <CollectionPanel
                    active={activeTab() === 'collection'}
                    onCountChange={setCollectionCount}
                    onDownloadMediaList={downloadMediaList}
                    showToast={showToast}
                />
            </div>

            <Toast message={message()} type={type()}/>
        </section>
    );
}
