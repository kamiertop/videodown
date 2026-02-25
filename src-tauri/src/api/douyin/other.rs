use serde::{Deserialize, Serialize};

#[derive(Serialize,Deserialize)]
pub struct Other {
    user: User
}

#[derive(Serialize,Deserialize)]
struct User {
    aweme_count: u32, // 抖音作品数
    ip_location: String,
    nickname: String,   // 昵称
    sed_uid: String,    // 抖音id
    signature: String,  // 个性签名
    uid: String,
    unique_id: String,
    avatar_300x300: Avatar
}

#[derive(Serialize,Deserialize)]
struct Avatar {
    height: u16,
    width: u16,
    uri: String,
    url_list: Vec<String>
}