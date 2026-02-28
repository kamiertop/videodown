use crate::api::bilibili::nav::get_nav_key;
use crate::api::utils::{build_client, enc_wbi_params};
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
    let (img_key, sub_key) = get_nav_key(mid, cookie).await?;
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
    let signed_params = enc_wbi_params(params, &img_key, &sub_key);

    let resp = build_client()?.
        get("https://api.bilibili.com/x/space/wbi/arc/search")
        .query(&signed_params)
        .header("Cookie", cookie)
        .header("Accept", "*/*")
        .header("Origin", "https://space.bilibili.com")
        .header("Accept-Encoding", "gzip, deflate, br, zstd")
        .header("Accept-Language", "zh-CN,zh;q=0.9")
        .header("Referer", format!("https://space.bilibili.com/{}/upload/video", mid))
        .header("Priority", "u=1, i")
        .header("Sec-Ch-Ua", "\"Not:A-Brand\";v=\"99\", \"Google Chrome\";v=\"145\", \"Chromium\";v=\"145\"")
        .header("Sec-Ch-Ua-Mobile", "?0")
        .header("Sec-Ch-Ua-Platform", "\"Windows\"")
        .header("Sec-Fetch-Dest", "empty")
        .header("Sec-Fetch-Mode", "cors")
        .header("Sec-Fetch-Site", "same-site")
        .header("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
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

