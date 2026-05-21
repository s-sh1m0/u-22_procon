import { useParams, Link } from 'react-router-dom'
import { useJobStatus } from '@/hooks/useJobStatus'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import AnalysisGraphView from '@/pages/AnalysisGraphView'
import type { JobPhase } from '@/types/api'

// STEPS と PHASE_ORDER は同じ並び・同じ長さ。各ステップがバックエンドの 1 フェーズに対応する。
const STEPS = [
  'リポジトリのクローン',
  '依存グラフの構築 (AST 解析)',
  '差分との突き合わせ',
  'クラスタ分類（4-way）',
  '可視化の生成',
] as const

const PHASE_ORDER: JobPhase[] = ['clone', 'build_graph', 'diff', 'cluster', 'visualize']

type StepStatus = 'done' | 'active' | 'pending'

// stepStatuses はジョブの実フェーズから各ステップの表示状態を導出する（時間ベースの推測はしない）。
function stepStatuses(jobStatus: string | undefined, phase: JobPhase | undefined): StepStatus[] {
  if (jobStatus === 'done') return STEPS.map(() => 'done')
  if (jobStatus === 'error') return STEPS.map(() => 'pending')

  // pending（キュー待ち）でフェーズ未設定なら全て pending。それ以外は現在フェーズまでを done/active。
  const current = phase ? PHASE_ORDER.indexOf(phase) : -1
  return STEPS.map((_, i) => {
    if (current < 0) return 'pending'
    if (i < current) return 'done'
    if (i === current) return 'active'
    return 'pending'
  })
}

function isNoPackagesError(msg: string | undefined) {
  return msg?.includes('no Go packages found') || msg?.includes('ErrNoPackages')
}

export default function Analysis() {
  const { jobId } = useParams<{ jobId: string }>()
  const { data: job, isLoading } = useJobStatus(jobId ?? '')
  const steps = stepStatuses(job?.status, job?.phase)

  if (isLoading || !job) {
    return (
      <div className="flex min-h-svh items-center justify-center bg-[#fafaf9]">
        <Spinner label="ジョブを読み込み中…" />
      </div>
    )
  }

  if (job.status === 'error') {
    const isNonGo = isNoPackagesError(job.error)
    return (
      <div className="flex min-h-svh items-center justify-center bg-[#fafaf9]">
        <Card className="w-full max-w-md border-red-200">
          <CardContent className="pt-6 space-y-3">
            <div className="flex items-center gap-2 text-red-600 font-semibold">
              <ErrorIcon />
              <span>解析に失敗しました</span>
            </div>
            <p className="text-sm text-stone-600">
              {isNonGo
                ? 'Go コードを含む PR ではありません。Go リポジトリの PR URL を指定してください。'
                : job.error}
            </p>
            <Button asChild variant="outline" size="sm" className="mt-2">
              <Link to="/">別の PR を試す</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (job.status === 'done') {
    return <AnalysisGraphView jobId={jobId ?? ''} />
  }

  return (
    <div className="flex min-h-svh items-center justify-center bg-[#fafaf9]">
      <div className="w-full max-w-sm space-y-5 px-4">
        <div className="flex items-center gap-4">
          <div className="flex size-11 items-center justify-center rounded-xl bg-teal-50 text-teal-600">
            <SpinnerIcon />
          </div>
          <div>
            <p className="text-lg font-semibold text-stone-900">解析中…</p>
            <p className="text-sm text-stone-400">依存関係を計算しています</p>
          </div>
        </div>

        <Card className="border-stone-200">
          <CardContent className="px-0 py-1">
            {STEPS.map((label, i) => {
              const s = steps[i]
              return (
                <div
                  key={label}
                  className={[
                    'flex items-center gap-3 px-4 py-2.5',
                    i < STEPS.length - 1 ? 'border-b border-stone-100' : '',
                  ].join(' ')}
                >
                  <StepIcon status={s} />
                  <span
                    className={[
                      'flex-1 text-sm',
                      s === 'pending' ? 'text-stone-400' : 'text-stone-800',
                      s === 'active' ? 'font-medium' : '',
                    ].join(' ')}
                  >
                    {label}
                  </span>
                  {s === 'done' && <span className="font-mono text-xs text-stone-400">done</span>}
                  {s === 'active' && <span className="font-mono text-xs text-teal-500">...</span>}
                </div>
              )
            })}
          </CardContent>
        </Card>

        <p className="font-mono text-xs text-stone-400">
          <span className="text-teal-500">►</span> job: {jobId}
        </p>
      </div>
    </div>
  )
}

function StepIcon({ status }: { status: StepStatus }) {
  if (status === 'done') {
    return (
      <span className="flex size-5 items-center justify-center rounded-full bg-teal-600 text-white">
        <svg
          width="10"
          height="10"
          viewBox="0 0 10 10"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
        >
          <path d="M2 5l2 2 4-4" />
        </svg>
      </span>
    )
  }
  if (status === 'active') {
    return (
      <span className="flex size-5 items-center justify-center rounded-full border-2 border-teal-500 bg-teal-50">
        <span className="size-2 rounded-full bg-teal-500" />
      </span>
    )
  }
  return <span className="size-5 rounded-full bg-stone-100" />
}

function Spinner({ label }: { label: string }) {
  return (
    <div className="flex items-center gap-3 text-stone-400">
      <SpinnerIcon />
      <span className="text-sm">{label}</span>
    </div>
  )
}

function SpinnerIcon() {
  return (
    <svg
      width="20"
      height="20"
      viewBox="0 0 20 20"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      className="animate-spin"
    >
      <circle cx="10" cy="10" r="7" strokeDasharray="32" strokeDashoffset="8" />
    </svg>
  )
}

function ErrorIcon() {
  return (
    <svg
      width="18"
      height="18"
      viewBox="0 0 18 18"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
    >
      <circle cx="9" cy="9" r="7" />
      <path d="M9 6v4M9 12.5v.5" strokeLinecap="round" />
    </svg>
  )
}
