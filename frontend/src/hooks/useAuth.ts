import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { ApiError, apiFetch, apiFetchVoid } from '@/lib/api'

type Me = { login: string }

export function useAuth() {
  const query = useQuery<Me | null>({
    queryKey: ['auth', 'me'],
    queryFn: async () => {
      try {
        return await apiFetch<Me>('/auth/me')
      } catch (e) {
        if (e instanceof ApiError && e.status === 401) return null
        throw e
      }
    },
    staleTime: Infinity,
    retry: false,
  })
  return {
    user: query.data ?? null,
    isLoading: query.isLoading,
    isAuthenticated: !!query.data,
  }
}

export function useLogout() {
  const qc = useQueryClient()
  const navigate = useNavigate()
  return useMutation({
    mutationFn: () => apiFetchVoid('/auth/logout', { method: 'POST' }),
    onSuccess: () => {
      qc.setQueryData(['auth', 'me'], null)
      navigate('/login', { replace: true })
    },
  })
}
