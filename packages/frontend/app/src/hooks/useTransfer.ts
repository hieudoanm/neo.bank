import { useQuery } from '@tanstack/react-query'
import { transfersApi } from '@/lib/api/client'
import { useAuth } from '@/context/AuthContext'

const POLL_INTERVAL_MS = 2000
// Stop polling after ~2 minutes; a saga stuck longer than that needs manual
// investigation rather than an open-ended request loop.
const MAX_POLL_FETCHES = 60

export const useTransfer = (transferId: string) => {
  const { state } = useAuth()
  const accessToken = state.accessToken ?? ''

  return useQuery({
    queryKey: ['transfer', transferId],
    queryFn: () => transfersApi.get(transferId, accessToken),
    enabled: Boolean(transferId) && Boolean(accessToken),
    refetchInterval: (query) => {
      const status = query.state.data?.status
      if (status === 'completed' || status === 'failed' || status === 'reversed') {
        return false
      }
      return query.state.dataUpdateCount >= MAX_POLL_FETCHES ? false : POLL_INTERVAL_MS
    },
  })
}
