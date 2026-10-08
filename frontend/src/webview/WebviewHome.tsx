import { Card, CardContent } from '@/components/ui/card'

/**
 * webview 内で解析画面以外に遷移したとき（エラー画面の「別の PR を試す」等）の案内。
 * PR の入力は VS Code のコマンドパレットから行うので、ここでは手順だけを示す。
 */
export default function WebviewHome() {
  return (
    <div className="flex min-h-svh items-center justify-center bg-[#fafaf9] px-4">
      <Card className="w-full max-w-md border-stone-200">
        <CardContent className="space-y-2 pt-6">
          <p className="text-sm font-semibold text-stone-800">別の PR を解析するには</p>
          <p className="text-sm text-stone-600">
            コマンドパレット（Ctrl+Shift+P / Cmd+Shift+P）から「DiffGraph: PR を解析」を実行し、PR
            の URL を入力してください。
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
