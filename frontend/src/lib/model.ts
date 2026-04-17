export interface MediaCardItem {
  id: number;
  title: string;
  cover: string;
  duration: number;
  bvid: string;
  link?: string;
  upperName: string;
  play?: number;
  danmaku?: number;
  pubtime?: number;
  /** 来源列表名称（收藏夹 / 个人合集 / 合集或系列标题），便于下载时按目录分类 */
  sourceListName: string;
  /** 来源类型：收藏夹、合集、系列、全部投稿 等 */
  sourceListKind: "收藏夹" | "合集" | "系列" | "全部投稿" | "解析结果" | string;

}
