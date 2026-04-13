import {createEffect, createSignal, type JSXElement} from "solid-js";
import {Favorites, FavoritesList} from "../../../../wailsjs/go/api/BiliBili";
import {model} from "../../../../wailsjs/go/models";
import {StarIcon} from "../../icons/IconStar";
import FavoriteCollectionView from "./FavoriteCollectionView";
import {type MediaCardItem} from "../../VideoCardGrid";
import {type SidebarListItem} from "../../SidebarList";

const FAVORITE_PAGE = 1;
const FAVORITE_PAGE_SIZE = 40;
const EMPTY_FAVORITE_LIST: readonly model.FavoriteItem[] = [];

interface FavSidebarItem extends SidebarListItem {
    raw: model.FavoriteItem;
}

function toFavSidebarItems(list: readonly model.FavoriteItem[]): FavSidebarItem[] {
    return list.map(item => ({id: item.id, title: item.title, count: item.media_count, raw: item}));
}

function toMediaCards(medias: model.FavoriteMedias[]): MediaCardItem[] {
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
    }));
}

export default function FavoritePanel(props: {
    active: boolean;
    onCountChange: (count: number) => void;
    onDownloadMediaList: (medias: MediaCardItem[], label: string) => Promise<void>;
    showToast: (message: string, type?: "success" | "error" | "warning" | "info") => void;
}): JSXElement {
    const [loading, setLoading] = createSignal(false);
    const [favorites, setFavorites] = createSignal<model.FavoritesData | null>(null);
    const [selectedItem, setSelectedItem] = createSignal<model.FavoriteItem | null>(null);
    const [detailLoading, setDetailLoading] = createSignal(false);
    const [detailError, setDetailError] = createSignal('');
    const [detail, setDetail] = createSignal<model.FavoriteData | null>(null);
    const [detailPage, setDetailPage] = createSignal(FAVORITE_PAGE);
    const [detailHasMore, setDetailHasMore] = createSignal(false);
    const [loadingMore, setLoadingMore] = createSignal(false);
    const [loadedOnce, setLoadedOnce] = createSignal(false);
    const [detailVersion, setDetailVersion] = createSignal(0);

    let detailRequestSeq = 0;

    const sidebarItems = () => toFavSidebarItems(favorites()?.list ?? EMPTY_FAVORITE_LIST);
    const mediaCards = () => toMediaCards(detail()?.medias ?? []);

    const loadFavoriteDetail = async (item: model.FavoriteItem, append = false) => {
        const targetPage = append ? detailPage() + 1 : FAVORITE_PAGE;
        setSelectedItem(item);
        if (append) {
            setLoadingMore(true);
        } else {
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
            if (seq !== detailRequestSeq) return;

            if (append) {
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
            props.onCountChange(data.count ?? 0);

            const current = selectedItem()?.id;
            const nextItem = data.list?.find(item => item.id === current) ?? data.list?.[0];
            if (nextItem) {
                void loadFavoriteDetail(nextItem);
            } else {
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
        if (props.active && !loadedOnce()) {
            void loadFavorites();
        }
    });

    return (
        <div class={props.active ? "flex h-full min-h-0 flex-1" : "hidden"}>
            <FavoriteCollectionView
                emptySidebarTitle="暂无收藏夹"
                emptyContentTitle="选择一个收藏夹查看内容"
                sidebarItems={sidebarItems}
                selectedSidebarId={() => selectedItem()?.id ?? null}
                onSelectSidebar={(item) => {
                    if (selectedItem()?.id === item.raw.id && detail() && !detailLoading()) return;
                    void loadFavoriteDetail(item.raw);
                }}
                sidebarIcon={StarIcon}
                sidebarLabel={() => '收藏夹'}
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
                    if (!item || !detailHasMore() || loadingMore()) return;
                    void loadFavoriteDetail(item, true);
                }}
                onDownloadMediaList={props.onDownloadMediaList}
            />
        </div>
    );
}
