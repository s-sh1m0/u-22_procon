import { useQuery } from '@tanstack/react-query'
import { apiFetch } from '@/lib/api'
import type { ClusterMode, GraphResponse } from '@/types/api'

export function useGraph(jobId: string, enabled: boolean, mode: ClusterMode = 'louvain') {
  return useQuery<GraphResponse>({
    queryKey: ['graph', jobId, mode],
    queryFn: () =>
      apiFetch<GraphResponse>(`/api/graph/${jobId}?cluster=${encodeURIComponent(mode)}`),
    enabled,
    staleTime: Infinity,
    retry: false,
  })
}
