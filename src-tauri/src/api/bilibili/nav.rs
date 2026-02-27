use crate::api::utils::build_client;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug)]
struct Nav {
    code: i32,
    message: String,
    ttl: i32,
    data: NavData,
}

#[derive(Serialize, Deserialize, Debug)]
struct NavData {
    #[serde(rename = "isLogin")]
    is_login: bool,
    mid: u64,
    money: f64,
    wbi_img: NavWbiImg,
}

#[derive(Serialize, Deserialize, Debug)]
pub(super) struct NavWbiImg {
    img_url: String,
    sub_url: String,
}

pub(super) async fn nav(cookie: &str) -> Result<NavWbiImg, String> {
    let resp = build_client()?
        .get("https://api.bilibili.com/x/web-interface/nav")
        .header("Cookie", cookie)
        .header("Accept", "*/*")
        .header("Origin", "https://space.bilibili.com")
        .header("Accept-Encoding", "gzip, deflate, br, zstd")
        .header("Accept-Language", "zh-CN,zh;q=0.9")
        .header("Sec-Ch-Ua", "\"Not:A-Brand\";v=\"99\", \"Google Chrome\";v=\"145\", \"Chromium\";v=\"145\"")
        .header("Sec-Ch-Ua-Mobile", "?0")
        .header("Sec-Ch-Ua-Platform", "\"Windows\"")
        .header("Sec-Fetch-Dest", "empty")
        .header("Sec-Fetch-Mode", "cors")
        .header("Sec-Fetch-Site", "same-site")
        .header("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
        .send().await
        .map_err(|e| e.to_string())?;

    let nav: Nav = resp.json().await.map_err(|e| e.to_string())?;

    Ok(nav.data.wbi_img)
}

/// 从 wbi 图片 URL 中提取 img_key 和 sub_key（去掉后缀的文件名部分）
pub(super) async fn get_nav_key(cookie: &str) -> Result<(String, String), String> {
    let wbi_url = nav(cookie).await?;
    let img_key = wbi_url
        .img_url
        .rsplit('/')          // 取最后一段文件名
        .next()
        .ok_or_else(|| "img_url 解析失败".to_string())?
        .split('.')
        .next()
        .ok_or_else(|| "img_key 提取失败".to_string())?
        .to_string();

    let sub_key = wbi_url
        .sub_url
        .rsplit('/')          // 取最后一段文件名
        .next()
        .ok_or_else(|| "sub_url 解析失败".to_string())?
        .split('.')
        .next()
        .ok_or_else(|| "sub_key 提取失败".to_string())?
        .to_string();

    Ok((img_key, sub_key))
}

#[derive(Serialize, Deserialize, Debug)]
struct NavNumResp {
    code: i32,
    message: String,
    ttl: i32,
    data: NavNumData,
}
#[derive(Serialize, Deserialize, Debug)]
struct NavNumData {
    video: u32, // 视频数量
    opus: u32, // 图文数量
}

#[tauri::command]
#[allow(dead_code)]
pub async fn video_num(mid: &str, cookie: &str) -> Result<u32, String> {
    let resp = build_client()?
        .get("https://api.bilibili.com/x/space/navnum")
        .query(&[("mid", mid), ("web_location", "333.1387")])
        .header("Cookie", cookie)
        .header("Accept", "*/*")
        .header("Origin", "https://space.bilibili.com")
        .header("Accept-Encoding", "gzip, deflate, br, zstd")
        .header("Accept-Language", "zh-CN,zh;q=0.9")
        .header("Sec-Ch-Ua", "\"Not:A-Brand\";v=\"99\", \"Google Chrome\";v=\"145\", \"Chromium\";v=\"145\"")
        .header("Sec-Ch-Ua-Mobile", "?0")
        .header("Sec-Ch-Ua-Platform", "\"Windows\"")
        .header("Sec-Fetch-Dest", "empty")
        .header("Sec-Fetch-Mode", "cors")
        .header("Sec-Fetch-Site", "same-site")
        .header("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
        .send().await.map_err(|e| e.to_string())?;
    if !resp.status().is_success() {
        return Err(resp.text().await.unwrap());
    }
    let response: NavNumResp = resp.json().await.map_err(|e| e.to_string())?;
    if response.code != 0 {
        return Err(format!("请求up主视频总数接口失败: {}", response.message));
    }

    Ok(response.data.video)
}
