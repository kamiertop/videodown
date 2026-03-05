use crate::api::bilibili::nav::get_nav_key;
use crate::api::utils::{bilibli_header_map, build_client, enc_wbi_params};
use serde::{Deserialize, Serialize};
use std::collections::{BTreeMap, HashMap};

#[derive(Serialize, Deserialize, Debug)]
struct SearchResp {
    code: i32,
    message: String,
    ttl: i32,
    data: Data,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct Data {
    list: List,
    page: Page,
}

#[derive(Serialize, Deserialize, Debug)]
struct List {
    tlist: HashMap<String, TList>,
    vlist: Vec<VList>,
}

#[derive(Serialize, Deserialize, Debug)]
struct TList {
    tid: u32,
    count: u32,
    name: String,
}

#[derive(Serialize, Deserialize, Debug)]
struct VList {
    comment: u32,
    typeid: i32,
    play: u32,
    pic: String,
    subtitle: String,
    description: String,
    copyright: String,
    title: String,
    author: String,
    mid: u32,
    created: u64,
    length: String, // 时长
    aid: u64,
    bvid: String,
    meta: Option<Meta>,
    is_self_view: bool,  // 是否为仅自己可见
    elec_arc_type: i8,      // 充电为1, 否则为0
    is_charging_arc: bool, // 是否是充电视频, 充电为1, 否则为0
    is_live_playback: u8,  // 是否为直播回放
    is_steins_gate: u8, // 是否为互动视频
    is_union_video: u8,  // 是否为合作视频
}

/// 所属合集或课堂
#[derive(Serialize, Deserialize, Debug)]
struct Meta {
    id: u64,
    title: String,
    cover: String,
    mid: u32,
}

#[derive(Serialize, Deserialize, Debug)]
struct Page {
    pn: u32,    // 页码
    ps: u32,    // 每页项数
    count: u32,
}

#[tauri::command]
#[allow(dead_code)]
pub async fn user_video_info(mid: &str, cookie: &str, pn: u32, ps: u32) -> Result<Data, String> {
    let mut pn_str = "1".to_string();
    let mut ps_str = "30".to_string();
    if pn > 1 {
        pn_str = pn.to_string();
    }
    if ps > 30 {
        ps_str = ps.to_string();
    }
    let mut params = BTreeMap::new();
    params.insert("mid", mid.to_string());
    params.insert("pn", pn_str);
    params.insert("ps", ps_str);
    params.insert("order", "pubdate".to_string());
    params.insert("tid", "".to_string());
    params.insert("keyword", "".to_string());
    params.insert("web_location", "333.1387".to_string());
    params.insert("order_avoided", "true".to_string());
    params.insert("platform", "web".to_string());
    params.insert("special_type", "".to_string());
    let (img_key, sub_key) = get_nav_key(cookie).await?;
    let signed_params = enc_wbi_params(params, &img_key, &sub_key);

    let resp = build_client()?.
        get("https://api.bilibili.com/x/space/wbi/arc/search")
        .query(&signed_params)
        .header("Cookie", cookie)
        .header("Referer", format!("https://space.bilibili.com/{}/upload/video", mid))
        .headers(bilibli_header_map())
        .send().await.map_err(|e| e.to_string())?;
    if !resp.status().is_success() {
        return Err(format!("获取视频列表信息失败: {:?}", resp.status()));
    }
    let response: SearchResp = resp.json().await.map_err(|e| e.to_string())?;
    if response.code != 0 {
        return Err(response.message);
    }

    Ok(response.data)
}

#[derive(Serialize, Deserialize, Debug)]
struct DetailResp {
    code: i32,
    message: String,
    ttl: u32,
    data: DetailData,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct DetailData {
    #[serde(rename = "View")]
    view: View,
}

#[derive(Serialize, Deserialize, Debug)]
struct View {
    bvid: Option<String>,
    aid: u64,
    videos: u32,    // 稿件分P总数, 默认为1
    cid: Option<u64>,       // 视频1Pcid
    dimension: Dimension, //视频1P分辨率
    pic: String,    // 稿件封面图片url
    title: String,  // 稿件标题
    pubdate: u64,   // 稿件发布时间
    ctime: u64,     // 用户投稿时间
    dynamic: String,    // 简介
    desc: String,   // 视频简介
    duration: u64,  // 稿件总时长(所有分P)
    owner: Option<Owner>,   // 视频作者信息
    is_upower_exclusive: Option<bool>, // 是否为充电视频
    is_story: Option<bool>, // 是否为动态视频
    pages: Option<Vec<VideoPage>>, // 视频分P信息,
    ugc_season: Option<Season>,
}
// 合集
#[derive(Serialize, Deserialize, Debug)]
struct Season {
    id: u64,        // 视频合集id
    title: String,  // 视频合集标题
    cover: String,  // 视频合集封面图片url
    mid: u64,       // 作者mid
    intro: String,  // 视频合集简介
    sections: Vec<SeasonSection>,
    ep_count: u32,  // 视频合集中的视频数量
    is_pay_season: bool, // 是否为付费合集
}
#[derive(Serialize, Deserialize, Debug)]
struct SeasonSection {
    season_id: u64, // 视频合集中分部所属视频合集id
    section_id: Option<u64>, // 视频合集分部id
    title: String, // 视频合集分部标题
    episodes: Vec<Episode>,
}
#[derive(Serialize, Deserialize, Debug)]
struct Episode {
    season_id: u64,
    section_id: u64,
    id: u64,
    aid: u64,
    cid: u64,
    title: String,
    arc: Option<View>,
    bvid: Option<String>,
    page: VideoPage,
    pages: Vec<VideoPage>,
}

#[derive(Serialize, Deserialize, Debug)]
struct Dimension {
    width: u32,
    height: u32,
    rotate: i32,
}

#[derive(Serialize, Deserialize, Debug)]
struct Owner {
    mid: u64,
    name: String,
    face: String,
}

#[derive(Serialize, Deserialize, Debug)]
struct VideoPage {
    cid: u64,
    page: u32,  // 分P序号
    part: String,   // 分P标题
    duration: u64,
    dimension: Dimension,
    first_frame: Option<String>,
    ctime: Option<u64>,
}

#[tauri::command]
#[allow(dead_code)]
pub async fn video_detail(avid: u64, cookie: &str) -> Result<DetailData, String> {
    let mut params = BTreeMap::new();
    params.insert("aid", avid.to_string());
    params.insert("need_view", "1".to_string());
    params.insert("isGaiaAvoided", "false".to_string());
    params.insert("web_location", "1315873".to_string());
    let (img_key, sub_key) = get_nav_key(cookie).await?;
    let signed_params = enc_wbi_params(params, &img_key, &sub_key);
    let resp = build_client()?
        .get("https://api.bilibili.com/x/web-interface/wbi/view/detail")
        .query(&signed_params)
        .header("Cookie", cookie)
        .headers(bilibli_header_map())
        .send().await.map_err(|e| e.to_string())?;
    if !resp.status().is_success() {
        return Err(format!("获取视频详情信息失败: {:?}", resp.status()));
    };

    let response: DetailResp = resp.json().await.map_err(|e| {
        eprintln!("反序列化错误：{:?}", e);
        format!("获取视频详情信息失败：{:?}", e)
    })?;

    if response.code != 0 {
        return Err(response.message);
    }

    Ok(response.data)
}
