import { Button } from '@/components/ui/button'
import { useAuth, useLogout } from '@/hooks/useAuth'
import Brand from '@/components/branding/Brand'
import type { PRInfo } from '@/types/api'

type Props = { pr: PRInfo }

export default function PRMetaBar({ pr }: Props) {
  const { user } = useAuth()
  const logout = useLogout()

  return (
    <header className="flex h-12 flex-shrink-0 items-center gap-3 border-b border-stone-200 bg-white px-4">
      <Brand size={14} />
      <div className="h-5 w-px bg-stone-200" />
      <div className="min-w-0 flex-1">
        <span className="text-sm font-medium text-stone-800">{pr.title || `PR #${pr.number}`}</span>
        <span className="ml-2 font-mono text-xs text-stone-400">
          {pr.owner}/{pr.repo} #{pr.number}
          {pr.head_ref && pr.base_ref && (
            <>
              {' '}
              · {pr.head_ref} → {pr.base_ref}
            </>
          )}
        </span>
      </div>
      <div className="flex items-center gap-2">
        {user && <span className="text-xs text-stone-400">{user.login}</span>}
        <Button
          variant="outline"
          size="sm"
          className="h-7 text-xs"
          onClick={() => logout.mutate()}
          disabled={logout.isPending}
        >
          ログアウト
        </Button>
      </div>
    </header>
  )
}
