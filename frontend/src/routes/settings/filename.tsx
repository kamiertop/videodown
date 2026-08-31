import {
  GetFilenameTemplate,
  GetGroupingRule,
  GetSavePreference,
  GetStorage,
  SetFilenameTemplate,
  SetGroupingRule,
  SetSavePreference
} from "@bindings/github.com/kamiertop/videodown/utils/settings";
import {createFileRoute} from "@tanstack/solid-router";
import {
  createSortable,
  DragDropProvider,
  DragDropSensors,
  SortableProvider,
  transformStyle
} from "@thisbeyond/solid-dnd";
import {createSignal, For, type JSXElement, onMount, Show} from "solid-js";
import Toast from "../../components/Toast";
import {useToast} from "../../hooks/useToast";

export const Route = createFileRoute('/settings/filename')({component: FilenameSection});

function SortablePart(props: { id: string; label: string }): JSXElement {
  const sortable = createSortable(props.id);
  const colors: Record<string, string> = {
    "{title}": "badge-primary",
    "{author}": "badge-secondary",
    "{folder}": "badge-accent",
    "{collection}": "badge-info",
    "{publish_date}": "badge-success",
    "{date}": "badge-warning",
    "{time}": "badge-error",
    "{id}": "badge-neutral",
    "{author_id}": "badge-secondary"
  };
  return (
      <div ref={sortable} style={transformStyle(sortable.transform)}
           class={`badge badge-lg gap-2 cursor-grab active:cursor-grabbing select-none ${colors[props.id] || "badge-primary"}`}
      >{props.label}
      </div>
  );
}

function FilenameSection(): JSXElement {
  const [template, setTemplate] = createSignal("");
  const [separator, setSeparator] = createSignal(" - ");
  const [parts, setParts] = createSignal<string[]>(["{title}"]);
  const [loaded, setLoaded] = createSignal(false);
  const [defaultPath, setDefaultPath] = createSignal("读取中...");
  const [grouping, setGrouping] = createSignal(true);
  const [groupRule, setGroupRule] = createSignal("author_source");
  const {message, type, showToast} = useToast();
  onMount(async () => {
    try {
      const value = await GetFilenameTemplate();
      setTemplate(value);
      setParts(value.match(/\{(?:title|source|author|folder|collection|publish_date|date|time|id|author_id)\}/g) || ["{title}"]);
      setDefaultPath(await GetStorage().catch(() => "未设置"));
      setGrouping(await GetSavePreference().catch(() => true));
      setGroupRule(await GetGroupingRule().catch(() => "author_source"));
    } catch {
      setTemplate("{title}");
    } finally {
      setLoaded(true);
    }
  });

  async function save() {
    const value = template().trim();
    if (!value) {
      showToast("命名模板不能为空", "warning");
      return;
    }
    try {
      await Promise.all([SetFilenameTemplate(value), SetSavePreference(grouping()), SetGroupingRule(groupRule())]);
      setTemplate(value);
      showToast("命名模板保存成功", "success");
    } catch (e) {
      showToast(e instanceof Error ? e.message : String(e), "error");
    }
  }

  const fields = [
    ["{title}", "视频标题", "示例视频标题"],
    ["{author}", "作者名称", "UP主 / 作者"],
    ["{folder}", "收藏夹名称", "从收藏夹选择时可用"],
    ["{collection}", "合集名称", "从合集选择时可用"],
    ["{id}", "视频 ID", "BV1xx / aweme_id"],
    ["{author_id}", "作者 ID", "用户唯一 ID"],
    ["{publish_date}", "发布时间", "视频发布时间，格式为 YYYY-MM-DD"],
    ["{date}", "下载日期", "2026-08-30"],
    ["{time}", "下载时间", "14-35-22"],
  ];

  function togglePart(token: string) {
    const next = parts().includes(token) ? parts().filter(p => p !== token) : [...parts(), token];
    setParts(next);
    setTemplate(next.join(separator()));
  }

  function movePart(from: number, to: number) {
    const next = [...parts()];
    const [item] = next.splice(from, 1);
    next.splice(from < to ? to - 1 : to, 0, item);
    setParts(next);
    setTemplate(next.join(separator()));
  }

  return <>
    <Show when={loaded()} fallback={
      <div class="card bg-base-100 shadow-xl">
        <div class="card-body"><span class="loading loading-dots loading-sm" aria-label="读取中"/>
        </div>
      </div>
    }>
      <div class="card bg-base-100 border border-base-300/60 shadow-xl max-w-4xl mx-auto overflow-hidden">
        <div class="card-body gap-6">
          <div><p class="label-text text-sm font-semibold mb-2">选择文件名内容</p>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
              <For each={fields}>
                {(field) => (
                    <button type="button"
                            class={`btn justify-start ${parts().includes(field[0]) ? "btn-primary" : "btn-outline"}`}
                            title={field[2]}
                            onClick={() => togglePart(field[0])}
                    >
                      <span class="text-lg">{parts().includes(field[0]) ? "✓" : "○"}</span>
                      <span>{field[1]}</span>
                      <span class="ml-auto text-xs opacity-60">{field[0]}</span>
                    </button>
                )}
              </For>
            </div>
          </div>
          <div><p class="label-text text-sm font-semibold mb-2">当前顺序 <span
              class="font-normal text-xs text-base-content/50">（拖动 ☰ 调整）</span></p>
            <DragDropProvider onDragEnd={({draggable, droppable}) => {
              if (droppable && draggable) {
                const from = parts().indexOf(String(draggable.id));
                const to = parts().indexOf(String(droppable.id));
                if (from >= 0 && to >= 0 && from !== to) movePart(from, to);
              }
            }}><DragDropSensors><SortableProvider ids={parts()}>
              <div class="flex flex-wrap gap-2 min-h-10 rounded-lg border border-dashed border-base-300 p-2"><For
                  each={parts()}>{(token) => <SortablePart id={token}
                                                           label={fields.find(f => f[0] === token)?.[1] || token}/>}</For>
              </div>
            </SortableProvider></DragDropSensors></DragDropProvider>
          </div>
          <div><p class="label-text text-sm font-semibold mb-2">字段之间的分隔符</p>
            <div class="flex flex-wrap gap-2">
              <For each={[[" - ", "短横线"], ["_", "下划线"], [" ", "空格"], [" | ", "竖线"]]}>{(item) =>
                  <button type="button"
                          class={`btn btn-sm ${separator() === item[0] ? "btn-secondary" : "btn-outline"}`}
                          onClick={() => {
                            setSeparator(item[0]);
                            setTemplate(parts().join(item[0]));
                          }}>
                    <span>{separator() === item[0] ? "✓" : "○"}</span>
                    <span>{item[1]}</span>
                    <span class="text-xl font-black leading-none min-w-5 text-center"
                          aria-label={`分隔符 ${item[1]}`}>{item[0] === " " ? "·" : item[0].trim()}
                  </span>
                  </button>
              }</For>
            </div>
          </div>
          <div class="rounded-xl bg-base-200/70 border-2 border-primary/30 shadow-sm p-4">
            <p class="text-xs uppercase tracking-wide text-base-content/50 mb-1">文件名预览</p>
            <p class="font-medium break-all text-primary">{template() || "（未选择字段）"}.mp4</p>
          </div>
          <label class="flex items-center gap-3">
            <span class="text-sm font-semibold whitespace-nowrap">直接编辑模板</span>
            <span class="text-xs text-base-content/50 whitespace-nowrap">可输入 {"{title}"} 等变量</span>
            <input class="input input-bordered flex-1"
                   value={template()}
                   onInput={e => setTemplate(e.currentTarget.value)}
                   onKeyDown={e => e.key === "Enter" && void save()}/>
          </label>
          <div class="rounded-xl border-2 border-secondary/30 bg-base-100 shadow-sm p-4">
            <div class="flex items-center justify-between gap-4">
              <div><p class="font-semibold">目录分组规则</p><p
                  class="text-xs text-base-content/50 mt-1">选择下载文件的目录组织方式</p></div>
              <input type="checkbox" class="toggle toggle-primary" checked={grouping()}
                     onChange={e => setGrouping(e.currentTarget.checked)}/>
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-2 mt-4" classList={{"opacity-50": !grouping()}}>
              <For
                  each={[["none", "不分组"], ["author", "按作者"], ["author_source", "作者 + 来源"]]}>{item => <button
                  type="button" disabled={!grouping()}
                  class={`btn btn-sm ${groupRule() === item[0] ? "btn-primary" : "btn-outline"}`}
                  onClick={() => setGroupRule(item[0])}>{item[1]}</button>}
              </For>
            </div>
            <div class="mt-4 rounded-lg bg-base-200/70 border border-secondary/30 px-3 py-2 text-sm">
              <span class="text-base-content/50">示例路径：</span>
              <code>
                {defaultPath()}{!grouping() ? "/" : groupRule() === "none" ? "/" : groupRule() === "author" ? "/作者名/" : "/作者名/收藏夹名/"}{template() || "视频标题"}.mp4
              </code>
            </div>
          </div>
          <button class="btn btn-primary" onClick={save}>保存模板</button>
        </div>
      </div>
    </Show>
    <Toast message={message()} type={type()}/>
  </>;
}
