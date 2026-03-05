use crate::api::bilibili::nav::get_nav_key;
use crate::api::utils::{bilibli_header_map, build_client, enc_wbi_params};
use serde::{Deserialize, Serialize};
use std::collections::BTreeMap;


#[derive(Serialize, Deserialize)]
pub struct Info {
    code: i32,
    message: String,
    ttl: u32,
    data: InfoData,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct InfoData {
    mid: u64,
    name: String,
    sex: String,
    face: String,   // 头像链接
    sign: String,
    level: u8,
}


#[tauri::command]
#[allow(dead_code)]
pub async fn info(mid: &str, cookie: &str) -> Result<InfoData, String> {
    let (img_key, sub_key) = get_nav_key(cookie).await?;
    // 2. 组装原始参数
    let mut params = BTreeMap::new();
    params.insert("mid", mid.to_string());
    params.insert("token", "".to_string());
    params.insert("platform", "web".to_string());
    params.insert("web_location", "1550101".to_string());
    // 这几个 dm_* 字段目前使用静态值，后续可按需要接入风险埋点
    params.insert("dm_img_list", "[]".to_string());
    params.insert(
        "dm_img_str",
        "V2ViR0wgMS4wIChPcGVuR0wgRVMgMi4wIENocm9taXVtKQ".to_string(),
    );
    params.insert(
        "dm_cover_img_str",
        "QU5HTEUgKE1pY3Jvc29mdCwgTWljcm9zb2Z0IEJhc2ljIFJlbmRlciBEcml2ZXIgKDB4MDAwMDAwOEMpIERpcmVjdDNEMTEgdnNfNV8wIHBzXzVfMCwgRDNEMTEpR29vZ2xlIEluYy4gKE1pY3Jvc29mdC".to_string(),
    );
    params.insert(
        "dm_img_inter",
        "%7B%22ds%22:[],%22wh%22:[3412,1439,78],%22of%22:[124,248,124]%7D".to_string(),
    );

    // 3. 做 WBI 签名，内部会添加 wts 和 w_rid
    let signed_params = enc_wbi_params(params, &img_key, &sub_key);

    // 4. 发起请求
    let resp = build_client()?
        .get("https://api.bilibili.com/x/space/wbi/acc/info")
        .query(&signed_params)
        .header("Cookie", cookie)
        .header("Referer", format!("https://space.bilibili.com/{}/upload/opus", mid))
        .headers(bilibli_header_map())
        .send().await.map_err(|e| format!("请求失败: {:?}", e))?;

    if !resp.status().is_success() {
        return Err("获取up主空间信息失败".to_string());
    }

    let response: Info = resp.json().await.map_err(|e| format!("解析响应失败: {:?}", e))?;

    if response.code != 0 {
        return Err(response.message);
    }

    Ok(response.data)
}
