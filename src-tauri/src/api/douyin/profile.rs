use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
#[allow(dead_code)]
pub struct Other {
    user: User,
}

#[derive(Serialize, Deserialize)]
#[allow(dead_code)]
pub struct User {
    aweme_count: u32, // 抖音作品数
    ip_location: String,
    nickname: String,   // 昵称
    sed_uid: String,    // 抖音id
    signature: String,  // 个性签名
    uid: String,
    unique_id: String,
    avatar_300x300: Avatar,
}

#[derive(Serialize, Deserialize)]
#[allow(dead_code)]
struct Avatar {
    height: u16,
    width: u16,
    uri: String,
    url_list: Vec<String>,
}

// #[tauri::command]
// pub async fn profile(user_id: &str) -> Result<User, String> {
//     let resp = build_client()?
//         .get("https://www-hj.douyin.com/aweme/v1/web/user/profile/other/")
//         .query(&[
//             ("device_platform","webapp"),
//             ("aid","6383"),
//             ("channel","channel_pc_web"),
//             ("sec_user_id",user_id),
//             ("publish_video_strategy_type","2"),
//             ("source","channel_pc_web"),
//             ("person_center_strategy","1"),
//             ("profile_other_record_enable","1"),
//             ("land_to","1"),
//             ("update_version_code","170400"),
//             ("pc_client_type","1"),
//             ("pc_libra_divert","Windows"),
//             ("support_h256","0"),
//             ("support_dash","0"),
//             ("cpu_core_num","24"),
//             ("version_code","170400"),
//             ("version_name","17.4.0"),
//             ("cookie_enabled","true"),
//             ("screen_width","2048"),
//             ("screen_height","1152"),
//             ("browser_language","zh-CN"),
//             ("browser_name","Chrome"),
//             ("browser_platform","Win32"),
//             ("browser_version","145.0.0.0"),
//             ("browser_online","true"),
//             ("engine_name","Blink"),
//             ("engine_version","145.0.0.0"),
//             ("os_name","Windows"),
//             ("os_version","10"),
//             ("device_memory","8"),  // 不知道这里为什么是8
//             ("platform","PC"),
//             ("downlink","10"),
//             ("effective_type","4g"),
//             ("round_trip_time","50"),
//             // ("webid","") TODO
//             // ("uifid","") TODO
//             // ("msToken",)
//             // a_bogus
//             // verifyFp
//             // fp
//         ])
//
// }