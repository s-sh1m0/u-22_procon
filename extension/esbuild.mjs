// 拡張ホスト側コードのバンドルと、webview 用成果物の取り込みを行う。
// webview は frontend 側で `npm run build:webview` した frontend/dist-webview を media/webview にコピーする
// （UI は Web と同じコードからビルドし、拡張では作り直さない）。
import { build } from 'esbuild'
import { cpSync, existsSync, rmSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const webviewSrc = fileURLToPath(new URL('../frontend/dist-webview', import.meta.url))
const webviewDest = fileURLToPath(new URL('./media/webview', import.meta.url))

if (!existsSync(`${webviewSrc}/webview.js`)) {
  console.error(
    'frontend/dist-webview/webview.js がありません。先に frontend で `npm run build:webview` を実行してください。',
  )
  process.exit(1)
}
rmSync(webviewDest, { recursive: true, force: true })
cpSync(webviewSrc, webviewDest, { recursive: true })

await build({
  entryPoints: ['src/extension.ts'],
  bundle: true,
  outfile: 'dist/extension.js',
  external: ['vscode'],
  platform: 'node',
  format: 'cjs',
  target: 'node20',
  sourcemap: false,
  minify: true,
  logLevel: 'info',
})
