use std::io::Error;

fn main() -> Result<(), Error> {
    let db = sled::open("db")?;
    let cookies = db.get("bilibili_cookies")?;
    if cookies.is_some() {
        if let Some(cookie) = cookies {
            // 将字节数据转换为字符串
            match std::str::from_utf8(&cookie) {
                Ok(cookie_str) => {
                    println!("Cookie内容:\n{}", cookie_str);
                }
                Err(e) => {
                    println!("Cookie解码失败: {}", e);
                    println!("原始字节: {:?}", cookie);
                }
            }
        }
    } else {
        println!("未找到 bilibili_cookies");
    }

    Ok(())
}