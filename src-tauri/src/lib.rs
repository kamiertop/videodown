// 声明 api 模块 → 对应 src/api.rs（或 src/api/mod.rs）
mod api;
mod settings;

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .invoke_handler(tauri::generate_handler![
            settings::get_storage,
            api::bilibili::qrcode::qrcode,  // generate_handler! 需要“模块::函数”完整路径，不能是重导出
            api::bilibili::qrcode::is_logged_in,
            api::bilibili::qrcode::poll_qrcode,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
