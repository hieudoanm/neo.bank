import { useQuery } from '@tanstack/react-query'
import { transfersApi } from '@/lib/api/client'
import { useAuth } from '@/context/AuthContext'

const POLL_INTERVAL_MS = 2000

export const useTransfer = (transferId: string) => {
  const { state } = useAuth()
  const accessToken = state.accessToken ?? ''

  return useQuery({
    queryKey: ['transfer', transferId],
    queryFn: () => transfersApi.get(transferId, accessToken),
    enabled: Boolean(transferId) && Boolean(accessToken),
    refetchInterval: (query) => {
      const status = query.state.data?.status
      return status === 'completed' || status === 'failed' || status === 'reversed'
        ? false
        : POLL_INTERVAL_MS
    },
  })
}
