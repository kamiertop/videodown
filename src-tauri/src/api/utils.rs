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

pub(super) fn bilibli_header_map() -> reqwest::header::HeaderMap {
    let mut headers = reqwest::header::HeaderMap::new();
    headers.insert(
        "User-Agent", reqwest::header::HeaderValue::from_static("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36"),
    );
    headers.insert(
        "Sec-Fetch-Dest", reqwest::header::HeaderValue::from_static("empty"),
    );
    headers.insert(
        "Sec-Fetch-Mode", reqwest::header::HeaderValue::from_static("cors"),
    );
    headers.insert(
        "Sec-Fetch-Site", reqwest::header::HeaderValue::from_static("same-site"),
    );
    headers.insert(
        "Sec-Ch-Ua-Platform", reqwest::header::HeaderValue::from_static("\"Windows\""),
    );
    headers.insert(
        "Sec-Ch-Ua-Mobile", reqwest::header::HeaderValue::from_static("?0"),
    );
    headers.insert(
        "Sec-Ch-Ua", reqwest::header::HeaderValue::from_static("\"Not:A-Brand\";v=\"99\", \"Google Chrome\";v=\"145\", \"Chromium\";v=\"145\""),
    );
    headers.insert(
        "Priority", reqwest::header::HeaderValue::from_static("u=1, i"),
    );

    headers.insert(
        "Accept-Encoding", reqwest::header::HeaderValue::from_static("gzip, deflate, br, zstd"),
    );
    headers.insert(
        "Accept-Language", reqwest::header::HeaderValue::from_static("zh-CN,zh;q=0.9"),
    );
    headers.insert(
        "Accept", reqwest::header::HeaderValue::from_static("*/*"),
    );
    headers.insert(
        "Origin", reqwest::header::HeaderValue::from_static("https://space.bilibili.com"),
    );

    headers
}

use std::{borrow::Cow, ptr};

use thiserror::Error;

const XOR_CODE: u64 = 23442827791579;
const MASK_CODE: u64 = 2251799813685247;

pub const MAX_AID: u64 = 1 << 51;
pub const MIN_AID: u64 = 1;

const BASE: u64 = 58;
const BV_LEN: usize = 12;
const PREFIX: &str = "BV1";

const ALPHABET: [u8; BASE as usize] = [
    b'F', b'c', b'w', b'A', b'P', b'N', b'K', b'T', b'M', b'u', b'g', b'3', b'G', b'V', b'5', b'L',
    b'j', b'7', b'E', b'J', b'n', b'H', b'p', b'W', b's', b'x', b'4', b't', b'b', b'8', b'h', b'a',
    b'Y', b'e', b'v', b'i', b'q', b'B', b'z', b'6', b'r', b'k', b'C', b'y', b'1', b'2', b'm', b'U',
    b'S', b'D', b'Q', b'X', b'9', b'R', b'd', b'o', b'Z', b'f',
];

#[rustfmt::skip]
fn rev(value: u8) -> Option<u8> {
  use std::option::Option::Some as S;
  match value {
    b'F' => S(0),  b'c' => S(1),  b'w' => S(2),  b'A' => S(3),  b'P' => S(4),  b'N' => S(5),  b'K' => S(6),  b'T' => S(7),  b'M' => S(8),
    b'u' => S(9),  b'g' => S(10), b'3' => S(11), b'G' => S(12), b'V' => S(13), b'5' => S(14), b'L' => S(15), b'j' => S(16), b'7' => S(17),
    b'E' => S(18), b'J' => S(19), b'n' => S(20), b'H' => S(21), b'p' => S(22), b'W' => S(23), b's' => S(24), b'x' => S(25), b'4' => S(26),
    b't' => S(27), b'b' => S(28), b'8' => S(29), b'h' => S(30), b'a' => S(31), b'Y' => S(32), b'e' => S(33), b'v' => S(34), b'i' => S(35),
    b'q' => S(36), b'B' => S(37), b'z' => S(38), b'6' => S(39), b'r' => S(40), b'k' => S(41), b'C' => S(42), b'y' => S(43), b'1' => S(44),
    b'2' => S(45), b'm' => S(46), b'U' => S(47), b'S' => S(48), b'D' => S(49), b'Q' => S(50), b'X' => S(51), b'9' => S(52), b'R' => S(53),
    b'd' => S(54), b'o' => S(55), b'Z' => S(56), b'f' => S(57),
    _ => None
  }
}

#[derive(Error, Debug, PartialEq)]
pub enum Error {
    #[error("Av {0} is smaller than {MIN_AID}")]
    AvTooSmall(u64),
    #[error("Av {0} is bigger than {MAX_AID}")]
    AvTooBig(u64),
    #[error("Bv is empty")]
    BvEmpty,
    #[error("Bv is too small")]
    BvTooSmall,
    #[error("Bv is too big")]
    BvTooBig,
    #[error("Bv prefix should be ignore-cased `BV1`")]
    BvInvalidPrefix,
    #[error("Bv is invalid, with invalid char code `{0}`")]
    BvInvalidChar(char),
    #[error("Bv with unicode char")]
    BvWithUnicode,
}

#[allow(dead_code)]
pub fn av2bv(avid: u64) -> Result<String, Error> {
    if avid < MIN_AID {
        return Err(Error::AvTooSmall(avid));
    }
    if avid >= MAX_AID {
        return Err(Error::AvTooBig(avid));
    }

    let mut bytes: [u8; BV_LEN] = [
        b'B', b'V', b'1', b'0', b'0', b'0', b'0', b'0', b'0', b'0', b'0', b'0',
    ];

    let mut bv_idx = BV_LEN - 1;
    let mut tmp = (MAX_AID | avid) ^ XOR_CODE;
    while tmp != 0 {
        let table_idx = tmp % BASE;
        // SAFETY: a positive number mod 58 is in 0..58
        let part = unsafe { ALPHABET.get_unchecked(table_idx as usize) };
        unsafe {
            let ele = bytes.get_unchecked_mut(bv_idx);
            *ele = *part;
        }
        tmp /= BASE;
        bv_idx -= 1;
    }

    // SAFETY, 3 < 4 < 7 < 9 < BV_LEN
    unsafe {
        unchecked_swap(&mut bytes, 3, 9);
        unchecked_swap(&mut bytes, 4, 7);
    }

    // SAFETY: bytes represent an ASCII string
    let str = unsafe { String::from_utf8_unchecked(bytes.to_vec()) };

    Ok(str)
}

#[allow(dead_code)]
pub fn bv2av<'a, S>(bvid: S) -> Result<u64, Error>
where
    S: Into<Cow<'a, str>>,
{
    let bvid: Cow<_> = bvid.into();
    if bvid.is_empty() {
        return Err(Error::BvEmpty);
    }

    if !bvid.is_ascii() {
        return Err(Error::BvWithUnicode);
    }

    match bvid.as_bytes().len().cmp(&BV_LEN) {
        std::cmp::Ordering::Less => return Err(Error::BvTooSmall),
        std::cmp::Ordering::Greater => return Err(Error::BvTooBig),
        _ => {}
    }

    // SAFETY: Already checked before
    let prefix = unsafe { bvid.get_unchecked(0..3) };
    if !prefix.eq_ignore_ascii_case(PREFIX) {
        return Err(Error::BvInvalidPrefix);
    }

    let mut bvid = match bvid {
        Cow::Borrowed(str) => str.to_string(),
        Cow::Owned(string) => string,
    };

    unsafe {
        let bv_vec = bvid.as_mut_vec();
        unchecked_swap(bv_vec, 3, 9);
        unchecked_swap(bv_vec, 4, 7);
    }

    let mut tmp = 0;

    for byte in &bvid.as_bytes()[3..] {
        let Some(idx) = rev(*byte) else {
            return Err(Error::BvInvalidChar(*byte as char));
        };
        tmp = tmp * BASE + idx as u64;
    }

    // Equivalence of: format!("{:b}", tmp).size()
    let bin_len = if tmp == 0 {
        0
    } else {
        u64::BITS - tmp.leading_zeros()
    };

    if bin_len > 52 {
        return Err(Error::BvTooBig);
    }

    if bin_len < 52 {
        return Err(Error::BvTooSmall);
    }

    let avid = (tmp & MASK_CODE) ^ XOR_CODE;

    if avid < MIN_AID {
        return Err(Error::BvTooSmall);
    }

    Ok(avid)
}

unsafe fn unchecked_swap<I>(array: &mut [I], index_a: usize, index_b: usize) {
    let pa = ptr::addr_of_mut!(array[index_a]);
    let pb = ptr::addr_of_mut!(array[index_b]);
    unsafe {
        ptr::swap(pa, pb);
    }
}