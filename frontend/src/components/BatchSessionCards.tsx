import {createSignal, type JSXElement, For, Show} from "solid-js";

/** 批量会话卡片的平台无关视图模型：status 由各平台自己收窄，这里只按字符串展示。 */
export interface BatchStateView {
  title: string;
  totalCount?: number;
  loadedCount: number;
  downloadedCount: number;
  skippedCount: number;
  failedCount: number;
  status: string;
  statusText: string;
  stopping: boolean;
  canResume: boolean;
  active: boolean;
}

export interface BatchSessionView {
  id: number;
  state: BatchStateView;
}

export interface BatchStatusBadge {
  text: string;
  badgeClass: string;
  spinning: boolean;
}

/** 排队中/已结束会话的单行摘要卡，收在展开区里，不占常驻空间。 */
function BatchSessionRowCard(props: {
  session: BatchSessionView;
  statusBadge: (state: BatchStateView) => BatchStatusBadge;
  onDismiss: (id: number) => void;
  onResume: (id: number) => void;
  onStop: (id: number) => void;
}): JSXElement {
  const batch = () => props.session.state;
  const badge = () => props.statusBadge(batch());

  return (
      <div class="flex min-w-0 items-center gap-2 rounded-lg border border-base-300 bg-base-100 px-3 py-1.5 shadow-sm">
        <span class={`badge badge-sm ${badge().badgeClass} shrink-0`}>{badge().text}</span>
        <h3 class="min-w-0 truncate text-sm font-semibold text-base-content" title={batch().title}>
          {batch().title}
        </h3>
        <Show when={batch().failedCount > 0}>
          <span class="badge badge-outline badge-error badge-sm shrink-0">失败 {batch().failedCount}</span>
        </Show>
        <span class="hidden max-w-48 shrink-0 truncate text-xs text-base-content/50 md:inline" title={batch().statusText}>
          {batch().statusText}
        </span>
        <div class="ml-auto flex shrink-0 items-center gap-1.5">
          <Show
              when={batch().active}
              fallback={
                <>
                  <Show when={batch().canResume}>
                    <button
                        class="btn btn-primary btn-xs"
                        type="button"
                        onClick={() => props.onResume(props.session.id)}
                    >
                      继续批量
                    </button>
                  </Show>
                  <button
                      class="btn btn-ghost btn-xs"
                      type="button"
                      onClick={() => props.onDismiss(props.session.id)}
                  >
                    关闭
                  </button>
                </>
              }
          >
            <button
                class="btn btn-warning btn-xs"
                type="button"
                onClick={() => props.onStop(props.session.id)}
            >
              取消排队
            </button>
          </Show>
        </div>
      </div>
  );
}

/**
 * 下载页顶部的批量任务卡片（B 站 / 抖音共用）：默认只占一条摘要——当前执行中
 * 会话的进度条、停止入口和队列概况（N 个排队 / N 个待处理）；排队和已结束的
 * 会话收在“展开”下拉里，单行卡片限高滚动，批量再多也不会把下载列表挤出可视区。
 * 多个批量按点击先后排队串行执行；全部成功的会话卡片几秒后自动关闭。
 */
export function BatchSessionCards(props: {
  sessions: () => readonly BatchSessionView[];
  statusBadge: (state: BatchStateView) => BatchStatusBadge;
  onDismiss: (id: number) => void;
  onResume: (id: number) => void;
  onStop: (id: number) => void;
}): JSXElement {
  const [expanded, setExpanded] = createSignal(false);
  const list = props.sessions;
  const running = () => list().find((session) => session.state.active && session.state.status !== "queued");
  const queuedCount = () => list().filter((session) => session.state.status === "queued").length;
  // 已结束（部分失败/已停止）待用户处理的会话数。
  const finishedCount = () => list().filter((session) => !session.state.active).length;
  // 除执行中外还有别的卡片才有展开的必要。
  const hasOthers = () => list().length > (running() ? 1 : 0);

  return (
      <Show when={list().length > 0}>
        <section class="mt-2 flex flex-col gap-2 rounded-lg border border-base-300 bg-base-100 p-3 shadow-sm">
          <div class="flex min-w-0 flex-wrap items-center gap-2">
            <Show
                when={running()}
                fallback={
                  <>
                    <span class="badge badge-sm badge-info shrink-0">批量任务</span>
                    <span class="min-w-0 truncate text-sm text-base-content/70">
                      {[
                        queuedCount() > 0 ? `${queuedCount()} 个排队` : "",
                        finishedCount() > 0 ? `${finishedCount()} 个待处理` : "",
                      ].filter(Boolean).join(" · ")}
                    </span>
                  </>
                }
            >
              {(session) => {
                const batch = () => session().state;
                const badge = () => props.statusBadge(batch());
                // 接口没给总数（如收藏视频）时用已加载数作分母，翻页过程中分母会持续增长。
                const total = () => batch().totalCount ?? batch().loadedCount;
                // 跳过的内容也算完成进度，否则增量下载时进度条会停在小于总数的位置。
                const done = () => batch().downloadedCount + batch().skippedCount;
                const percent = () => total() > 0
                    ? Math.min(100, Math.round((done() / total()) * 100))
                    : 0;

                return (
                    <>
                      <Show when={badge().spinning}>
                        <span class="loading loading-spinner loading-xs text-primary"></span>
                      </Show>
                      <span class={`badge badge-sm ${badge().badgeClass} shrink-0`}>{badge().text}</span>
                      <h3 class="min-w-0 truncate text-sm font-bold text-base-content" title={batch().title}>
                        {batch().title}
                      </h3>
                      <span class="shrink-0 text-xs tabular-nums text-base-content/60">{percent()}%</span>
                      <span class="hidden max-w-40 truncate text-xs text-base-content/50 md:inline" title={batch().statusText}>
                        {batch().statusText}
                      </span>
                      <Show when={batch().skippedCount > 0}>
                        <span class="badge badge-ghost badge-sm shrink-0">跳过 {batch().skippedCount}</span>
                      </Show>
                      <Show when={queuedCount() > 0}>
                        <span class="badge badge-ghost badge-sm shrink-0">+{queuedCount()} 排队</span>
                      </Show>
                      <Show when={finishedCount() > 0}>
                        <span class="badge badge-ghost badge-sm shrink-0">{finishedCount()} 待处理</span>
                      </Show>
                    </>
                );
              }}
            </Show>

            <div class="ml-auto flex shrink-0 items-center gap-1.5">
              <Show when={running()}>
                {(session) => (
                    <div class="group relative shrink-0">
                      <button
                          class="btn btn-warning btn-xs"
                          type="button"
                          onClick={() => props.onStop(session().id)}
                          disabled={session().state.stopping}
                      >
                        {session().state.stopping ? "正在停止..." : "停止批量"}
                      </button>
                      <span
                          role="tooltip"
                          class="pointer-events-none absolute right-0 top-full z-50 mt-1.5 w-max max-w-[min(240px,calc(100vw-24px))] whitespace-normal rounded-md bg-neutral px-2.5 py-1.5 text-center text-[11px] leading-4 text-neutral-content opacity-0 shadow-lg transition-opacity group-hover:opacity-100 group-focus-within:opacity-100"
                      >
                        当前批次下载完成（含任务间休眠）后才会停止，期间请耐心等待
                      </span>
                    </div>
                )}
              </Show>
              <Show when={hasOthers()}>
                <button
                    class="btn btn-ghost btn-xs gap-1"
                    type="button"
                    onClick={() => setExpanded((value) => !value)}
                    aria-expanded={expanded()}
                >
                  {expanded() ? "收起" : "展开"}
                  <svg
                      class={`h-3 w-3 shrink-0 transition-transform ${expanded() ? "rotate-180" : ""}`}
                      viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2"
                      stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"
                  >
                    <path d="M4 6l4 4 4-4"/>
                  </svg>
                </button>
              </Show>
            </div>
          </div>

          <Show when={running()}>
            {(session) => {
              const batch = () => session().state;
              const total = () => batch().totalCount ?? batch().loadedCount;
              const done = () => batch().downloadedCount + batch().skippedCount;

              return (
                  <div class="flex items-center gap-3">
                    <progress
                        class="progress progress-primary h-2 min-w-0 flex-1"
                        value={done()}
                        max={Math.max(1, total())}
                    />
                    <span class="shrink-0 text-xs tabular-nums text-base-content/70">
                      已下载 {batch().downloadedCount} / {total()}
                    </span>
                  </div>
              );
            }}
          </Show>

          <Show when={expanded() && hasOthers()}>
            <div class="flex max-h-60 flex-col gap-2 overflow-y-auto pr-1">
              <For each={list().filter((session) => session.id !== running()?.id)}>
                {(session) => (
                    <BatchSessionRowCard
                        session={session}
                        statusBadge={props.statusBadge}
                        onDismiss={props.onDismiss}
                        onResume={props.onResume}
                        onStop={props.onStop}
                    />
                )}
              </For>
            </div>
          </Show>
        </section>
      </Show>
  );
}
