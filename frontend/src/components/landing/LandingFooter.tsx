import Brand from '@/components/branding/Brand'
import { Button } from '@/components/ui/button'
import Wrap from './Wrap'
import { GitHubGlyph } from './glyphs'
import { REPO_URL } from './constants'

type Link = { label: string; href?: string }
type Col = { title: string; links: Link[] }

const COLS: Col[] = [
  {
    title: 'プロダクト',
    links: [
      { label: '機能', href: '#features' },
      { label: '仕組み', href: '#how' },
      { label: '料金' },
      { label: 'ロードマップ' },
      { label: '変更履歴' },
    ],
  },
  {
    title: 'ドキュメント',
    links: [
      { label: 'はじめに' },
      { label: 'API リファレンス' },
      { label: 'セルフホスト' },
      { label: 'CLI' },
    ],
  },
  {
    title: '会社',
    links: [
      { label: 'About' },
      { label: 'ブログ' },
      { label: 'お問い合わせ' },
      { label: 'プレスキット' },
    ],
  },
]

function FooterLink({ link }: { link: Link }) {
  if (link.href) {
    return (
      <a href={link.href} className="text-[13px] text-stone-500 hover:text-stone-700">
        {link.label}
      </a>
    )
  }
  return <span className="text-[13px] text-stone-500">{link.label}</span>
}

export default function LandingFooter() {
  return (
    <footer className="border-t border-stone-200 bg-white pt-18 pb-10">
      <Wrap className="grid grid-cols-1 gap-8 sm:grid-cols-2 lg:grid-cols-[1.4fr_repeat(3,1fr)] lg:gap-12">
        <div>
          <Brand size={16} />
          <p className="my-3.5 mb-5 max-w-[280px] text-[13px] leading-relaxed text-stone-500">
            PR を変更行ではなく依存グラフとして読むためのレビュー支援ツール。
          </p>
          <Button asChild variant="outline" size="sm" className="border-stone-300 text-stone-900">
            <a href={REPO_URL} target="_blank" rel="noreferrer">
              <GitHubGlyph size={12} />
              GitHub
            </a>
          </Button>
        </div>
        {COLS.map((c) => (
          <div key={c.title}>
            <div className="mb-3 text-xs font-semibold text-stone-700">{c.title}</div>
            <ul className="m-0 flex list-none flex-col gap-2 p-0">
              {c.links.map((l) => (
                <li key={l.label}>
                  <FooterLink link={l} />
                </li>
              ))}
            </ul>
          </div>
        ))}
      </Wrap>
      <Wrap className="mt-14 flex flex-wrap justify-between gap-3 border-t border-stone-200 pt-6 text-xs text-stone-500">
        <span>© 2026 diffmap. — u-22 プログラミングコンテスト</span>
        <span className="font-mono">v0.4.2-beta</span>
      </Wrap>
    </footer>
  )
}
