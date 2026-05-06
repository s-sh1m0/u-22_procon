import { useQuery } from '@tanstack/react-query'
import { apiFetch } from '@/lib/api'
import type { JobResponse } from '@/types/api'

export function useJobStatus(jobId: string) {
  return useQuery<JobResponse>({
    queryKey: ['jobs', jobId],
    queryFn: () => apiFetch<JobResponse>(`/api/jobs/${jobId}`),
    refetchInterval: (query) => {
      const status = query.state.data?.status
      if (status === 'done' || status === 'error') return false
      return 1500
    },
    enabled: !!jobId,
  })
}
