import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { parsePrUrl } from '@/lib/prUrl'
import { useAnalyzeMutation } from '@/hooks/useAnalysis'
import { ApiError } from '@/lib/api'
import { CheckGlyph, GitHubGlyph } from './glyphs'
import { btn, btnInk } from './styles'

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
    <div style={{ display: 'flex', flexDirection: 'column', gap: 14, maxWidth: 560 }}>
      <div
        style={{
          background: 'white',
          border: '1px solid var(--line)',
          borderRadius: 14,
          padding: 18,
          boxShadow: '0 1px 2px rgba(0,0,0,0.02), 0 8px 28px rgba(15,118,110,0.06)',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            marginBottom: 8,
          }}
        >
          <label
            className="mono"
            htmlFor="pr-url"
            style={{
              fontSize: 11,
              color: 'var(--ink-3)',
              letterSpacing: '0.04em',
              textTransform: 'uppercase',
            }}
          >
            Step 1 / 2 · PR URL
          </label>
          {valid && (
            <span
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 4,
                fontSize: 11,
                color: 'var(--added)',
                fontWeight: 500,
              }}
            >
              <CheckGlyph color="var(--added)" /> 有効な URL
            </span>
          )}
        </div>
        <form onSubmit={handleSubmit}>
          <div style={{ display: 'flex', gap: 8 }}>
            <div
              style={{
                flex: 1,
                display: 'flex',
                alignItems: 'center',
                gap: 10,
                border: `1.5px solid ${valid ? 'var(--brand)' : 'var(--line-strong)'}`,
                borderRadius: 10,
                padding: '12px 14px',
                background: 'var(--bg)',
                transition: 'border-color 100ms ease, box-shadow 100ms ease',
                boxShadow: valid ? '0 0 0 4px rgba(15,118,110,0.08)' : 'none',
              }}
            >
              <GitHubGlyph size={15} color="var(--ink-3)" />
              <input
                id="pr-url"
                value={url}
                onChange={(e) => {
                  setUrl(e.target.value)
                  setSubmitError(null)
                }}
                placeholder="https://github.com/owner/repo/pull/123"
                className="mono"
                style={{
                  flex: 1,
                  border: 'none',
                  outline: 'none',
                  fontSize: 13,
                  color: 'var(--ink)',
                  background: 'transparent',
                }}
              />
            </div>
            <button
              type="submit"
              disabled={!valid || mutation.isPending}
              style={{
                ...btn,
                ...btnInk,
                height: 48,
                padding: '0 18px',
                fontSize: 14,
                opacity: valid && !mutation.isPending ? 1 : 0.45,
                cursor: valid && !mutation.isPending ? 'pointer' : 'not-allowed',
              }}
            >
              {mutation.isPending ? '送信中…' : '解析を開始 →'}
            </button>
          </div>
          {!empty && !valid && (
            <div style={{ marginTop: 10, fontSize: 12, color: 'var(--removed)' }}>
              有効な GitHub PR URL を入力してください
            </div>
          )}
          {submitError && (
            <div
              style={{
                marginTop: 10,
                fontSize: 12,
                color: 'var(--removed)',
                background: 'var(--removed-bg)',
                border: '1px solid var(--removed-bd)',
                borderRadius: 8,
                padding: '8px 12px',
              }}
            >
              {submitError}
            </div>
          )}
          {valid && parsed && (
            <div
              style={{
                marginTop: 12,
                padding: '10px 14px',
                background: 'var(--bg-alt)',
                borderRadius: 8,
                fontSize: 12,
                color: 'var(--ink-2)',
                display: 'grid',
                gridTemplateColumns: 'auto 1fr',
                gap: '4px 16px',
              }}
            >
              <span style={{ color: 'var(--ink-3)' }}>リポジトリ</span>
              <span className="mono" style={{ textAlign: 'right' }}>
                {parsed.owner}/{parsed.repo}
              </span>
              <span style={{ color: 'var(--ink-3)' }}>PR</span>
              <span className="mono" style={{ textAlign: 'right' }}>
                #{parsed.number}
              </span>
            </div>
          )}
        </form>
      </div>
    </div>
  )
}
