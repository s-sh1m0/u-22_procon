import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { parsePrUrl } from '@/lib/prUrl'
import { useAnalyzeMutation } from '@/hooks/useAnalysis'
import { ApiError } from '@/lib/api'
import { CheckGlyph, GitHubGlyph } from './glyphs'

/**
 * Logged-in hero CTA: a real PR URL form wired to the analyze mutation.
 * Mirrors the markup of the Claude Design handoff but submits to the backend
 * and navigates to the analysis page on success.
 */
export default function PrInputForm() {
  const [url, setUrl] = useState('')
  const [submitError, setSubmitError] = useState<string | null>(null)
  const navigate = useNavigate()
  const mutation = useAnalyzeMutation()

  const parsed = parsePrUrl(url)
  const valid = parsed !== null
  const empty = url.trim() === ''

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!valid) return
    setSubmitError(null)
    mutation.mutate(url.trim(), {
      onSuccess: (data) => navigate(`/analysis/${data.job_id}`),
      onError: (err) =>
        setSubmitError(err instanceof ApiError ? err.message : '予期しないエラーが発生しました'),
    })
  }

  return (
    <div className="flex max-w-[560px] flex-col gap-3.5">
      <div className="rounded-[14px] border border-stone-200 bg-white p-[18px] shadow-[0_1px_2px_rgba(0,0,0,0.02),0_8px_28px_rgba(15,118,110,0.06)]">
        <div className="mb-2 flex items-center justify-between">
          <label
            htmlFor="pr-url"
            className="font-mono text-[11px] uppercase tracking-[0.04em] text-stone-500"
          >
            Step 1 / 2 · PR URL
          </label>
          {valid && (
            <span className="inline-flex items-center gap-1 text-[11px] font-medium text-green-600">
              <CheckGlyph className="text-green-600" /> 有効な URL
            </span>
          )}
        </div>
        <form onSubmit={handleSubmit}>
          <div className="flex gap-2">
            <div
              className={cn(
                'flex flex-1 items-center gap-2.5 rounded-lg border-[1.5px] bg-stone-50 px-3.5 py-3 transition-[border-color,box-shadow]',
                valid ? 'border-teal-700 ring-4 ring-teal-700/10' : 'border-stone-300',
              )}
            >
              <GitHubGlyph size={15} className="text-stone-500" />
              <Input
                id="pr-url"
                value={url}
                onChange={(e) => {
                  setUrl(e.target.value)
                  setSubmitError(null)
                }}
                placeholder="https://github.com/owner/repo/pull/123"
                className="h-auto border-0 bg-transparent p-0 font-mono text-[13px] text-stone-900 shadow-none focus-visible:ring-0"
              />
            </div>
            <Button
              type="submit"
              disabled={!valid || mutation.isPending}
              className="h-12 bg-stone-900 px-[18px] text-sm text-white hover:bg-stone-800"
            >
              {mutation.isPending ? '送信中…' : '解析を開始 →'}
            </Button>
          </div>

          {!empty && !valid && (
            <div className="mt-2.5 text-xs text-red-600">
              有効な GitHub PR URL を入力してください
            </div>
          )}
          {submitError && (
            <div className="mt-2.5 rounded-lg border border-red-300 bg-red-100 px-3 py-2 text-xs text-red-600">
              {submitError}
            </div>
          )}
          {valid && parsed && (
            <div className="mt-3 grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 rounded-lg bg-stone-100 px-3.5 py-2.5 text-xs text-stone-700">
              <span className="text-stone-500">リポジトリ</span>
              <span className="text-right font-mono">
                {parsed.owner}/{parsed.repo}
              </span>
              <span className="text-stone-500">PR</span>
              <span className="text-right font-mono">#{parsed.number}</span>
            </div>
          )}
        </form>
      </div>
    </div>
  )
}
