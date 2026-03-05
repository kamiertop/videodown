use crate::api::utils::{bilibli_header_map, build_client};
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
        .headers(bilibli_header_map())
        .send().await
        .map_err(|e| e.to_string())?;

    if !resp.status().is_success() {
        return Err("请求nav接口获取wbi算法img_key,sub_key失败".to_string());
    }

    let nav: Nav = resp.json().await.map_err(|e| format!("解析响应失败: {:?}", e))?;
    if nav.code != 0 {
        return Err(format!("请求nav接口失败: {}", nav.message));
    }

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
        .header("Referer", format!("https://space.bilibili.com/{}/upload/opus", mid))
        .headers(bilibli_header_map())
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