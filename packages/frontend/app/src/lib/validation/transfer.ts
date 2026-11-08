const UUID_PATTERN = /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/

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

// amountToCents parses the decimal string directly so values like "19.99"
// never pass through a lossy float multiply.
export const amountToCents = (value: string): number => {
  const [whole = '0', fraction = ''] = value.split('.')
  const cents = Number(whole) * 100 + Number(fraction.padEnd(2, '0').slice(0, 2))
  return cents
}

export const newIdempotencyKey = (): string => crypto.randomUUID()
