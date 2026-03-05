//! 收藏夹

use crate::api::utils::{bilibli_header_map, build_client};
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
struct StarsResp {
    code: u32,
    message: String,
    ttl: u32,
    data: ListData,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct ListData {
    count: u32,
    list: Vec<StarItem>,
}

#[derive(Serialize, Deserialize, Debug)]
struct StarItem {
    id: u64,
    fid: u64,
    mid: u64,
    attr: u32,
    title: String,
    fav_state: i32,
    media_count: u64,
}

/// 获取自己的收藏夹列表
#[tauri::command]
#[allow(dead_code)]
pub async fn list_all(mid: &str, cookie: &str) -> Result<ListData, String> {
    let resp = build_client()?
        .get("https://api.bilibili.com/x/v3/fav/folder/created/list-all")
        .query(&[("up_mid", mid), ("web_location", "333.1387")])
        .header("Cookie", cookie)
        .headers(bilibli_header_map())
        .send().await.map_err(|e| e.to_string())?;
    if !resp.status().is_success() {
        return Err("请求收藏夹列表接口失败".to_string());
    }
    let resp: StarsResp = resp.json().await.map_err(|e| e.to_string())?;

    if resp.code != 0 {
        return Err(resp.message);
    }

    Ok(resp.data)
}

#[derive(Serialize, Deserialize, Debug)]
struct StarResp {
    code: u32,
    message: String,
    ttl: u32,
    data: StarRespData,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct StarRespData {
    info: Info,
    medias: Vec<MediaItem>,
    has_more: bool,
    ttl: u64,
}

#[derive(Serialize, Deserialize, Debug)]
struct Info {
    fid: u64,
    id: u64,
    mid: u32,
    attr: u32,
    title: String,  // 当前是哪个收藏夹
    cover: String,   // 收藏夹封面
    cover_type: u32,
    ctime: u64,
    mtime: u64,
    state: i32,
}

#[derive(Serialize, Deserialize, Debug)]
struct MediaItem {
    id: u64,
    #[serde(rename = "type")]
    media_type: u32,
    title: String,
    cover: String,
    intro: String,
    page: u32,
    duration: u64,
    attr: u32,
    bv_id: String,
    bvid: String,
}

/// 获取一个收藏夹中的视频列表
#[tauri::command]
#[allow(dead_code)]
pub async fn star_video_list(media_id: &str, cookie: &str, mut pn: u64, mut ps: u64) -> Result<StarRespData, String> {
    if pn <= 0 {
        pn = 1;
    }
    if ps <= 0 {
        ps = 20;
    }
    let resp = build_client()?
        .get("https://api.bilibili.com/x/v3/fav/resource/list")
        .query(&[("media_id", media_id),
            ("pn", pn.to_string().as_str()),
            ("ps", ps.to_string().as_str()),
            ("order", "mtime"), ("type", "0"), ("keyword", ""),
            ("platform", "web"), ("web_location", "333.1387"), ("tid", "0")
        ])
        .header("Cookie", cookie)
        .headers(bilibli_header_map())
        .send().await.map_err(|e| e.to_string())?;

    if !resp.status().is_success() {
        return Err("请求收藏夹中视频列表接口失败".to_string());
    }
    let resp: StarResp = resp.json().await.map_err(|e| e.to_string())?;
    if resp.code != 0 {
        return Err(resp.message);
    }

    Ok(resp.data)
}
