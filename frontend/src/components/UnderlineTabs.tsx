import {For, type JSXElement} from "solid-js";

export interface UnderlineTabItem<T extends string> {
  key: T;
  label: string;
}

/**
 * 下划线式 Tab 切换条（B站 UP 详情页同款）：
 * 通栏等宽按钮，选中项文字高亮 + 底部 2px 下划线。
 * `activeClass` 传完整选中态样式，须含 `border-b-2`，如 B站 'border-b-2 border-success text-success'。
 */
export default function UnderlineTabs<T extends string>(props: {
  tabs: readonly UnderlineTabItem<T>[];
  active: T;
  onChange: (tab: T) => void;
  activeClass: string;
}): JSXElement {
  return (
    <div class="flex shrink-0 border-b border-base-300" role="tablist">
      <For each={props.tabs}>
        {(tab) => (
          <button
              class="flex-1 py-3 text-center text-sm font-bold transition-colors"
              classList={{
                [props.activeClass]: props.active === tab.key,
                'text-base-content/60 hover:text-base-content': props.active !== tab.key,
              }}
              type="button"
              role="tab"
              aria-selected={props.active === tab.key}
              onClick={() => props.onChange(tab.key)}
          >
            {tab.label}
          </button>
        )}
      </For>
    </div>
  );
}
