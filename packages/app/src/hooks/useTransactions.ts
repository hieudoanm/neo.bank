import { useInfiniteQuery } from '@tanstack/react-query'
import { transactionsApi } from '@/lib/api/client'
import { useAuth } from '@/context/AuthContext'

export const TRANSACTIONS_PAGE_SIZE = 20

export const useTransactions = (limit: number = TRANSACTIONS_PAGE_SIZE) => {
  const { state } = useAuth()
  const accountId = state.user?.accountId ?? ''
  const accessToken = state.accessToken ?? ''

  return useInfiniteQuery({
    queryKey: ['transactions', accountId, limit],
    queryFn: ({ pageParam }) =>
      transactionsApi.list(accountId, accessToken, pageParam ?? undefined, limit),
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) => (lastPage.hasMore ? lastPage.nextCursor : null),
    enabled: Boolean(accountId) && Boolean(accessToken),
  })
}
