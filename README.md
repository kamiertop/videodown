# VideoDown

## 免责声明

本项目仅用于个人学习和研究，请勿用于商业用途或侵犯他人权益的行为。

## 功能

### 哔哩哔哩

- 二维码登录
- 关注的UP主
- UP主的所有视频
- 收藏夹、合集、系列
- 解析视频、UP主
- 批量下载视频（收藏夹、系列、UP主）

### 抖音

- 使用Cookie登录（下面2个项目都推荐使用Cookie，使用扫码登录可能会导致抖音账号被风控。没尝试过，但听劝）
    - https://github.com/JoeanAmier/TikTokDownloader
    - https://github.com/Evil0ctal/Douyin_TikTok_Download_API

## 使用说明

1. 由于哔哩哔哩返回的是音频流+视频流，所以需要ffmpeg，请自行安装并配置环境变量（自动下载ffmpeg的功能暂未考虑）

## 说明

- 抖音接口返回的数据太庞大了（难怪要自研[sonic](https://github.com/bytedance/sonic)
  ），所以只记录了关键的字段，如果需要二次开发，可以参考[代码](./douyin/api)
  ，自行在浏览器开发者工具中查看
- 感谢 [bilibili-API-collect](https://sessionhu.github.io/bilibili-API-collect/)
- 哔哩哔哩[代码](./bilibili)中记录了完整的字段（虽然很多字段也没用上），如有需要可以自行参考



