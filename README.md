# VideoDown

> ⚠️下载哔哩哔哩视频请安装FFmpeg

VideoDown 是一个使用 [Go](https://go.dev)、[MyGo](https://github.com/egoist/mygo) 和
[MujicaUI](https://github.com/ELY-Open-Source/MujicaUI) 开发的原生桌面视频下载工具，支持哔哩哔哩和抖音。
界面与业务逻辑均使用 Go 编写，构建不需要 CGO、Node.js 或浏览器运行时。

> 本项目仍可能受到平台接口变化影响。如果接口返回结构、鉴权参数或风控策略变化，部分功能可能需要同步调整。

## 项目初衷🤓

每逢月末流量不够😭

大家自行使用😊，🫵 **不要贩卖**😡

## 软件预览🥳

![](./docs/home.png)

### 设置

| 命名模板                                     | 设置                                     |
|----------------------------------------------|------------------------------------------|
| ![命名模板](./docs/setting_tmpl.png)         | ![设置](./docs/setting_general.png)      |
| 哔哩哔哩设置                                 | 下载设置                                 |
| ![哔哩哔哩设置](./docs/setting_bilibili.png) | ![下载设置](./docs/setting_download.png) |

### 哔哩哔哩

| UP 主页面                               | 下载页                                                 |
|-----------------------------------------|--------------------------------------------------------|
| ![UP 主页面](./docs/up.png)             | ![B 站下载页](./docs/bilibili_download.png)            |
| 收藏夹                                  | 收藏夹合集                                             |
| ![B 站收藏夹](./docs/bilibili_star.png) | ![B 站收藏夹合集](./docs/bilibili_star_collection.png) |
| 关注动态                                | 播放历史                                               |
| ![关注动态](./docs/bilibili_dyn.png)    | ![播放历史](./docs/bilibili_play_history.png)          |

### 抖音

| 用户页                                | 下载页                                           |
|---------------------------------------|--------------------------------------------------|
| ![抖音用户页](./docs/douyin_user.png) | ![抖音下载页](./docs/douyin_download.png)        |
| 收藏页                                | 用户作品页                                       |
| ![抖音收藏页](./docs/douyin_star.png) | ![抖音用户作品页](./docs/douyin_user_detail.png) |
| 关注动态                              | 播放历史                                         |
| ![关注动态](./docs/douyin_dyn.png)    | ![播放历史](./docs/douyin_play_history.png)      |

## 功能特性🎉

### 平台支持

- [x] Windows
- [x] Linux
- [ ] macOS（原生构建，实际桌面行为仍需在设备上验证）

### 哔哩哔哩

- 二维码登录和账号状态展示
- 关注 UP 主列表
- UP 主稿件、合集、系列、分 P 浏览
- 视频链接或 BV 号解析
- 收藏夹和收藏夹内合集浏览
- 下载队列、清晰度/音轨选择、下载进度
- 批量下载收藏夹、合集、系列、UP 主视频
- 下载历史查看和删除
- 检测登录状态是否过期
- 播放历史
- 关注 UP 主的动态

### 抖音

- Cookie 登录和账号信息展示
- 用户搜索、用户作品列表、用户合集
- 收藏视频、收藏合集浏览
- 分享链接、分享文案或视频 ID 解析
- 普通视频和图片合集识别
- 多清晰度选项展示
- 下载队列、批量下载、下载进度
- 下载来源区分：收藏视频、收藏合集、用户作品、用户合集、手动解析
- 下载历史查看和删除
- 播放历史
- 关注博主的动态

## 技术栈

- 原生界面与窗口：MyGo、MujicaUI
- 业务逻辑：Go、req/v3
- 配置与 Cookie：结构化 TOML
- 下载历史：SQLite、sqlx、[glebarez/go-sqlite](https://github.com/glebarez/go-sqlite) 纯 Go 驱动（无需 CGO）

## 开发

需要 Go 1.27.3 或更新版本。B 站视频通常需要合并视频流和音频流，请安装 FFmpeg，并在设置中选择路径或确保其在 PATH 中。

启动应用：

```sh
just run
# 或构建便携目录后启动
CGO_ENABLED=0 go run ./cmd/package
./dist/videodown/videodown
```

运行测试和原生 UI 检查：

```sh
CGO_ENABLED=0 go test ./...
CGO_ENABLED=0 go tool mygo vet .
```

构建应用：

```sh
CGO_ENABLED=0 go build -o build/bin/videodown .
# Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-H=windowsgui" -o build/bin/videodown.exe .
```

使用 `just package` 或 `go run ./cmd/package` 构建便携文件夹 `dist/videodown`，其中包含主程序、独立更新器和 `appicon.png`；Windows 打包会同时将图标写入 EXE 资源，供资源管理器显示。可通过 `-os`、`-arch`、`-version` 和 `-out` 指定目标与输出位置。`just run` 启动开发版本，`just check` 执行检查。界面代码位于 `internal/gui`，启动与服务初始化位于 `main.go`。
发布工作流使用原生 Go 构建并生成便携归档。启动时遵循设置中的自动更新开关，在后台检查新版本；发现更新后展示版本和发布说明，可打开对应发布页。检查失败不会阻塞应用。

## 使用说明

便携文件夹首次运行后的结构：

```text
videodown/
  videodown                 # Windows 为 videodown.exe
  videodown-updater         # Windows 带 .exe 后缀
  appicon.png               # 图标同时内嵌在程序中
  config.toml               # 首次启动自动创建，含账号凭证，请勿公开
  history.sqlite            # 首次启动自动创建
  download/                 # 下载时创建
  tmp/                      # 应用及子进程临时文件
  cache/                    # 界面库等组件的缓存
  videodown.log             # 生产模式日志
  videodown-updater.log     # 更新器日志
```

应用自身文件以主程序实际存放目录为根：配置、历史、图标、日志、更新器、更新包和更新解压文件均位于该目录，默认下载与临时文件分别位于其子目录；视频下载目录允许用户修改。启动不依赖当前工作目录，也不提供 `-data-dir` 覆盖。开发请使用 `just run` 构建并启动便携目录；直接 `go run .` 的数据会位于 Go 的临时二进制目录中。目录不可写时直接报错，不回退到用户目录。

### 基础设置

1. 打开“设置”页面，按需修改下载目录。
2. 根据需要设置下载后休眠时间、并发数量和保存时是否自动分组。
3. 确认 `ffmpeg` 已安装，尤其是下载 B 站 DASH 视频时。

### 哔哩哔哩

1. 进入 B 站账号页，使用二维码登录。
2. 可以从关注 UP 主、收藏夹、合集、系列或下载页手动解析视频。
3. 将视频加入下载队列后，在下载页选择画质和音轨。
4. 点击单个下载或批量下载。

### 抖音

1. 进入抖音账号页，粘贴浏览器中的完整 Cookie。
2. Cookie 建议至少包含 `sessionid`、`ttwid`、`UIFID`、`s_v_web_id`。
3. 可以从收藏、用户页、合集页加入下载队列，也可以在下载页粘贴分享链接手动解析。
4. 如果视频是图片合集，下载时会保存为一个目录，图片按序号落盘。
5. 如果接口返回多个视频清晰度，可以在下载卡片中选择对应选项。

### YouTube 下载

支持粘贴公开视频或 Shorts 链接加入待下载区；选择画质或音质后手动开始，已启动任务按公共下载并发数执行，
可在下载期间继续解析视频，支持暂停后续任务、取消、失败重试与清理已结束任务。
在下载页点击卡片右侧“音频”切换为仅音频，可直接保存原始音轨（M4A / WebM），不进行转码；音频与视频分别记录下载历史。
播放列表页支持粘贴包含 `list=` 的链接、点击卡片高亮选择、全选、移除及加载更多，再批量加入待下载区；
“全选”针对当前已加载的视频，支持移除已选或一键清空解析列表；这些操作不影响下载页中的任务。加入队列不会自动下载；清晰度及视频／仅音频设置在下载页选择，支持逐条下载或整组开始。列表来源会标记任务状态。
队列按播放列表主题分组，默认折叠，可展开选择格式、单条移除或整组移除；移除会取消未完成任务，保留已下载文件。
列表与队列封面按可见区域加载并共享缓存。队列保留在本次运行中，重启后可从历史查看已完成文件。
YouTube 左侧菜单提供下载、播放列表、下载历史和账号页；在账号页保存浏览器单行 Cookie 请求头，留空保存可清除；
也可在 `config.toml` 的 `[youtube]` 下设置 `cookie`。
分离音视频会通过现有 FFmpeg 合并为 MKV；已有音频的视频格式直接保存。
服务器可能根据账号和网络限制部分格式；Cookie 不保证消除播放器或媒体请求的限制。
本功能不携带 yt-dlp 或 Python。移植代码、上游版本与许可证见 [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)。

## 数据和文件

> 本项目绝对不会私自收集、上传用户信息

- 配置和 Cookie 存储在可执行文件同目录的 `config.toml` 中，下载历史存储在同目录的 `history.sqlite` 中。首次启动自动创建缺失的配置、数据库和图标，已有配置会保留。配置结构参考 [config.example.toml](config.example.toml)。
- 默认下载目录为程序文件夹中的 `download/`；相对路径以配置文件所在目录为基准，整体移动文件夹后仍然有效。允许在设置页修改为其他位置，包括程序目录之外的文件夹。
- 启用自动分组后，后端会按来源类型创建目录，例如收藏合集、用户合集、收藏视频或用户作品。

## 注意事项

- 平台接口和风控策略可能随时变化，Cookie、二维码登录或解析接口可能失效。
- 抖音接口返回字段非常庞大，当前模型只保留了项目使用到的关键字段。如需二次开发，可以参考 `douyin/model` 和 `douyin/api`。
- B 站相关模型保留字段较多，方便后续扩展。
- 请不要将账号 Cookie 分享给他人，也不要提交到公开仓库。

## 贡献指南

- 欢迎提交 Pull Request 或 Issue 讨论功能需求、bug 修复或代码优化。
- 邮箱联系我：kamiertop@gmail.com

## 免责声明⚠️

本项目仅用于个人学习、研究和技术交流。请遵守相关平台服务条款、版权法律法规和内容所有者授权要求。使用本项目产生的任何风险和后果由使用者自行承担。

请勿将本项目用于商业用途、批量抓取、侵犯版权、绕过平台限制或其他违法违规行为。

## License

许可证暂未最终确定。

如果你计划复制、修改、分发或二次开发本项目，请先确认仓库中的最终许可证文件和第三方依赖许可证。当前阶段不建议在未确认许可证的情况下进行商业使用或再分发。

## 致谢🤗

- 视频流说明、接口签名等：[bilibili-API-collect](https://sessionhu.github.io/bilibili-API-collect/)
- aBogus，提示不要使用二维码登录，用Cookie（我听劝）：[TikTokDownloader](https://github.com/JoeanAmier/TikTokDownloader)
- 提示使用Cookie：[Douyin_TikTok_Download_API](https://github.com/Evil0ctal/Douyin_TikTok_Download_API)
- Codex 和 ChatGPT 在代码生成和优化方面提供了巨大帮助
- 借助 Codex-5.6-Sol 和 [js-reverse-mcp](https://github.com/zhizhuodemao/js-reverse-mcp) 实现了 `x-secsdk-web-signature`
