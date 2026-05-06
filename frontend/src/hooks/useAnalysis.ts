import { useMutation } from '@tanstack/react-query'
import { apiFetch } from '@/lib/api'
import type { JobResponse } from '@/types/api'

export function useAnalyzeMutation() {
  return useMutation({
    mutationFn: (prUrl: string) =>
      apiFetch<JobResponse>('/api/analyze', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ pr_url: prUrl }),
      }),
  })
}
