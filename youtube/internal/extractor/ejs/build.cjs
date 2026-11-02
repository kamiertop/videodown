// Build-time only: node build.cjs /path/to/yt-dlp-ejs
const path = require('node:path');
require('esbuild').buildSync({
 entryPoints:[path.join(process.argv[2],'src/yt/solver/main.ts')],
 outfile:path.join(__dirname,'solver.js'),
 bundle:true,format:'iife',globalName:'EJSSolver',target:'es2020',minify:true,
 legalComments:'external',nodePaths:(process.env.NODE_PATH||'').split(path.delimiter).filter(Boolean),
});
