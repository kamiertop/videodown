use crate::api::utils::{bilibli_header_map, build_client};
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug)]
struct SeriesResp {
    code: i32,
    message: String,
    ttl: u32,
    data: Data,
}
#[derive(Serialize, Deserialize, Debug)]
pub struct Data {
    items_lists: Series,
}
#[derive(Serialize, Deserialize, Debug)]
struct Series {
    page: Page,
    series_list: Vec<SeriesItem>,
    seasons_list: Vec<SeasonsItem>,
}

#[derive(Serialize, Deserialize, Debug)]
struct SeriesItem {
    archives: Vec<Archive>,
    meta: SeriesMeta,
    recent_aids: Vec<u32>,
}

#[derive(Serialize, Deserialize, Debug)]
struct SeasonsItem {
    archives: Vec<Archive>,
    meta: SeasonsMeta,
    recent_aids: Vec<u32>,
}

#[derive(Serialize, Deserialize, Debug)]
struct Archive {
    aid: u64,
    bvid: String,
    ctime: u64,
    duration: u64,
    pic: String,
    title: String,
    pubdate: u64,
}
#[derive(Serialize, Deserialize, Debug)]
struct SeasonsMeta {
    category: i32,
    cover: String,
    description: String,
    mid: u64,
    name: String,
    season_id: u64,
    total: u32,
    title: String,
}

#[derive(Serialize, Deserialize, Debug)]
struct SeriesMeta {
    category: i32,
    cover: String,
    ctime: u64,
    description: String,
    keywords: Vec<String>,
    last_update_ts: u64,
    mtime: u64,
    name: String,
    raw_keywords: String,
    series_id: u64,
    state: i32,
    total: u32,
}

#[derive(Serialize, Deserialize, Debug)]
struct Page {
    page_num: u32,
    page_size: u32,
    total: u32,
}

#[tauri::command]
#[allow(dead_code)]
pub async fn seasons_series_list(mid: &str, cookie: &str) -> Result<Data, String> {
    let resp = build_client()?
        .get("https://api.bilibili.com/x/polymer/web-space/seasons_series_list")
        .query(&[("mid", mid), ("page_size", "20"), ("page_num", "1"), ("web_location", "333.1387")])
        .header("cookie", cookie)
        .headers(bilibli_header_map())
        .send()
        .await
        .map_err(|e| e.to_string())?
        .json::<SeriesResp>()
        .await
        .map_err(|e| format!("{:?}", e))?;

    Ok(resp.data)
}

#[tokio::test]
async fn test_seasons_series_list() {
    let resp = seasons_series_list("21241234", "buvid3=65D342B3-95E2-A774-700B-BD581E2F648B56847infoc; b_nut=1762910656; _uuid=F83E8A67-FD1C-167A-F369-5B51056210FA6357168infoc; buvid4=75CEDBA2-1EF3-B37D-46CE-4610A3278C0A57736-025111209-5otwGjAybxfmCoc2t4RgEA%3D%3D; buvid_fp=ac3b8f33e35fe595e9f480efb69876fc; DedeUserID=480886769; DedeUserID__ckMd5=07bc883e752e334a; theme-tip-show=SHOWED; rpdid=|(J~JmuJJRY|0J'u~Ykmu|~ku; theme-avatar-tip-show=SHOWED; LIVE_BUVID=AUTO9117629158752385; CURRENT_QUALITY=80; hit-dyn-v2=1; theme-switch-show=SHOWED; bili_ticket=eyJhbGciOiJIUzI1NiIsImtpZCI6InMwMyIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzI3NjMzMzMsImlhdCI6MTc3MjUwNDA3MywicGx0IjotMX0._DCUVeo1nB2S9b7fVo1Pxd41LZV2bCEWlMo7AvTxSNk; bili_ticket_expires=1772763273; SESSDATA=6c6dc758%2C1788056134%2Ca3e9e%2A32CjAdW10E8CPkD3h8TRD4irNMosQTOpTdLf2oAENvYIU3qiaE8kB7RQr9SyTI4-38lZ8SVkZPY0puV2VBRlE2TWdtLUtoNW9YWVBtaG1VT0NlamVqTkxkNGkwNnBUb2tYMXI1cUx6T0xFQjVISzM3VWFVTUs5U3Y4di16ZGZGY0x3cmFuQzc3SXR3IIEC; bili_jct=8bcd7f0950c73df1166f197d9fec04d8; sid=4xni660u; PVID=1; home_feed_column=5; bp_t_offset_480886769=1176314439121502208; browser_resolution=1536-711; CURRENT_FNVAL=2000; b_lsid=F5EA9667_19CBE8A2352").await;
    println!("{:#?}", resp);
}