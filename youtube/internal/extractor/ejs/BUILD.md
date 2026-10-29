# Rebuilding the embedded solver

Upstream: https://github.com/yt-dlp/ejs
Revision: d48c6303a5a6455a42ef9e0b445149cadea51279
Entry: src/yt/solver/main.ts

From the application repository root, with Node/npm installed for this build only:

```sh
git clone https://github.com/yt-dlp/ejs.git /tmp/videodown-ejs-source
git -C /tmp/videodown-ejs-source checkout d48c6303a5a6455a42ef9e0b445149cadea51279
mkdir -p /tmp/videodown-ejs-build
npm install --prefix /tmp/videodown-ejs-build --ignore-scripts --no-audit --no-fund esbuild@0.28.1 astring@1.9.0 meriyah@6.1.4
NODE_PATH=/tmp/videodown-ejs-build/node_modules node youtube/internal/extractor/ejs/build.cjs /tmp/videodown-ejs-source
```

Keep all license files beside solver.js, including solver.js.LEGAL.txt if the build generates it.
Go embeds the JavaScript bundle; no Node installation is needed to build the Go
application from the checked-in bundle or to run it.
