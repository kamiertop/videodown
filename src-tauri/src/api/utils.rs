use percent_encoding::{utf8_percent_encode, AsciiSet, NON_ALPHANUMERIC};
use reqwest::Client;
use std::collections::BTreeMap;
use std::time::{SystemTime, UNIX_EPOCH};

pub fn build_client() -> Result<Client, String> {
    Client::builder()
        .brotli(true)
        .gzip(true)
        .deflate(true)
        .zstd(true)
        .build()
        .map_err(|e| format!("创建HTTP客户端失败: {}", e))
}

#[allow(dead_code)]
pub fn timestamp_str() -> Result<String, String> {
    Ok(SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map_err(|e| format!("获取时间戳失败: {}", e))?
        .as_secs().to_string())
}

#[allow(dead_code)]
pub fn timestamp_u64() -> Result<u64, String> {
    Ok(SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map_err(|e| format!("获取时间戳失败: {}", e))?
        .as_secs())
}

// 与 encodeURIComponent 行为一致的编码集合（保留 -_.~）
const WBI_ENCODE_SET: &AsciiSet = &NON_ALPHANUMERIC
    .remove(b'-')
    .remove(b'_')
    .remove(b'.')
    .remove(b'~');

// 来自文档的打乱表：https://sessionhu.github.io/bilibili-API-collect/docs/misc/sign/wbi.html#rust
const MIXIN_KEY_ENC_TAB: [usize; 64] = [
    46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
    33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40,
    61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11,
    36, 20, 34, 44, 52,
];

/// 根据 img_key + sub_key 生成 mixin_key
fn gen_mixin_key(raw_wbi_key: &str) -> String {
    let bytes = raw_wbi_key.as_bytes();
    let mut mixin = String::with_capacity(32);
    for &idx in MIXIN_KEY_ENC_TAB.iter() {
        if let Some(&b) = bytes.get(idx) {
            mixin.push(b as char);
        }
    }
    mixin.truncate(32);
    mixin
}

/// 为参数做 WBI 签名：添加 wts，并计算 w_rid
pub(super) fn enc_wbi_params<'a>(
    mut params: BTreeMap<&'a str, String>,
    img_key: &str,
    sub_key: &str,
) -> BTreeMap<&'a str, String> {
    // 1. 生成 mixin_key
    let mixin_key = gen_mixin_key(&format!("{}{}", img_key, sub_key));

    // 2. 添加 wts（当前秒级时间戳）
    let now = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_secs();
    params.insert("wts", now.to_string());

    // 3. 排序（BTreeMap 已经按 key 升序）
    // 4. 过滤 value 中的 "!'()*" 并按 encodeURIComponent 编码
    let mut encoded_pairs = Vec::new();
    for (k, v) in &params {
        let key_enc = utf8_percent_encode(k, WBI_ENCODE_SET).to_string();
        let filtered: String = v.chars().filter(|c| !"!'()*".contains(*c)).collect();
        let val_enc = utf8_percent_encode(&filtered, WBI_ENCODE_SET).to_string();
        encoded_pairs.push(format!("{}={}", key_enc, val_enc));
    }
    let query = encoded_pairs.join("&");

    // 5. 计算 w_rid = MD5(query + mixin_key)
    let sign = format!("{:x}", md5::compute(format!("{}{}", query, mixin_key)));
    params.insert("w_rid", sign);

    params
}