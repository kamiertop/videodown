# Third-party source code

## kkdai/youtube

- Repository: https://github.com/kkdai/youtube
- Revision: `3876742b95fd1cd96f836f396c2d54be8912a297`
- Local directory: `youtube/internal/extractor/`
- License: MIT; full copyright and permission notice in `youtube/internal/extractor/LICENSE`.
- Source files: client.go, video.go, video_id.go, response_data.go, format_list.go,
  player_cache.go, player_parse.go, decipher.go, decipher_operations.go, protobuilder.go, errors.go.
  playlist.go adapts the upstream renderer paths and continuation protocol using encoding/json,
  explicit pagination and current WEB lockupViewModel support.
- Changes: removed transcript and upstream download machinery, CLI and debug artifacts;
  scoped optional Cookie headers to YouTube; bounded metadata responses and script execution;
  removed age-restriction fallback and the unnecessary watch-page fetch after successful metadata parsing.
- The application uses its existing Goja dependency. No yt-dlp executable or Python runtime is bundled.

## YouTube client protocol references

The visionOS client name/version, user-agent and OS fields were checked against
https://github.com/yt-dlp/yt-dlp/blob/master/yt_dlp/extractor/youtube/_base.py
on 2026-10-29. These replace the upstream Go library's obsolete Android VR default.
Application changes also keep visitor identifiers consistent between request headers
and bodies, propagate cancellation into visitor fetches, and correct the upstream
`n` query-parameter handling. The application downloads media itself and uses the
existing FFmpeg executable only for merging.

The SID authorization protocol and authenticated TV client fields are also based
on yt-dlp's `_get_sid_authorization_header` and `INNERTUBE_CLIENTS` documentation
in the same source file. The Go implementation adds origin-scoped session
signatures, authenticated player timestamps, and a public anonymous fallback.

## yt-dlp EJS player solver

- Repository: https://github.com/yt-dlp/ejs
- Revision: `d48c6303a5a6455a42ef9e0b445149cadea51279`
- Entry point: `src/yt/solver/main.ts`; the bundled solver is embedded in `youtube/internal/extractor/ejs/solver.js`.
- License: Unlicense, preserved in `youtube/internal/extractor/ejs/LICENSE`.
- Bundled dependencies: meriyah 6.1.4 (ISC) and astring 1.9.0 (MIT); full notices in `MERIYAH_LICENSE` and `ASTRING_LICENSE` beside the bundle.
- Built with esbuild 0.28.1; `ejs/build.cjs` is the reproducible build entry point. The application executes this script with its existing Goja runtime, with a hard timeout and no host filesystem/network bindings. Node is needed only to rebuild the bundle.
- TV client device fields follow yt-dlp PRs https://github.com/yt-dlp/yt-dlp/pull/17723 and https://github.com/yt-dlp/yt-dlp/pull/17805.

## YouTube audio rendition identity

The `isDrc` flag and distinction between stream renditions were checked against
[yt-dlp's YouTube format extraction](https://github.com/yt-dlp/yt-dlp/blob/master/yt_dlp/extractor/youtube/_video.py)
on 2026-10-30. The Go picker filters duplicate default-language renditions, gives
DRC its own stable selection ID, and uses the same selection rules for the UI,
audio downloads and video/audio merging.
