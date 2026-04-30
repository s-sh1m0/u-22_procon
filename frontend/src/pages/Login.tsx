import { Navigate, useLocation } from 'react-router-dom'
import { buttonVariants } from '@/components/ui/button'
import { useAuth } from '@/hooks/useAuth'

export default function Login() {
  const { isAuthenticated, isLoading } = useAuth()
  const location = useLocation()
  const from = (location.state as { from?: string } | null)?.from ?? '/'

  if (isLoading) return <div className="p-8 text-muted-foreground">読み込み中…</div>
  if (isAuthenticated) return <Navigate to={from} replace />

  return (
    <div className="flex min-h-svh items-center justify-center">
      <div className="w-full max-w-sm space-y-6 rounded-xl border bg-card p-8 shadow-sm">
        <div className="space-y-1 text-center">
          <h1 className="text-2xl font-bold">ReviewArena</h1>
          <p className="text-sm text-muted-foreground">ログインして PR 依存グラフを可視化</p>
        </div>
        <a
          href="/auth/github"
          className={buttonVariants({ variant: 'default', size: 'lg' }) + ' w-full justify-center'}
        >
          GitHub でログイン
        </a>
      </div>
    </div>
  )
}
