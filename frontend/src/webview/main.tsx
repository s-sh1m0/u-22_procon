import '@xyflow/react/dist/style.css'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from '@/lib/queryClient'
import { setApiTransport } from '@/lib/api'
import { elkWorkerUrl, setElkWorkerFactory } from '@/lib/graphLayout'
import { EditorBridgeContext, type EditorBridge } from '@/lib/editorBridge'
import Analysis from '@/pages/Analysis'
import WebviewHome from './WebviewHome'
import { createHostTransport, type VsCodeApi } from './transport'
import type { WebviewConfig } from './protocol'
import '@/index.css'
import './webview.css'

// VS Code 拡張の webview 用エントリ。Web の解析画面（pages/Analysis 以下）をそのまま描画し、
// 通信とエディタ連携だけを VS Code 向けに差し替える。

declare global {
  function acquireVsCodeApi(): VsCodeApi
}

function readConfig(): WebviewConfig {
  const el = document.getElementById('diffgraph-config')
  return JSON.parse(el?.textContent ?? '{}') as WebviewConfig
}

async function boot() {
  const config = readConfig()
  const vscode = acquireVsCodeApi()
  setApiTransport(createHostTransport(vscode))

  // webview は拡張フォルダの URL から直接 Worker を作れない（blob: / data: のみ可）ため、
  // ELK の worker スクリプトを取得して Blob URL から生成する。
  const workerSource = await fetch(elkWorkerUrl).then((res) => res.text())
  const workerBlobUrl = URL.createObjectURL(new Blob([workerSource], { type: 'text/javascript' }))
  setElkWorkerFactory(() => new Worker(workerBlobUrl))

  const bridge: EditorBridge = {
    openDiff: (request) => vscode.postMessage({ type: 'openDiff', request }),
  }

  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <MemoryRouter initialEntries={[`/analysis/${encodeURIComponent(config.jobId)}`]}>
        <QueryClientProvider client={queryClient}>
          <EditorBridgeContext.Provider value={bridge}>
            <Routes>
              <Route path="/analysis/:jobId" element={<Analysis />} />
              <Route path="*" element={<WebviewHome />} />
            </Routes>
          </EditorBridgeContext.Provider>
        </QueryClientProvider>
      </MemoryRouter>
    </StrictMode>,
  )
}

boot().catch((err: unknown) => {
  const root = document.getElementById('root')
  if (root) root.textContent = `DiffGraph の画面を初期化できませんでした: ${String(err)}`
})
