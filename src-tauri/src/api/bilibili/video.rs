use crate::api::bilibili::nav::get_nav_key;
use crate::api::utils::{bilibli_header_map, build_client, enc_wbi_params};
use scraper::{Html, Selector};
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
pub async fn user_video_info(mid: &str, cookie: &str, mut pn: u32, mut ps: u32) -> Result<Data, String> {
    if pn <= 0 {
        pn = 1;
    }
    if ps > 30 || ps <= 0 {
        ps = 30;
    }
    let mut params = BTreeMap::new();
    params.insert("mid", mid.to_string());
    params.insert("pn", pn.to_string());
    params.insert("ps", ps.to_string());
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
    videos: u32,            // 稿件分P总数, 默认为1
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

/// 视频详情
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

    let response: DetailResp = resp.json().await.map_err(|e| { format!("获取视频详情信息失败：{:?}", e) })?;

    if response.code != 0 {
        return Err(response.message);
    }

    Ok(response.data)
}

#[tauri::command]
#[allow(dead_code)]
pub async fn play_url(b_vid: &str, cookie: &str) -> Result<PlayUrlResp, String> {
    let resp = build_client()?
        .get(format!("https://www.bilibili.com/video/{b_vid}/"))
        .header("Cookie", cookie)
        .header("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
        .header("Accept-Encoding", "gzip, deflate, br, zstd")
        .header("Accept-Language", "zh-CN,zh;q=0.9")
        .header("Cache-Control", "max-age=0")
        .header("Priority", "u=0, i")
        .header("Referer", "https://www.bilibili.com/")
        .header("Sec-Ch-Ua", r#""Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145""#)
        .header("Sec-Ch-Ua-Mobile", "?0")
        .header("Sec-Ch-Ua-Platform", "\"Windows\"")
        .header("Sec-Fetch-Dest", "document")
        .header("Sec-Fetch-Mode", "navigate")
        .header("Sec-Fetch-Site", "same-origin")
        .header("Sec-Fetch-User", "?1")
        .header("Upgrade-Insecure-Requests", "1")
        .header("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
        .send().await.map_err(|e| e.to_string())?;
    if !resp.status().is_success() {
        return Err(format!("获取视频播放页失败: {:?}", resp.status()));
    }

    let html = resp.text().await.map_err(|e| { format!("获取视频播放页失败：{:?}", e) })?;
    let html = Html::parse_document(html.as_str());
    let element = html
        .select(&Selector::parse("script").map_err(|e| format!("解析视频播放页失败: {:?}", e))?)
        .nth(3);
    if element.is_none() {
        return Err("解析script标签失败".to_string());
    }
    let inner_html = element.unwrap().inner_html();
    let element = inner_html.trim_start_matches("window.__playinfo__=");

    Ok(
        serde_json::from_str::<PlayUrlResp>(element).map_err(|e| { format!("解析视频数据失败: {:?}", e) })?
    )
}

#[tokio::test]
async fn test_video_detail() {
    let detail = play_url("BV1JuPmz4EeS", "buvid3=65D342B3-95E2-A774-700B-BD581E2F648B56847infoc; b_nut=1762910656; _uuid=F83E8A67-FD1C-167A-F369-5B51056210FA6357168infoc; buvid4=75CEDBA2-1EF3-B37D-46CE-4610A3278C0A57736-025111209-5otwGjAybxfmCoc2t4RgEA%3D%3D; buvid_fp=ac3b8f33e35fe595e9f480efb69876fc; DedeUserID=480886769; DedeUserID__ckMd5=07bc883e752e334a; theme-tip-show=SHOWED; rpdid=|(J~JmuJJRY|0J'u~Ykmu|~ku; theme-avatar-tip-show=SHOWED; LIVE_BUVID=AUTO9117629158752385; CURRENT_QUALITY=80; hit-dyn-v2=1; theme-switch-show=SHOWED; PVID=1; bp_t_offset_480886769=1176314439121502208; bmg_af_switch=1; bmg_src_def_domain=i0.hdslb.com; bili_ticket=eyJhbGciOiJIUzI1NiIsImtpZCI6InMwMyIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzMwMzU4NDQsImlhdCI6MTc3Mjc3NjU4NCwicGx0IjotMX0.JsJB-HDT4_mJGuks4iKldF1cKAkpV1ORke1M_L203r0; bili_ticket_expires=1773035784; SESSDATA=de332b00%2C1788328644%2C9a5b3%2A32CjBBVFCxnJCWvsugb3mAlUwwlEuMEJHwbf0jb8GWxBMrOQ5YdJsaPVCiAIj-yW1u78kSVl9lNlJhWE9xak9yai1HMTJ3d1VfeXFQd00tcVZTVHdaRXJET2FobnBlX19sbHJSZmhGZ3ptZllLN3lOMXk5Z2I1TXZCb0NnRXdnT041a29nRXJtXzNnIIEC; bili_jct=482822a6a6eff224f0e4757ebe42c258; sid=6tnvy4ml; home_feed_column=4; browser_resolution=2048-999; CURRENT_FNVAL=4048; b_lsid=CC1BF83C_19CC1D7E820").await;
    println!("{:#?}", detail)
}

#[derive(Serialize, Deserialize, Debug)]
pub struct PlayUrlResp {
    code: i64,
    message: String,
    ttl: i64,
    data: PlayUrlData,
    session: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct PlayUrlData {
    from: String,
    result: String,
    message: String,
    quality: i64,
    format: String,
    timelength: i64,
    accept_format: String,
    accept_description: Vec<String>,
    accept_quality: Vec<i64>,
    video_codecid: i64, // 视频编码
    seek_param: String,
    seek_type: String,
    dash: Dash,
    support_formats: Vec<SupportFormat>,
    last_play_time: i64,
    last_play_cid: i64,
    cur_language: String,
    cur_production_type: i64,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct Dash {
    duration: u64,
    min_buffer_time: f64,
    video: Vec<Video>,
    audio: Vec<Audio>,
}

#[derive(Serialize, Deserialize, Debug)]
struct SupportFormat {
    quality: u32,
    format: String, //格式, 比如: hdflv2
    new_description: String, //格式的描述, 比如: 4K 超高清
    display_desc: String,
    superscript: String,
    codecs: Vec<String>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct Video {
    id: u64,    // 音视频清晰度代码
    base_url: String,
    backup_url: Vec<String>,
    bandwidth: u64,
    mime_type: String,
    width: u32,
    height: u32,
    frame_rate: String, //帧率
    sar: String,
    codecid: i32,   // 7->AVC编码,8k视频不支持; 12->HEVC编码; 13->AV1编码
}

#[derive(Serialize, Deserialize, Debug)]
pub struct Audio {
    id: u64,
    base_url: String,
    backup_url: Vec<String>,
    bandwidth: u64,
    mime_type: String,
    width: u32,
    height: u32,
    frame_rate: String, //帧率
    sar: String,
    codecid: i32,
}
