use crate::api::client::build_client;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
pub struct Info {}

#[tauri::command]
#[allow(dead_code)]
pub async fn info(mid: String) -> Result<Info, String> {
    let resp = build_client()?.get("https://api.bilibili.com/x/space/wbi/acc/info")
        .query(&[
            ("mid", mid),
            ("platform", "web".to_string())
        ])
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
        return Err("获取up主空间信息失败".to_string());
    }
    let response = resp.json().await.map_err(|e| e.to_string())?;

    Ok(response)
}
