import { describe, it, expect, vi, afterEach } from 'vitest'
import {
  MAX_AMOUNT_CENTS,
  amountToCents,
  newIdempotencyKey,
  validateAmount,
  validateDestinationAccountId,
} from './transfer'

const VALID_UUID = '11111111-2222-3333-4444-555566667777'

describe('validateDestinationAccountId', () => {
  it('requires a destination account id', () => {
    expect(validateDestinationAccountId('')).toBe('Destination account ID is required')
  })

  it('rejects a non-UUID value', () => {
    expect(validateDestinationAccountId('ACC-8842')).toBe('Enter a valid account ID')
  })

  it('accepts a UUID', () => {
    expect(validateDestinationAccountId(VALID_UUID)).toBeNull()
  })
})

describe('validateAmount', () => {
  it('requires an amount', () => {
    expect(validateAmount('')).toBe('Amount is required')
  })

  it('rejects a malformed amount', () => {
    expect(validateAmount('12.345')).toBe('Enter a valid amount')
    expect(validateAmount('abc')).toBe('Enter a valid amount')
  })

  it('rejects an amount that rounds to zero', () => {
    expect(validateAmount('0')).toBe('Amount must be greater than zero')
  })

  it('rejects an amount above the maximum', () => {
    expect(validateAmount(String(MAX_AMOUNT_CENTS / 100 + 1))).toBe('Amount is too large')
  })

  it('accepts a valid amount', () => {
    expect(validateAmount('12.50')).toBeNull()
  })
})

describe('amountToCents', () => {
  it('converts a decimal string to rounded minor units', () => {
    expect(amountToCents('12.34')).toBe(1234)
    expect(amountToCents('0.1')).toBe(10)
  })
})

describe('newIdempotencyKey', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('produces a non-empty, space-free key', () => {
    const key = newIdempotencyKey()
    expect(key).not.toHaveLength(0)
    expect(key).not.toMatch(/\s/)
  })
})
