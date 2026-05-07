import { useQuery } from '@tanstack/react-query'
import { apiFetch } from '@/lib/api'
import type { GraphResponse } from '@/types/api'

export function useGraph(jobId: string, enabled: boolean) {
  return useQuery<GraphResponse>({
    queryKey: ['graph', jobId],
    queryFn: () => apiFetch<GraphResponse>(`/api/graph/${jobId}`),
    enabled,
    staleTime: Infinity,
    retry: false,
  })
}
