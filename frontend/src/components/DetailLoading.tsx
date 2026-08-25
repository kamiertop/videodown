import type {JSXElement} from "solid-js";

export default function DetailLoading(): JSXElement {
  return (
    <div
      class="flex h-full min-h-0 w-full min-w-0 flex-1 items-center justify-center"
      role="status"
      aria-label="正在加载"
    >
      <span class="loading loading-spinner loading-md text-primary" aria-hidden="true"/>
    </div>
  );
}
