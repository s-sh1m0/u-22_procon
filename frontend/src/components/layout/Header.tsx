import { Button } from '@/components/ui/button'
import { useAuth, useLogout } from '@/hooks/useAuth'

export default function Header() {
  const { user } = useAuth()
  const logout = useLogout()

  return (
    <header className="flex h-14 items-center justify-between border-b px-6">
      <span className="font-semibold">DiffGraph</span>
      <div className="flex items-center gap-3">
        {user && <span className="text-sm text-muted-foreground">{user.login}</span>}
        <Button
          variant="outline"
          size="sm"
          onClick={() => logout.mutate()}
          disabled={logout.isPending}
        >
          ログアウト
        </Button>
      </div>
    </header>
  )
}
