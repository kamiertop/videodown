import {createEffect, createSignal, type JSXElement} from "solid-js";
import {Favorites, FavoritesList} from "../../../../wailsjs/go/api/BiliBili";
import {model} from "../../../../wailsjs/go/models";
import type {MediaCardItem} from "../../../lib/model.ts";
import {StarIcon} from "../../icons/IconStar";
import {type SidebarListItem} from "../../SidebarList";
import FavoriteCollectionView from "./FavoriteCollectionView";


/**
 * 收藏夹面板
 */

// 收藏夹详情默认起始页
const FAVORITE_PAGE = 1;
// 每页数量：兼顾首屏速度和请求次数
const FAVORITE_PAGE_SIZE = 30;
const EMPTY_FAVORITE_LIST: readonly model.FavoriteItem[] = [];

interface FavSidebarItem extends SidebarListItem {
  raw: model.FavoriteItem;
}

function toFavSidebarItems(list: readonly model.FavoriteItem[]): FavSidebarItem[] {
  return list.map(item => ({
    id: item.id,
    title: item.title,
    count: item.media_count,
    subtitle: '',
    raw: item,
  }));
}

function toMediaCards(medias: model.FavoriteMedias[], listName: string): MediaCardItem[] {
  const name: string = listName.trim();
  return medias.map(media => ({
    id: media.id,
    title: media.title,
    cover: media.cover,
    duration: media.duration,
    bvid: media.bvid,
    link: media.link,
    upperName: media.upper?.name || '未知',
    play: media.cnt_info?.play,
    danmaku: media.cnt_info?.danmaku,
    pubtime: media.pubtime,
    sourceListName: name,
    sourceListKind: '收藏夹',
  }));
}

export default function FavoritePanel(props: {
  active: boolean;
  showToast: (message: string, type?: "success" | "error" | "warning" | "info") => void;
}): JSXElement {
  // 左侧收藏夹列表加载态
  const [loading, setLoading] = createSignal<boolean>(false);
  // 收藏夹列表数据
  const [favorites, setFavorites] = createSignal<model.FavoritesData | null>(null);
  // 当前选中的收藏夹
  const [selectedItem, setSelectedItem] = createSignal<model.FavoriteItem | null>(null);
  // 右侧详情首屏加载态（不包含“加载更多”）
  const [detailLoading, setDetailLoading] = createSignal(false);
  // 右侧详情错误文案
  const [detailError, setDetailError] = createSignal('');
  // 当前收藏夹详情
  const [detail, setDetail] = createSignal<model.FavoriteData | null>(null);
  // 当前页码（用于分页追加）
  const [detailPage, setDetailPage] = createSignal(FAVORITE_PAGE);
  // 是否还有下一页
  const [detailHasMore, setDetailHasMore] = createSignal(false);
  // “加载更多”按钮加载态
  const [loadingMore, setLoadingMore] = createSignal(false);
  // 是否已至少加载过一次（用于懒加载）
  const [loadedOnce, setLoadedOnce] = createSignal(false);
  // 详情版本号：切换收藏夹时递增，驱动子组件重置内部状态
  const [detailVersion, setDetailVersion] = createSignal(0);

  // 请求序号：避免并发请求返回乱序时旧数据覆盖新数据
  let detailRequestSeq = 0;

  const sidebarItems = () => toFavSidebarItems(favorites()?.list ?? EMPTY_FAVORITE_LIST);
  const mediaCards = () => toMediaCards(detail()?.medias ?? [], selectedItem()?.title ?? "");

  const loadFavoriteDetail = async (item: model.FavoriteItem, append = false) => {
    const targetPage = append ? detailPage() + 1 : FAVORITE_PAGE;
    setSelectedItem(item);
    if (append) {
      // 追加分页时只打开“加载更多”态，保留已渲染内容
      setLoadingMore(true);
    } else {
      // 切换收藏夹或首次加载：重置详情区状态
      setDetailVersion(prev => prev + 1);
      setDetailLoading(true);
      setDetailError('');
      setDetail(null);
      setDetailPage(FAVORITE_PAGE);
      setDetailHasMore(false);
    }

    const seq = ++detailRequestSeq;
    try {
      const data = await Favorites(item.id, targetPage, FAVORITE_PAGE_SIZE);
      // 若存在更新请求，本次结果已过期，直接丢弃
      if (seq !== detailRequestSeq) return;

      if (append) {
        // 追加模式：把新一页拼接到末尾
        setDetail(prev => {
          const prevMedias = prev?.medias ?? [];
          return model.FavoriteData.createFrom({...data, medias: [...prevMedias, ...(data.medias ?? [])]});
        });
      } else {
        setDetail(data);
      }
      setDetailPage(targetPage);
      setDetailHasMore(data.has_more);
    } catch (error) {
      // 与成功分支一致：过期请求不应影响当前界面
      if (seq !== detailRequestSeq) return;
      const msg = error instanceof Error ? error.message : String(error);
      if (append) props.showToast(`加载更多失败: ${msg}`, 'warning');
      else {
        setDetail(null);
        setDetailError(msg);
      }
    } finally {
      if (seq === detailRequestSeq) {
        if (append) setLoadingMore(false);
        else setDetailLoading(false);
      }
    }
  };

  const loadFavorites = async () => {
    setLoading(true);
    try {
      const data = await FavoritesList();
      setFavorites(data);
      setLoadedOnce(true);

      // 刷新后优先保留原选中项；不存在时回退到第一项
      const current = selectedItem()?.id;
      const nextItem = data.list?.find(item => item.id === current) ?? data.list?.[0];
      if (nextItem) {
        void loadFavoriteDetail(nextItem);
      } else {
        // 空列表兜底：清空右侧详情和分页状态
        setSelectedItem(null);
        setDetail(null);
        setDetailError('');
        setDetailPage(FAVORITE_PAGE);
        setDetailHasMore(false);
        setDetailVersion(prev => prev + 1);
      }
    } catch (error) {
      props.showToast(error instanceof Error ? error.message : String(error), 'error');
    } finally {
      setLoading(false);
    }
  };

  createEffect(() => {
    // 仅在页签首次激活时加载，避免后台无意义请求
    if (props.active && !loadedOnce()) {
      void loadFavorites();
    }
  });

  return (
    <div class={props.active ? "flex h-full min-h-0 flex-1" : "hidden"}>
      <FavoriteCollectionView
        sidebarItems={sidebarItems}
        selectedSidebarId={() => selectedItem()?.id ?? null}
        onSelectSidebar={(item) => {
          // 已选中且详情可用时，不重复请求同一收藏夹
          if (selectedItem()?.id === item.raw.id && detail() && !detailLoading()) return;
          void loadFavoriteDetail(item.raw);
        }}
        sidebarIcon={StarIcon()}
        sidebarLabel={'收藏夹'}
        sidebarCount={() => favorites()?.count ?? sidebarItems().length}
        sidebarLoading={loading}
        onRefresh={() => void loadFavorites()}
        hasSelection={() => !!selectedItem()}
        detailLoading={detailLoading}
        detailError={detailError}
        onRetryDetail={() => {
          const item = selectedItem();
          if (item) void loadFavoriteDetail(item);
        }}
        detailTitle={() => selectedItem()?.title ?? ''}
        detailMediaCount={() => detail()?.info?.media_count ?? selectedItem()?.media_count ?? 0}
        mediaCards={mediaCards}
        detailVersion={detailVersion}
        hasMore={detailHasMore}
        loadingMore={loadingMore}
        onLoadMore={() => {
          const item = selectedItem();
          // 防重入：未选中、无下一页、正在加载更多时直接返回
          if (!item || !detailHasMore() || loadingMore()) return;
          void loadFavoriteDetail(item, true);
        }}
      />
    </div>
  );
}
