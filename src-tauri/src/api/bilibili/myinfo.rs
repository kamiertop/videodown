use crate::api::client::build_client;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
#[allow(dead_code)]
pub struct Info {
    code: i32,
    message: String,
    ttl: i32,
    data: Data,
}

#[derive(Serialize, Deserialize, Debug)]
#[allow(dead_code)]
pub struct Data {
    coins: f32,
    // following: u32, // 这两个字段不准
    // follower: u32,  // 这两个字段不准
    profile: Profile,
}

#[derive(Serialize, Deserialize, Debug)]
#[allow(dead_code)]
struct Profile {
    mid: u32,
    name: String,
    sex: String,
    face: String,
    sign: String,
    rank: i32,
    level: u8,
    birthday: u64,
}

#[derive(Serialize, Deserialize)]
struct StatInfo {
    code: i32,
    ttl: i32,
    message: String,
    data: StatInfoData,
}

#[derive(Serialize, Deserialize, Debug)]
#[allow(dead_code)]
pub struct StatInfoData {
    mid: u64,
    following: u32,
    follower: u32,
    black: u32,
}

#[tauri::command]
#[allow(dead_code)]
pub async fn my_info<T: AsRef<str>>(cookie: T) -> Result<Data, String> {
    let resp = build_client()?.get("https://api.bilibili.com/x/space/v2/myinfo")
        .header("Accept", "*/*")
        .header("Accept-Language", "zh-CN,zh;q=0.9")
        .header("Accept-Encoding", "gzip, deflate, br, zstd")
        .header("Origin", "https://space.bilibili.com")
        .header("Priority", "u=1,i")
        .header("Cookie", cookie.as_ref())    // 传入 Cookie
        .header("Sec-Ch-Ua", "\"Not:A-Brand\";v=\"99\", \"Google Chrome\";v=\"145\", \"Chromium\";v=\"145\"")
        .header("Sec-Ch-Ua-Mobile", "?0")
        .header("Sec-Ch-Ua-Platform", "\"Windows\"")
        .header("Sec-Fetch-Dest", "empty")
        .header("Sec-Fetch-Mode", "cors")
        .header("Sec-Fetch-Site", "same-site")
        .header("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
        .send().await.map_err(|e| e.to_string())?;
    if !resp.status().is_success() {
        return Err(format!("请求失败: {}", resp.status()));
    }

    let info: Info = resp.json().await.map_err(|e| format!("反序列化响应体失败: {:?}", e))?;

    if info.code != 0 || info.message.as_str() != "OK" {
        return Err(format!("请求失败: {}", info.message));
    }

    Ok(info.data)
}

#[allow(dead_code)]
#[tauri::command]
pub async fn fans_stat(vmid: String) -> Result<StatInfoData, String> {
    let resp = build_client()?.get("https://api.bilibili.com/x/relation/stat")
        .query(&[("vmid", &vmid)])
        .header("Accept", "*/*")
        .header("Accept-Language", "zh-CN,zh;q=0.9")
        .header("Accept-Encoding", "gzip, deflate, br, zstd")
        .header("Origin", "https://space.bilibili.com")
        .header("Priority", "u=1,i")
        .header("Sec-Fetch-Dest", "empty")
        .header("Sec-Fetch-Mode", "cors")
        .header("Sec-Fetch-Site", "same-site")
        .header("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
        .header("Sec-Ch-Ua", "\"Not:A-Brand\";v=\"99\", \"Google Chrome\";v=\"145\", \"Chromium\";v=\"145\"")
        .header("Sec-Ch-Ua-Mobile", "?0")
        .header("Sec-Ch-Ua-Platform", "\"Windows\"")
        .send().await.map_err(|e| format!("发送请求失败: {}", e))?;

    if !resp.status().is_success() {
        return Err("获取粉丝数/关注数/黑名单数失败".to_string());
    }

    // 尝试解析JSON
    let info: StatInfo = resp.json().await.map_err(|e| format!("反序列化响应体失败: {:?}", e))?;
    if info.code != 0 { // 此处的message是: "0"
        return Err(format!("获取粉丝数/关注数/黑名单数: {}", info.message));
    }

    Ok(info.data)
}

#[cfg(test)]
mod test_my_info {
    use super::*;
    #[tokio::test]
    async fn test_fans_stat() {
        println!("开始测试获取粉丝统计数据...");
        let result = fans_stat("325428150".to_string()).await;
        match result {
            Ok(data) => {
                println!("测试成功! 获取到的数据:");
                println!("  用户ID: {}", data.mid);
                println!("  关注数: {}", data.following);
                println!("  粉丝数: {}", data.follower);
                println!("  黑名单数: {}", data.black);
            }
            Err(err) => {
                println!("测试失败: {:?}", err);
            }
        }
    }
}