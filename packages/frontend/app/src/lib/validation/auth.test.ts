import { describe, it, expect } from 'vitest'
import { validateConfirmPassword, validateEmail, validatePassword } from './auth'

describe('validateEmail', () => {
  it('requires an email', () => {
    expect(validateEmail('')).toBe('Email is required')
  })

  it('rejects a malformed email', () => {
    expect(validateEmail('user@example')).toBe('Enter a valid email address')
    expect(validateEmail('user example.com')).toBe('Enter a valid email address')
  })

  it('accepts a valid email', () => {
    expect(validateEmail('user@example.com')).toBeNull()
  })
})

describe('validatePassword', () => {
  it('requires a password', () => {
    expect(validatePassword('')).toBe('Password is required')
  })

  it('rejects a short password', () => {
    expect(validatePassword('Str0ng!')).toBe('Password must be at least 12 characters')
  })

  it('rejects a password missing a class of character', () => {
    expect(validatePassword('allowercase123!')).toBe(
      'Password must contain uppercase, lowercase, number, and symbol',
    )
    expect(validatePassword('ALLUPPERCASE123!')).toBe(
      'Password must contain uppercase, lowercase, number, and symbol',
    )
    expect(validatePassword('NoNumbersHere!!')).toBe(
      'Password must contain uppercase, lowercase, number, and symbol',
    )
    expect(validatePassword('NoSymbolsHere123')).toBe(
      'Password must contain uppercase, lowercase, number, and symbol',
    )
  })

  it('accepts a strong password', () => {
    expect(validatePassword('Str0ng!Passw0rd')).toBeNull()
  })
})

describe('validateConfirmPassword', () => {
  it('requires confirmation', () => {
    expect(validateConfirmPassword('Str0ng!Passw0rd', '')).toBe('Please confirm your password')
  })

  it('rejects a mismatch', () => {
    expect(validateConfirmPassword('Str0ng!Passw0rd', 'Different!Pass1')).toBe(
      'Passwords do not match',
    )
  })

  it('accepts a match', () => {
    expect(validateConfirmPassword('Str0ng!Passw0rd', 'Str0ng!Passw0rd')).toBeNull()
  })
})
