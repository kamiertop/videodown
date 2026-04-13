import {createFileRoute, Link, useNavigate} from '@tanstack/solid-router'
import {type JSXElement, onMount} from "solid-js";
import UpCommonLayout from "../../../components/bilibili/up/UpCommonLayout.tsx";
import UpDetailHeaderRight from "../../../components/bilibili/up/UpDetailHeaderRight.tsx";
import UpDetailBody from "../../../components/bilibili/up/UpDetailBody.tsx";
import Toast from "../../../components/Toast";
import {useToast} from "../../../hooks/useToast";
import {createUpDetailLogic} from "./-upDetail.logic.ts";
import IconChevronLeft from "../../../components/icons/IconChevronLeft.tsx";

export const Route = createFileRoute('/bilibili/up/$mid')({
    component: UpDetail,
})

function UpDetail(): JSXElement {
    const params = Route.useParams();
    const navigate = useNavigate();
    // 说明：路由组件仅负责“装配”页面，不在这里堆积业务逻辑。
    // - 数据请求、状态管理、竞态处理等都在 `-upDetail.logic.ts` 中
    // - 这里把 signals/actions 透传给纯 UI 组件 `UpDetailBody`
    const {message, type, showToast} = useToast();
    const logic = createUpDetailLogic(() => params().mid, navigate, showToast);

    onMount(() => {
        // 首次进入页面时触发初始化加载（Info + 默认视频列表）。
        logic.init();
    })

    return (
        <UpCommonLayout
            headerLeft={
                <>
                    <Link
                        to="/bilibili/up"
                        class="btn btn-ghost btn-sm gap-1"
                    >
                        <IconChevronLeft class="h-4 w-4"/>
                        返回
                    </Link>
                    <div class="h-5 w-px bg-base-300"></div>
                    <h2 class="text-sm font-bold text-base-content">UP主详情</h2>
                    <span class="rounded-full bg-base-200 px-2 py-0.5 text-xs tabular-nums text-base-content/60">
                        mid: {params().mid}
                    </span>
                </>
            }
            headerRight={
                <UpDetailHeaderRight loading={logic.loading()} info={logic.info()}/>
            }
        >
            <UpDetailBody
                activeTab={logic.activeTab}
                setActiveTab={logic.setActiveTab}
                videoLoading={logic.videoLoading}
                videoError={logic.videoError}
                videoCards={logic.videoCards}
                videoTotal={logic.videoTotal}
                selectedMediaIds={logic.selectedMediaIds}
                allSelected={logic.allSelected}
                hasMoreVideos={logic.hasMoreVideos}
                videoLoadingMore={logic.videoLoadingMore}
                selectedSet={logic.selectedSet}
                onToggleSelectAll={logic.toggleSelectAllMedia}
                onClearSelection={logic.clearSelection}
                onDownloadSelected={() => void logic.downloadSelectedMedia()}
                onDownloadAll={() => void logic.downloadAllMedia()}
                onToggleSelect={logic.toggleSelectMedia}
                onDownloadOne={(m) => void logic.downloadOneMedia(m)}
                onRetryVideo={() => void logic.loadVideoList(false)}
                onLoadMoreVideos={() => void logic.loadVideoList(true)}
                ssLoadedOnce={logic.ssLoadedOnce}
                ssLoading={logic.ssLoading}
                ssError={logic.ssError}
                listSidebarItems={logic.listSidebarItems as any}
                selectedListId={() => logic.selectedListItem()?.id ?? null}
                listDetailLoading={logic.listDetailLoading}
                listDetailError={logic.listDetailError}
                listCards={logic.listCards}
                listTotal={logic.listTotal}
                listLoadingMore={logic.listLoadingMore}
                hasMoreListVideos={logic.hasMoreListVideos}
                selectedListItem={() => {
                    const it = logic.selectedListItem();
                    return it ? {subtitle: it.subtitle ?? '', title: it.title, count: it.count} : null;
                }}
                onEnsureLoadSeasonsSeries={() => {
                    if (!logic.ssLoadedOnce()) void logic.loadSeasonsSeriesAll();
                }}
                onRefreshSeasonsSeries={() => void logic.loadSeasonsSeriesAll()}
                onSelectListItem={logic.handleSelectListItem as any}
                onRetryListDetail={logic.retryListDetail}
                onLoadMoreList={logic.handleLoadMoreList}
            />
            <Toast message={message()} type={type()}/>
        </UpCommonLayout>
    );
}
