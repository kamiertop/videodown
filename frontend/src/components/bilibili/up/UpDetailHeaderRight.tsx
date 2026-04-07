import {Match, Switch, type JSXElement} from "solid-js";
import type {model} from "../../../../wailsjs/go/models";

/**
 * Bilibili UP 详情页头部右侧信息块：
 * - loading 时展示“获取UP主信息...”
 * - 成功后展示头像、昵称、等级、关注状态
 */
export default function UpDetailHeaderRight(props: {
    loading: boolean;
    info: model.UserInfoData | null;
}): JSXElement {
    return (
        <Switch>
            <Match when={props.loading}>
                <div class="flex items-center gap-2">
                    <span class="loading loading-spinner loading-xs text-primary"></span>
                    <span class="text-xs text-base-content/50">获取UP主信息...</span>
                </div>
            </Match>
            <Match when={!props.loading && props.info}>
                <div class="flex min-w-0 items-center gap-2">
                    <div class="h-8 w-8 shrink-0 overflow-hidden rounded-full bg-base-200 ring-2 ring-base-200">
                        <img
                            src={props.info!.face}
                            alt={props.info!.name}
                            referrerPolicy="no-referrer"
                            class="h-full w-full object-cover"
                        />
                    </div>
                    <div class="min-w-0">
                        <div class="flex min-w-0 items-center gap-2">
                            <span class="max-w-[16rem] truncate text-sm font-black text-base-content">
                                {props.info!.name}
                            </span>
                            <span class="badge badge-outline badge-sm">Lv.{props.info!.level}</span>
                            <span
                                class={`badge badge-sm ${props.info!.is_followed ? 'badge-primary' : 'badge-ghost'}`}
                            >
                                {props.info!.is_followed ? '已关注' : '未关注'}
                            </span>
                        </div>
                    </div>
                </div>
            </Match>
        </Switch>
    );
}

