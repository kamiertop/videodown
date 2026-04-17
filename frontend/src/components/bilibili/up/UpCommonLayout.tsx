import type {JSXElement} from "solid-js";

/**
 * Bilibili UP 页面通用布局：
 * - 顶部 header 支持左右区块 + 可选下方扩展区域
 * - 主体内容区固定滚动容器
 *
 * 布局要点：
 * - 外层 `section` 使用 `flex-col` + `overflow-hidden`，避免页面整体出现双滚动条
 * - 主内容区 `main` 使用 `min-h-0 flex-1 overflow-auto`，把滚动“收敛”到内容区
 * - header 左侧用 `min-w-0 flex-1`，保证右侧固定内容不被挤压，左侧超长时自动截断
 */
export default function UpCommonLayout(props: {
  headerLeft: JSXElement;
  headerRight?: JSXElement;
  headerBelow?: JSXElement;
  children: JSXElement;
}): JSXElement {
  return (
    <section class="flex h-full min-h-0 flex-col gap-3 overflow-hidden bg-base-200/40 p-3">
      <header
        class="flex h-12 shrink-0 items-center gap-3 overflow-hidden rounded-xl border border-base-300 bg-base-100 px-4"
      >
        {/* 左侧永远可见；右侧超出则截断 */}
        <div class="flex min-w-0 flex-1 items-center gap-2">
          {props.headerLeft}
        </div>
        {props.headerRight ? (
          <div class="flex shrink-0 items-center gap-2">
            {props.headerRight}
          </div>
        ) : null}
      </header>
      {props.headerBelow ? (
        <div class="shrink-0">
          {props.headerBelow}
        </div>
      ) : null}

      <main class="min-h-0 flex-1 overflow-auto rounded-xl border border-base-300 bg-base-100">
        {props.children}
      </main>
    </section>
  );
}

