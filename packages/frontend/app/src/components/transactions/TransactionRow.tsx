import type { TransactionResponse, TransactionStatus } from '@/lib/api/client'
import { formatCurrency, formatTransactionDate, shortenAccountId } from '@/lib/format'

const statusStyles: Record<TransactionStatus, string> = {
  completed: 'bg-success-bg text-success-fg',
  failed: 'bg-error-bg text-error-fg',
  pending: 'bg-warning-bg text-warning-fg',
  reversed: 'bg-info-bg text-info-fg',
}

export const TransactionRow = ({ transaction }: { transaction: TransactionResponse }) => {
  const isCredit = transaction.type === 'credit'
  const title = `${isCredit ? 'Transfer from' : 'Transfer to'} ${shortenAccountId(
    transaction.counterpartyAccountId,
  )}`

  return (
    <li className="flex items-center gap-4 px-4 py-3 border-b border-border-light last:border-b-0">
      <span
        aria-hidden="true"
        className={`inline-flex h-10 w-10 items-center justify-center rounded-full text-xs font-semibold ${
          isCredit ? 'bg-success-bg text-success-fg' : 'bg-error-bg text-error-fg'
        }`}
      >
        {isCredit ? 'IN' : 'OUT'}
      </span>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-text-primary">{title}</p>
        <p className="text-xs text-text-secondary">
          {formatTransactionDate(transaction.createdAt)}
        </p>
      </div>
      <div className="text-right">
        <p
          className={`text-sm font-semibold ${
            isCredit ? 'text-success-fg' : 'text-error-fg'
          }`}
        >
          {isCredit ? '+' : '-'}
          {formatCurrency(transaction.amountCents)}
        </p>
        {transaction.status && (
          <span
            className={`mt-1 inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
              statusStyles[transaction.status]
            }`}
          >
            {transaction.status}
          </span>
        )}
      </div>
      <span className="sr-only">{isCredit ? 'credit' : 'debit'} transaction</span>
    </li>
  )
}
