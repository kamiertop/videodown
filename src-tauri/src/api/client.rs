use reqwest::Client;

pub fn build_client() -> Result<Client, String> {
    Client::builder()
        .brotli(true)
        .gzip(true)
        .deflate(true)
        .zstd(true)
        .build()
        .map_err(|e| format!("创建HTTP客户端失败: {}", e))
}

