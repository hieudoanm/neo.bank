import { useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { transfersApi } from '@/lib/api/client'
import { useAuth } from '@/context/AuthContext'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Card } from '@/components/ui/Card'
import { Alert } from '@/components/ui/Alert'
import {
  amountToCents,
  newIdempotencyKey,
  validateAmount,
  validateDestinationAccountId,
} from '@/lib/validation/transfer'

export const TransferPage = () => {
  const navigate = useNavigate()
  const { state } = useAuth()
  const accountId = state.user?.accountId ?? ''
  const accessToken = state.accessToken ?? ''
  const idempotencyKey = useRef(newIdempotencyKey())

  const [destination, setDestination] = useState('')
  const [amount, setAmount] = useState('')
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [formError, setFormError] = useState<string | null>(null)

  const mutation = useMutation({
    mutationFn: () =>
      transfersApi.create(
        { destinationAccountId: destination, amountCents: amountToCents(amount) },
        accessToken,
        idempotencyKey.current,
      ),
    onSuccess: (data) => {
      navigate(`/transfer/${data.transferId}`)
    },
    onError: (err: unknown) => {
      setFieldErrors({})
      setFormError(null)
      const error = err as { status?: number; error?: { code?: string; message?: string } }
      const code = error.error?.code
      if (code === 'INSUFFICIENT_FUNDS') {
        setFormError('Insufficient funds. Your balance is unchanged.')
      } else if (code === 'SAME_ACCOUNT') {
        setFieldErrors({ destinationAccountId: 'You cannot transfer to your own account' })
      } else if (code === 'INVALID_AMOUNT') {
        setFieldErrors({ amount: 'Enter a valid amount' })
      } else if (code === 'ACCOUNT_NOT_FOUND' || error.status === 404) {
        setFieldErrors({ destinationAccountId: 'Account not found' })
      } else if (error.status === 409) {
        setFormError('This transfer was already submitted.')
      } else if (error.status === 401) {
        setFormError('Your session has expired. Please sign in again.')
      } else {
        setFormError('Unable to submit the transfer. Please try again.')
      }
    },
  })

  const handleChange =
    (key: 'destinationAccountId' | 'amount', setter: (value: string) => void) =>
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setter(e.target.value)
      if (fieldErrors[key]) {
        setFieldErrors((prev) => {
          const next = { ...prev }
          delete next[key]
          return next
        })
      }
      if (formError) setFormError(null)
    }

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setFieldErrors({})
    setFormError(null)

    const next: Record<string, string> = {}
    const destinationError = validateDestinationAccountId(destination)
    const amountError = validateAmount(amount)
    if (destinationError) next.destinationAccountId = destinationError
    if (!destinationError && destination === accountId) {
      next.destinationAccountId = 'You cannot transfer to your own account'
    }
    if (amountError) next.amount = amountError
    if (Object.keys(next).length > 0) {
      setFieldErrors(next)
      return
    }

    mutation.mutate()
  }

  const isLoading = mutation.isPending

  return (
    <div className="mx-auto max-w-2xl">
      <header className="mb-6">
        <h1 className="text-text-primary text-2xl font-semibold tracking-tight">
          Transfer money
        </h1>
        <p className="text-text-secondary mt-1 text-sm">Send money to another NEO account.</p>
      </header>

      <Card>
        <h2 className="text-text-primary text-base font-semibold">Transfer details</h2>
        <p className="text-text-secondary mt-1 text-sm">
          Enter the destination account ID and the amount to send.
        </p>

        {formError && (
          <Alert type="error" className="mt-4">
            {formError}
          </Alert>
        )}

        <form onSubmit={handleSubmit} className="mt-6 space-y-4" noValidate>
          <Input
            label="Destination account ID"
            name="destinationAccountId"
            value={destination}
            onChange={handleChange('destinationAccountId', setDestination)}
            error={fieldErrors.destinationAccountId}
            placeholder="00000000-0000-0000-0000-000000000000"
            required
          />
          <Input
            label="Amount"
            name="amount"
            inputMode="decimal"
            value={amount}
            onChange={handleChange('amount', setAmount)}
            error={fieldErrors.amount}
            placeholder="0.00"
            required
          />

          <div className="flex items-center gap-3 pt-2">
            <Button type="submit" variant="primary" className="flex-1" disabled={isLoading}>
              {isLoading ? 'Submitting…' : 'Confirm transfer'}
            </Button>
            <Button
              type="button"
              variant="outline"
              disabled={isLoading}
              onClick={() => navigate('/')}
            >
              Cancel
            </Button>
          </div>
        </form>
      </Card>
    </div>
  )
}
