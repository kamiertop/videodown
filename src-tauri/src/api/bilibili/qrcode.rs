use serde::{Deserialize, Serialize};

#[derive(Deserialize, Serialize)]
pub struct Response {
    code: i32,
    message: String,
    ttl: u32,
    data: ResponseData,
}

#[derive(Deserialize, Serialize)]
struct ResponseData {
    url: String,
    qrcode_key: String,
}

#[tauri::command]
pub fn qrcode() -> Result<Response, String> {
    let client = reqwest::blocking::Client::new();

    let response = client.get("https://passport.bilibili.com/x/passport-login/web/qrcode/generate")
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
        .send()
        .map_err(|e| format!("网络请求失败: {}", e))?;

    let response: Response = response.json()
        .map_err(|e| format!("接口返回数据解析失败: {}", e))?;

    Ok(response)
}

#[tauri::command]
pub fn is_logged_in() -> bool {
    false
}