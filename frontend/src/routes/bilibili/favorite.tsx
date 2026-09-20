import {createFileRoute} from '@tanstack/solid-router'
import {createSignal, type JSXElement, onCleanup, onMount} from "solid-js";
import CollectionPanel from "../../components/bilibili/favorite/CollectionPanel";
import FavoritePanel from "../../components/bilibili/favorite/FavoritePanel";
import Toast from "../../components/Toast";
import UnderlineTabs, {type UnderlineTabItem} from "../../components/UnderlineTabs.tsx";
import {useToast} from "../../hooks/useToast";

export const Route = createFileRoute('/bilibili/favorite')({
  component: Favorite,
})

type FavoriteTab = 'favorite' | 'collection';

const FAVORITE_TABS: readonly UnderlineTabItem<FavoriteTab>[] = [
  {key: 'favorite', label: '收藏夹'},
  {key: 'collection', label: '合集'},
];

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
      <div class="flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-base-300 bg-base-100">
        <UnderlineTabs
            tabs={FAVORITE_TABS}
            active={activeTab()}
            onChange={setActiveTab}
            activeClass="border-b-2 border-success text-success"
        />
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
