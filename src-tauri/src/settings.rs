#[tauri::command]
pub fn get_storage() -> Result<String, String> {
    let value = std::fs::read_to_string("path")
        .map_err(|e| format!("读取失败: {}", e))?;
    Ok(value)
}

#[tauri::command]
#[allow(dead_code)]
pub fn set_storage(_value: &str) {}