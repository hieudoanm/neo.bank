import { useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { useTransfer } from '@/hooks/useTransfer'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Alert } from '@/components/ui/Alert'
import { formatCurrency, formatDateTime, shortenAccountId } from '@/lib/format'

const failureMessages: Record<string, string> = {
  INSUFFICIENT_FUNDS: 'Insufficient funds. Your balance is unchanged.',
  ACCOUNT_NOT_FOUND: 'The destination account could not be found.',
  SAME_ACCOUNT: 'You cannot transfer to your own account.',
  INVALID_AMOUNT: 'The amount is not valid.',
  SERVICE_UNAVAILABLE: 'The service is temporarily unavailable. Please try again.',
}

const errorStatus = (err: unknown): number | undefined =>
  (err as { status?: number } | undefined)?.status

const CheckIcon = () => (
  <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" className="h-8 w-8">
    <path
      d="M20 6 9 17l-5-5"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
)

const CrossIcon = () => (
  <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" className="h-8 w-8">
    <path
      d="M18 6 6 18M6 6l12 12"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
)

export const TransferStatusPage = () => {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { transferId } = useParams<{ transferId: string }>()
  const { data, isLoading, isError, error, refetch, isFetching } = useTransfer(transferId ?? '')

  const httpStatus = isError ? errorStatus(error) : undefined
  const isAuthError = httpStatus === 401
  const isPermissionError = httpStatus === 404

  const transferStatus = data?.status
  useEffect(() => {
    if (transferStatus === 'completed') {
      void queryClient.invalidateQueries({ queryKey: ['balance'] })
      void queryClient.invalidateQueries({ queryKey: ['transactions'] })
    }
  }, [transferStatus, queryClient])

  const renderDetails = (destination: string, amountCents: number, createdAt: string) => (
    <dl className="mt-6 flex flex-col border-t border-border text-left text-sm">
      <div className="flex items-center justify-between border-b border-border py-3">
        <dt className="text-text-secondary">Transfer ID</dt>
        <dd className="font-mono text-xs text-text-primary">{transferId}</dd>
      </div>
      <div className="flex items-center justify-between border-b border-border py-3">
        <dt className="text-text-secondary">To</dt>
        <dd className="font-mono text-xs text-text-primary">{shortenAccountId(destination)}</dd>
      </div>
      <div className="flex items-center justify-between border-b border-border py-3">
        <dt className="text-text-secondary">Amount</dt>
        <dd className="text-sm font-semibold text-text-primary">
          {formatCurrency(amountCents)}
        </dd>
      </div>
      <div className="flex items-center justify-between py-3">
        <dt className="text-text-secondary">Date</dt>
        <dd className="text-text-primary">{formatDateTime(createdAt)}</dd>
      </div>
    </dl>
  )

  return (
    <div className="mx-auto max-w-2xl">
      <header className="mb-6">
        <h1 className="text-text-primary text-2xl font-semibold tracking-tight">
          Transfer status
        </h1>
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

      {isPermissionError && <Alert type="error">Transfer not found.</Alert>}

      {isError && !isAuthError && !isPermissionError && (
        <Alert type="error">
          Unable to load this transfer. Please try again.
          <div className="mt-2">
            <Button variant="outline" disabled={isFetching} onClick={() => void refetch()}>
              Try again
            </Button>
          </div>
        </Alert>
      )}

      {isLoading && (
        <Card className="items-center text-center">
          <span
            aria-hidden="true"
            className="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-primary"
          />
          <p className="mt-3 text-sm text-text-secondary" aria-live="polite">
            Checking transfer status…
          </p>
        </Card>
      )}

      {!isLoading && !isError && data && (data.status === 'pending' || data.status === 'processing') && (
        <Card className="items-center text-center">
          <span
            aria-hidden="true"
            className="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-primary"
          />
          <p className="mt-3 text-base font-semibold text-text-primary" aria-live="polite">
            Processing your transfer
          </p>
          <p className="mt-1 text-sm text-text-secondary">
            This usually takes a few seconds. The page updates automatically.
          </p>
          {renderDetails(data.destinationAccountId, data.amountCents, data.createdAt)}
        </Card>
      )}

      {!isLoading && !isError && data && data.status === 'completed' && (
        <Card className="items-center text-center">
          <span className="flex h-16 w-16 items-center justify-center rounded-full bg-success-bg text-success-fg">
            <CheckIcon />
          </span>
          <h2 className="mt-4 text-xl font-semibold text-text-primary" aria-live="polite">
            Transfer complete
          </h2>
          <p className="mt-1 text-sm text-text-secondary">
            Your money is on its way to {shortenAccountId(data.destinationAccountId)}.
          </p>
          {renderDetails(data.destinationAccountId, data.amountCents, data.createdAt)}
          <div className="mt-6 flex flex-wrap items-center justify-center gap-3">
            <Button variant="primary" onClick={() => navigate('/')}>
              Back to dashboard
            </Button>
            <Button variant="outline" onClick={() => navigate('/transactions')}>
              View transactions
            </Button>
            <Button variant="outline" onClick={() => navigate('/transfer')}>
              Make another transfer
            </Button>
          </div>
        </Card>
      )}

      {!isLoading && !isError && data && data.status === 'failed' && (
        <Card className="items-center text-center">
          <span className="flex h-16 w-16 items-center justify-center rounded-full bg-error-bg text-error-fg">
            <CrossIcon />
          </span>
          <h2 className="mt-4 text-xl font-semibold text-text-primary" aria-live="polite">
            Transfer failed
          </h2>
          <p className="mt-1 text-sm text-text-secondary">
            {failureMessages[data.failureReason ?? ''] ??
              'The transfer could not be completed.'}
          </p>
          {renderDetails(data.destinationAccountId, data.amountCents, data.createdAt)}
          <div className="mt-6 flex flex-wrap items-center justify-center gap-3">
            <Button variant="primary" onClick={() => navigate('/transfer')}>
              Try again
            </Button>
            <Button variant="outline" onClick={() => navigate('/')}>
              Back to dashboard
            </Button>
          </div>
        </Card>
      )}

      {!isLoading && !isError && data && data.status === 'reversed' && (
        <Card className="items-center text-center">
          <span className="flex h-16 w-16 items-center justify-center rounded-full bg-info-bg text-info-fg">
            <CrossIcon />
          </span>
          <h2 className="mt-4 text-xl font-semibold text-text-primary" aria-live="polite">
            Transfer reversed
          </h2>
          <p className="mt-1 text-sm text-text-secondary">
            This transfer was reversed. Your balance is unchanged.
          </p>
          {renderDetails(data.destinationAccountId, data.amountCents, data.createdAt)}
          <div className="mt-6 flex flex-wrap items-center justify-center gap-3">
            <Button variant="outline" onClick={() => navigate('/')}>
              Back to dashboard
            </Button>
          </div>
        </Card>
      )}
    </div>
  )
}
