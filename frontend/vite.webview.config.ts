import path from 'node:path'
import { defineConfig, mergeConfig } from 'vite'
import baseConfig from './vite.config'

// VS Code 拡張の webview 用ビルド（解析画面のみ。エントリは src/webview/main.tsx）。
// 拡張が HTML を生成して webview.js / webview.css を読み込むので、エントリ名を固定する。
// 成果物は拡張フォルダ配下に置かれるため、アセット参照は相対パス（base: './'）にする。
export default mergeConfig(
  baseConfig,
  defineConfig({
    base: './',
    publicDir: false,
    build: {
      outDir: 'dist-webview',
      emptyOutDir: true,
      cssCodeSplit: false,
      rolldownOptions: {
        input: path.resolve(__dirname, 'src/webview/main.tsx'),
        output: {
          entryFileNames: 'webview.js',
          chunkFileNames: 'assets/[name]-[hash].js',
          assetFileNames: (asset) =>
            asset.names.some((n) => n.endsWith('.css'))
              ? 'webview.css'
              : 'assets/[name]-[hash][extname]',
        },
      },
    },
  }),
)
