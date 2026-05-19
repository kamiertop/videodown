# CHANGELOG

## v0.11.0

### Fixed

- 下载BiliBili长视频时Context被提前关闭导致下载失败的问题

### Feature

- BiliBili选择视频下载时自动检测分P并展开

### Enhancement

- 重构了下载页面的视频卡片
- 提升了BiliBili视频展示页面的性能
- MacOS平台直接发布二进制，跳过签名

## v0.12.0

### Fixed

- 未使用a_bogus导致获取抖音用户信息失败
- 未设置重试机制导致抖音下载视频时偶尔出现下载失败的问题：net/http: TLS handshake timeout
- 抖音视频封面下载选择错误

### Feature

- 哔哩哔哩和抖音增加搜索下载历史的功能
- 抖音和哔哩哔哩保存下载封面到历史记录
- 解析抖音精选视频：形如：https://www.douyin.com/jingxuan?modal_id=7636717753054891300

### Enhancement

- 修改封面保存路径至作者目录下

### Refactor

- 增加代码注释，提升代码可读性
- 调整部分代码结构

## v0.12.1

### Fixed

- 修复默认存储目录和程序名相同导致无法创建的问题
- 修复使用运行时环境变量判断模式导致日志无法记录到文件的问题，改为编译时写入

### Enhancement

- 将日志文件记录到程序所在目录
- 确保存储文件 `videodown.db` 在程序所在目录
- 缩小程序窗口默认大小，下个版本新增配置以允许用户自定义（适配多屏幕+居中显示没成功）