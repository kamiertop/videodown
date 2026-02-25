use serde::{Deserialize, Serialize};

#[derive(Serialize,Deserialize)]
pub struct Info {
    code: i32,
    message: String,
    ttl: i32,
    data: Data
}

#[derive(Serialize,Deserialize)]
struct Data {
    coins: f32,
    following: u32,
    follower: u32,
    profile: Profile,
    level_exp: LevelExp,
}

#[derive(Serialize,Deserialize)]
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

#[derive(Serialize,Deserialize)]
struct LevelExp {
    current_level: u8,
    current_min: u32,
    current_exp: u32,
    next_level: i32,
    level_up: i32
}