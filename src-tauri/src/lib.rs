mod api;
mod settings;

use log::LevelFilter;

fn init_logger() {
    let _ = env_logger::builder()
        .filter_level(LevelFilter::Info)
        .try_init();
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    init_logger();
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .invoke_handler(tauri::generate_handler![
            settings::get_storage,
            api::bilibili::qrcode::qrcode,  // generate_handler! 需要“模块::函数”完整路径，不能是重导出
            api::bilibili::qrcode::is_logged_in,
            api::bilibili::qrcode::poll_qrcode,
            api::bilibili::myinfo::my_info,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
