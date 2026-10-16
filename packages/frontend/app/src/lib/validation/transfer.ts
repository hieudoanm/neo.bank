const UUID_PATTERN =
  /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/

export const MAX_AMOUNT_CENTS = 9999999999

export const validateDestinationAccountId = (value: string): string | null => {
  if (!value) return 'Destination account ID is required'
  if (!UUID_PATTERN.test(value)) return 'Enter a valid account ID'
  return null
}

export const validateAmount = (value: string): string | null => {
  if (!value) return 'Amount is required'
  if (!/^\d+(\.\d{1,2})?$/.test(value)) return 'Enter a valid amount'
  const cents = amountToCents(value)
  if (cents < 1) return 'Amount must be greater than zero'
  if (cents > MAX_AMOUNT_CENTS) return 'Amount is too large'
  return null
}

export const amountToCents = (value: string): number => Math.round(Number(value) * 100)

export const newIdempotencyKey = (): string =>
  `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}-${Math.random()
    .toString(36)
    .slice(2)}`
