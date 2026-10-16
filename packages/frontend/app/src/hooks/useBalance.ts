import { useQuery } from '@tanstack/react-query'
import { accountsApi } from '@/lib/api/client'
import { useAuth } from '@/context/AuthContext'

export const useBalance = () => {
  const { state } = useAuth()
  const accountId = state.user?.accountId ?? ''
  const accessToken = state.accessToken ?? ''

  return useQuery({
    queryKey: ['balance', accountId],
    queryFn: () => accountsApi.getBalance(accountId, accessToken),
    enabled: Boolean(accountId) && Boolean(accessToken),
    staleTime: 10_000,
    refetchInterval: 30_000,
  })
}
