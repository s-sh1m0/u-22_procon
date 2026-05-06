import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { GitBranch } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import Header from '@/components/layout/Header'
import { parsePrUrl } from '@/lib/prUrl'
import { useAnalyzeMutation } from '@/hooks/useAnalysis'
import { ApiError } from '@/lib/api'

export default function Home() {
  const [url, setUrl] = useState('')
  const [submitError, setSubmitError] = useState<string | null>(null)
  const navigate = useNavigate()
  const mutation = useAnalyzeMutation()

  const parsed = parsePrUrl(url)
  const isValid = parsed !== null
  const isEmpty = url.trim() === ''

  const urlError =
    !isEmpty && !isValid
      ? '有効な GitHub PR URL を入力してください（例: https://github.com/owner/repo/pull/123）'
      : null

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!isValid) return
    setSubmitError(null)
    mutation.mutate(url.trim(), {
      onSuccess: (data) => navigate(`/analysis/${data.job_id}`),
      onError: (err) => {
        if (err instanceof ApiError) {
          setSubmitError(err.message)
        } else {
          setSubmitError('予期しないエラーが発生しました')
        }
      },
    })
  }

  return (
    <div className="min-h-svh bg-[#fafaf9]">
      <Header />
      <main className="mx-auto max-w-2xl px-6 py-16">
        <p className="mb-2 font-mono text-xs text-stone-400 tracking-widest">STEP 1 / 2</p>
        <h1 className="mb-2 text-3xl font-semibold tracking-tight text-stone-900">
          解析するプルリクエストを指定
        </h1>
        <p className="mb-8 text-sm text-stone-500">
          GitHub の PR URL を貼り付けると、リポジトリをクローンし、依存解析を始めます。
        </p>

        <Card className="border-stone-200 shadow-sm">
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-stone-600">PR URL</CardTitle>
            <CardDescription className="sr-only">GitHub PR の URL を入力</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit}>
              <div className="flex gap-2">
                <div className="relative flex-1">
                  <Input
                    value={url}
                    onChange={(e) => {
                      setUrl(e.target.value)
                      setSubmitError(null)
                    }}
                    placeholder="https://github.com/owner/repo/pull/123"
                    className={[
                      'font-mono text-sm pr-16',
                      isValid ? 'border-teal-600 focus-visible:ring-teal-600/30' : '',
                    ].join(' ')}
                    aria-invalid={!!urlError}
                    autoFocus
                  />
                  {isValid && (
                    <span className="absolute right-3 top-1/2 -translate-y-1/2 text-xs font-medium text-teal-600">
                      ✓ 有効
                    </span>
                  )}
                </div>
                <Button
                  type="submit"
                  disabled={!isValid || mutation.isPending}
                  className="shrink-0 bg-stone-900 hover:bg-stone-700"
                >
                  {mutation.isPending ? '送信中…' : '解析を開始'}
                </Button>
              </div>

              {urlError && <p className="mt-2 text-xs text-red-500">{urlError}</p>}
              {submitError && (
                <p className="mt-2 rounded-md bg-red-50 px-3 py-2 text-xs text-red-600 border border-red-200">
                  {submitError}
                </p>
              )}
            </form>

            {isValid && parsed && (
              <div className="mt-4 rounded-lg bg-stone-50 border border-stone-100 px-4 py-3 text-xs text-stone-600 space-y-1">
                <div className="flex justify-between">
                  <span className="text-stone-400">リポジトリ</span>
                  <span className="font-mono">
                    {parsed.owner}/{parsed.repo}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-stone-400">PR #</span>
                  <span className="font-mono flex items-center gap-1">
                    <GitBranch className="size-3" />
                    {parsed.number}
                  </span>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </main>
    </div>
  )
}
