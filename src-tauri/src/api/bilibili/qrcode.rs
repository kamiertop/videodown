use std::collections::{HashMap, HashSet};
use std::fmt::format;
use serde::{Deserialize, Serialize};
use tauri::http::HeaderMap;

#[derive(Deserialize, Serialize)]
pub struct Response {
    code: i32,
    message: String,
    ttl: i32,
    data: ResponseData,
}

#[derive(Deserialize, Serialize)]
struct ResponseData {
    url: String,
    qrcode_key: String,
}

#[derive(Deserialize, Serialize)]
pub struct PollQrcodeResponse {
    code: i32,
    message: String,
    ttl: i32,
    data: PollQrcodeResponseData,
}

#[derive(Deserialize, Serialize)]
pub struct PollQrcodeResponseData {
    url: String,
    refresh_token: String,
    timestamp: u64,
    code: u32,
    message: String,
}

#[tauri::command]
pub async fn qrcode() -> Result<Response, String> {
    let client = reqwest::Client::new();

    let response = client.get("https://passport.bilibili.com/x/passport-login/web/qrcode/generate")
        .query(&[
            ("source","main-fe-header"),("go_url","https://www.bilibili.com"),("x-bili-locale-json","")
        ])
        .header("accept","*/*")
        .header("accept-language", "zh-CN,zh;q=0.9")
        .header("priority","u=1, i")
        .header("sec-ch-ua","\"Not:A-Brand\";v=\"99\", \"Google Chrome\";v=\"145\", \"Chromium\";v=\"145\"")
        .header("sec-ch-ua-mobile","?0")
        .header("sec-ch-ua-platform","\"Windows\"")
        .header("sec-fetch-dest","empty")
        .header("sec-fetch-mode","cors")
        .header("sec-fetch-site","same-site")
        .header("Referer","https://www.bilibili.com/")
        .send().await
        .map_err(|e| format!("网络请求失败: {}", e))?;

    let response: Response = response.json().await
        .map_err(|e| format!("接口返回数据解析失败: {}", e))?;

    Ok(response)
}

#[tauri::command]
pub fn is_logged_in() -> bool {
    if let Ok(tree) = sled::open("db") {
        if let Ok(res) = tree.get("bilibili_cookies") {
            if res.is_some() {
                return true;
            }
        }
    }

    false
}

/// code: 86101: 未扫码
/// code: 86038: 二维码已失效
/// code: 86090: 二维码已扫描未确认
/// code: 0: 扫码登录成功
#[tauri::command]
pub async fn poll_qrcode(qrcode_key: &str) -> Result<PollQrcodeResponseData, String>{
    let client = reqwest::Client::new();
    let resp = client.get("https://passport.bilibili.com/x/passport-login/web/qrcode/poll")
        .query(&[
            ("qrcode_key", qrcode_key),
            ("source","main-fe-header"),
            ("x-bili-locale-json","%7B%22c_locale%22:%7B%22language%22:%22zh%22,%22region%22:%22CN%22%7D,%22always_translate%22:true%7D")
        ])
        .header("Referer","https://www.bilibili.com/")
        .header("sec-ch-ua","\"Not:A-Brand\";v=\"99\", \"Google Chrome\";v=\"145\", \"Chromium\";v=\"145\"")
        .header("sec-ch-ua-mobile","?0")
        .header("sec-ch-ua-platform","\"Windows\"")
        .send().await
        .map_err(|e|format!("轮询二维码扫描状态失败: {}", e))?;

    if resp.status() != reqwest::StatusCode::OK {
        return Err(format!("轮询二维码扫描状态失败: {}", resp.status()));
    }
    let resp_headers = resp.headers().clone();
    let response = resp.json::<PollQrcodeResponse>().await.map_err(|e|format!("接口返回数据解析失败: {}", e))?;
    if response.code != 0 || response.message != "OK" {
        return Err(format!("轮询二维码扫描状态失败: {}", response.message));
    }
    if response.data.code == 0 { // 扫码成功, 保存登录信息
        if let Some(err_str) = save_cookies(resp_headers) {
            return Err(err_str);
        }
    }

    Ok(response.data)
}

fn save_cookies(header_map: HeaderMap) -> Option<String>{
    let mut result = String::new();
    let mut cookie_set = HashSet::new();
    // 保存所有的 set-cookie 到集合中
    for cookies in header_map.get_all("set-cookie") {
        if let Ok(cookie_str) = cookies.to_str() {
            let cookie_list: Vec<&str> = cookie_str.split("; ").collect();
            for cookie in cookie_list {
                if cookie.contains("=") {
                    cookie_set.insert(cookie);
                }
            }
        }
    }
    let required = ["SESSDATA","bili_jct", "DedeUserID","DedeUserID__ckMd5","sid"];
    // 提取必要的 cookie, 拼接成字符串
    for cookie in required {
        if let Some(value) = cookie_set.get(cookie) {
            result.push_str(value);
            result.push(';');
        }
    }
    // 移除多余的分号
    if result.ends_with(";") {
        result.pop();
    }

    if let Ok(tree) = sled::open("db") {
        let result = tree.insert("bilibili_cookies", result.as_bytes());
        if result.is_ok() {
            log::info!("保存cookie成功");
            None
        } else {
            Some(format!("保存cookie失败: {:?}", result.err()))
        }
    } else {
        Some("打开数据库失败".to_string())
    }
}