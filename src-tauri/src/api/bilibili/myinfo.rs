use crate::api::utils::{bilibli_header_map, build_client};
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
pub async fn my_info(cookie: &str) -> Result<Data, String> {
    let resp = build_client()?.get("https://api.bilibili.com/x/space/v2/myinfo")
        .query(&[("web_location", "333.1387")])
        .header("Cookie", cookie)
        .headers(bilibli_header_map())
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
pub async fn fans_stat(vmid: &str) -> Result<StatInfoData, String> {
    let resp = build_client()?.get("https://api.bilibili.com/x/relation/stat")
        .query(&[("vmid", vmid), ("web_location", "333.1387")])
        .headers(bilibli_header_map())
        .send().await.map_err(|e| format!("发送请求失败: {}", e))?;

    if !resp.status().is_success() {
        return Err("获取粉丝数/关注数/黑名单数失败".to_string());
    }

    // 尝试解析JSON
    let info: StatInfo = resp.json().await.map_err(|e| format!("反序列化响应体失败: {:?}", e))?;
    if info.code != 0 { // 此处的message是: "0"
        return Err(format!("获取粉丝数/关注数/黑名单数失败: {}", info.message));
    }

    Ok(info.data)
}

#[cfg(test)]
mod test_my_info {
    use super::*;
    #[tokio::test]
    async fn test_fans_stat() {
        println!("开始测试获取粉丝统计数据...");
        let result = fans_stat("325428150").await;
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