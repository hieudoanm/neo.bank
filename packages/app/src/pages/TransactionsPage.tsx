import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTransactions } from '@/hooks/useTransactions'
import { Alert } from '@/components/ui/Alert'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { TransactionRow } from '@/components/transactions/TransactionRow'

const errorStatus = (err: unknown): number | undefined =>
  (err as { status?: number } | undefined)?.status

export const TransactionsPage = () => {
  const navigate = useNavigate()
  const [pageIndex, setPageIndex] = useState(0)
  const {
    data,
    isLoading,
    isError,
    error,
    refetch,
    isFetching,
    fetchNextPage,
    isFetchingNextPage,
  } = useTransactions()

  const pages = data?.pages ?? []
  const page = pages[pageIndex]
  const status = isError ? errorStatus(error) : undefined

  const isAuthError = status === 401
  const isPermissionError = status === 404
  const isValidationError = status === 400

  const handleNext = async () => {
    if (!page) return
    if (pageIndex + 1 < pages.length) {
      setPageIndex((current) => current + 1)
      return
    }
    if (!page.hasMore) return
    await fetchNextPage()
    setPageIndex((current) => current + 1)
  }

  return (
    <div className="space-y-6">
      <header>
        <h1 className="text-text-primary text-2xl font-semibold tracking-tight">
          Transaction history
        </h1>
        <p className="mt-1 text-sm text-text-secondary">Your recent account activity</p>
      </header>

      {isAuthError && (
        <Alert type="error">
          Your session has expired. Please sign in again.
          <div className="mt-2">
            <Button variant="outline" onClick={() => navigate('/login')}>
              Sign in
            </Button>
          </div>
        </Alert>
      )}

      {isPermissionError && <Alert type="error">Account not found.</Alert>}

      {isValidationError && (
        <Alert type="error">
          We couldn&apos;t load these transactions.
          <div className="mt-2">
            <Button
              variant="outline"
              disabled={isFetching}
              onClick={() => {
                setPageIndex(0)
                void refetch()
              }}
            >
              Try again
            </Button>
          </div>
        </Alert>
      )}

      {isError && !isAuthError && !isPermissionError && !isValidationError && (
        <Alert type="error">
          Unable to load your transactions. Please check your connection and try again.
          <div className="mt-2">
            <Button variant="outline" disabled={isFetching} onClick={() => void refetch()}>
              Try again
            </Button>
          </div>
        </Alert>
      )}

      {isLoading && (
        <Card className="overflow-hidden p-0">
          <ul
            aria-busy="true"
            aria-label="Loading transactions"
            className="divide-y divide-border-light"
          >
            {[0, 1, 2, 3, 4].map((key) => (
              <li key={key} className="flex items-center gap-4 px-4 py-3">
                <div className="h-10 w-10 animate-pulse rounded-full bg-background-secondary" />
                <div className="flex-1 space-y-2">
                  <div className="h-3 w-40 animate-pulse rounded bg-background-secondary" />
                  <div className="h-3 w-24 animate-pulse rounded bg-background-secondary" />
                </div>
                <div className="h-3 w-16 animate-pulse rounded bg-background-secondary" />
              </li>
            ))}
          </ul>
        </Card>
      )}

      {!isLoading && !isError && page && page.transactions.length === 0 && (
        <Card className="items-center text-center">
          <p className="text-lg font-semibold text-text-primary">No transactions yet</p>
          <p className="mt-1 text-sm text-text-secondary">
            You haven&apos;t made any transfers. Start by sending money to another account.
          </p>
          <div className="mt-4">
            <Button variant="primary" onClick={() => navigate('/transfer')}>
              Transfer money
            </Button>
          </div>
        </Card>
      )}

      {!isLoading && !isError && page && page.transactions.length > 0 && (
        <Card className="overflow-hidden p-0">
          <ul aria-label="Transactions">
            {page.transactions.map((transaction) => (
              <TransactionRow key={transaction.transactionId} transaction={transaction} />
            ))}
          </ul>
          <div className="flex items-center justify-between border-t border-border px-4 py-3">
            <p className="text-xs text-text-secondary">
              Page {pageIndex + 1} · {page.transactions.length} transaction
              {page.transactions.length === 1 ? '' : 's'}
            </p>
            <div className="flex gap-2">
              <Button
                variant="outline"
                disabled={pageIndex === 0 || isFetchingNextPage}
                onClick={() => setPageIndex((current) => Math.max(0, current - 1))}
              >
                Previous
              </Button>
              <Button
                variant="outline"
                disabled={!page.hasMore || isFetchingNextPage}
                onClick={() => void handleNext()}
              >
                {isFetchingNextPage ? 'Loading…' : 'Next'}
              </Button>
            </div>
          </div>
        </Card>
      )}
    </div>
  )
}
