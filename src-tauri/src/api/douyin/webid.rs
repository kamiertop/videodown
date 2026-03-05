use crate::api::utils::build_client;
use serde::{Deserialize, Serialize};


/// 获取web_id
#[tauri::command]
pub async fn get_web_id() -> Result<String, String> {
    #[derive(Serialize, Deserialize, Debug)]
    struct WebID {
        e: i32,
        web_id: String,
    }
    let resp = build_client()?.post("https://mcs.zijieapi.com/webid")
        .query(&[("aid", "6383"), ("sdk_version", "5.1.24_dy")])
        .body(r#"
        {
            "app_id": 6383,
            "referer": "https://www.douyin.com/",
            "url": "https://www.douyin.com/jingxuan",
            "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36",
            "user_unique_id": ""
        }
        "#)
        .header("Accept", "*/*")
        .header("Accept-Encoding", "gzip, deflate, br, zstd")
        .header("Accept-Language", "zh-CN,zh;q=0.9")
        .header("Content-Type", "application/json; charset=UTF-8")
        .header("Origin", "https://www.douyin.com")
        .header("Priority", "u=1, i")
        .header("Referer", "https://www.douyin.com/")
        .header("Sec-Ch-Ua", r#"Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145""#)
        .header("Sec-Ch-Ua-Mobile", "?0")
        .header("Sec-Ch-Ua-Platform", "Windows")
        .header("Sec-Fetch-Dest", "empty")
        .header("Sec-Fetch-Mode", "cors")
        .header("Sec-Fetch-Site", "same-site")
        .header("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
        .send().await.map_err(|e| e.to_string())?;

    if !resp.status().is_success() {
        return Err(format!("{}", resp.status()));
    }

    Ok(resp.json::<WebID>().await.map_err(|e| format!("{:?}", e))?.web_id)
}

#[tokio::test]
async fn test_get_web_id() {
    println!("{:?}", get_web_id().await);
}