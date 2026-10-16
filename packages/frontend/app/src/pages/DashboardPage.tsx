import { useBalance } from '@/hooks/useBalance'
import { Alert } from '@/components/ui/Alert'
import { Card } from '@/components/ui/Card'
import { useAuth } from '@/context/AuthContext'
import { formatCurrency } from '@/lib/format'

export const DashboardPage = () => {
  const { state } = useAuth()
  const { data, isLoading, isError, refetch, isFetching } = useBalance()

  return (
    <div className="space-y-6">
      <h1 className="text-text-primary text-2xl font-semibold tracking-tight mb-2">
        Dashboard
      </h1>

      {state.user && (
        <p className="text-text-secondary text-sm">Account: {state.user.accountId}</p>
      )}

      {isError && (
        <Alert type="error" className="mb-4">
          Unable to load your balance. Please try again.
          <button
            type="button"
            className="btn btn-sm btn-outline mt-2"
            onClick={() => refetch()}
            disabled={isFetching}
          >
            Retry
          </button>
        </Alert>
      )}

      {isLoading && (
        <Card className="w-full max-w-md">
          <h2 className="text-sm font-medium text-text-secondary">Current balance</h2>
          <div className="mt-2 h-8 bg-background-secondary rounded animate-pulse" />
        </Card>
      )}

      {!isLoading && !isError && data && (
        <Card className="w-full max-w-md">
          <h2 className="text-sm font-medium text-text-secondary">Current balance</h2>
          <p className="mt-2 text-3xl font-bold text-text-primary">
            {formatCurrency(data.balanceCents, data.currency)}
          </p>
          <p className="mt-1 text-xs text-text-tertiary">
            {data.balanceCents} minor units ({data.currency})
          </p>
        </Card>
      )}
    </div>
  )
}
