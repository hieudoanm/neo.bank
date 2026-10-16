import { describe, it, expect } from 'vitest'
import { formatCurrency, formatDateTime, formatTransactionDate, shortenAccountId } from './format'

describe('formatCurrency', () => {
  it('formats minor units as USD currency by default', () => {
    expect(formatCurrency(123456)).toBe('$1,234.56')
  })

  it('formats zero', () => {
    expect(formatCurrency(0)).toBe('$0.00')
  })

  it('falls back to USD when currency is an empty string', () => {
    expect(formatCurrency(500, '')).toBe('$5.00')
  })

  it('honours an explicit currency', () => {
    expect(formatCurrency(500, 'EUR')).toContain('5.00')
  })
})

describe('formatTransactionDate', () => {
  it('formats a valid ISO timestamp as a readable date', () => {
    const expected = new Intl.DateTimeFormat(undefined, {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
    }).format(new Date('2026-03-15T10:30:00Z'))
    expect(formatTransactionDate('2026-03-15T10:30:00Z')).toBe(expected)
  })

  it('returns the input unchanged when the date is invalid', () => {
    expect(formatTransactionDate('not-a-date')).toBe('not-a-date')
  })
})

describe('formatDateTime', () => {
  it('formats a valid ISO timestamp with a UTC suffix', () => {
    expect(formatDateTime('2026-03-15T10:30:00Z')).toMatch(/\d{2}.*2026.*\d{2}:\d{2} UTC/)
  })

  it('returns the input unchanged when the date is invalid', () => {
    expect(formatDateTime('nope')).toBe('nope')
  })
})

describe('shortenAccountId', () => {
  it('renders an ACC- prefix with the last four characters uppercased', () => {
    expect(shortenAccountId('11111111-2222-3333-4444-555566667777')).toBe('ACC-7777')
  })

  it('strips dashes before taking the tail', () => {
    expect(shortenAccountId('abcd-ef01')).toBe('ACC-EF01')
  })
})
