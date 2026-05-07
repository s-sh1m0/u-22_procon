import { useQuery } from '@tanstack/react-query'
import { apiFetch } from '@/lib/api'
import type { DiffResponse } from '@/types/api'

export function useDiff(jobId: string, enabled: boolean) {
  return useQuery<DiffResponse>({
    queryKey: ['diff', jobId],
    queryFn: () => apiFetch<DiffResponse>(`/api/diff/${jobId}`),
    enabled,
    staleTime: Infinity,
    retry: false,
  })
}
