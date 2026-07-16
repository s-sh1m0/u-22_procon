import Brand from '@/components/branding/Brand'
import { Button } from '@/components/ui/button'
import { API_BASE_URL } from '@/lib/api'
import Wrap from './Wrap'
import { GitHubGlyph } from './glyphs'
import { REPO_URL } from './constants'

type Props = { isAuthenticated: boolean; login?: string | null }

const navLink = 'inline-flex items-center gap-1.5 text-stone-700 hover:text-stone-900'

export default function LandingNav({ isAuthenticated, login }: Props) {
  return (
    <nav className="sticky top-0 z-50 border-b border-stone-200 bg-stone-50/85 backdrop-blur-md">
      <Wrap className="flex h-16 items-center justify-between">
        <Brand size={17} />
        <div className="flex items-center gap-7 text-[13px] text-stone-700">
          <a className={`${navLink} hidden md:inline-flex`} href="#features">
            機能
          </a>
          <a className={`${navLink} hidden md:inline-flex`} href="#how">
            仕組み
          </a>
          <a className={`${navLink} hidden md:inline-flex`} href="#showcase">
            クラスタリング
          </a>
          <a
            className={`${navLink} hidden md:inline-flex`}
            href={REPO_URL}
            target="_blank"
            rel="noreferrer"
          >
            <GitHubGlyph size={13} /> GitHub
          </a>
          <div className="hidden h-[18px] w-px bg-stone-200 md:block" />
          {isAuthenticated ? (
            <div className="flex items-center gap-2.5">
              <span className="font-mono text-xs text-stone-500">{login}</span>
              <div className="flex size-[30px] items-center justify-center rounded-full bg-linear-to-br from-teal-700 to-teal-800 text-[11px] font-semibold text-white uppercase">
                {login?.[0] ?? '?'}
              </div>
            </div>
          ) : (
            <Button asChild className="h-[34px] bg-stone-900 px-3.5 text-[13px] hover:bg-stone-800">
              <a href={`${API_BASE_URL}/auth/github`}>ログイン</a>
            </Button>
          )}
        </div>
      </Wrap>
    </nav>
  )
}
