import { describe, it, expect, beforeEach } from 'vitest'
import { act, renderHook } from '@testing-library/react'
import type { ReactNode } from 'react'
import { AuthProvider, useAuth } from './AuthContext'

const wrapper = ({ children }: { children: ReactNode }) => <AuthProvider>{children}</AuthProvider>

describe('AuthContext', () => {
  beforeEach(() => {
    sessionStorage.clear()
  })

  it('starts logged out', () => {
    const { result } = renderHook(() => useAuth(), { wrapper })
    expect(result.current.state).toEqual({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
    })
  })

  it('logs in, refreshes tokens, and logs out', () => {
    const { result } = renderHook(() => useAuth(), { wrapper })

    act(() => {
      result.current.login({ userId: 'u1', accountId: 'a1', email: 'user@example.com' }, 'at', 'rt')
    })
    expect(result.current.state.isAuthenticated).toBe(true)
    expect(result.current.state.user?.accountId).toBe('a1')
    expect(result.current.state.accessToken).toBe('at')
    expect(result.current.state.refreshToken).toBe('rt')

    act(() => {
      result.current.refresh('at2', 'rt2')
    })
    expect(result.current.state.accessToken).toBe('at2')
    expect(result.current.state.refreshToken).toBe('rt2')
    expect(result.current.state.isAuthenticated).toBe(true)
    expect(result.current.state.user?.accountId).toBe('a1')

    act(() => {
      result.current.logout()
    })
    expect(result.current.state.isAuthenticated).toBe(false)
    expect(result.current.state.accessToken).toBeNull()
    expect(result.current.state.user).toBeNull()
  })

  it('restores the session from sessionStorage on mount', () => {
    sessionStorage.setItem(
      'neo.bank.session',
      JSON.stringify({
        user: { userId: 'u9', accountId: 'a9', email: 'user@example.com' },
        accessToken: 'stored-at',
        refreshToken: 'stored-rt',
      }),
    )
    const { result } = renderHook(() => useAuth(), { wrapper })
    expect(result.current.state.isAuthenticated).toBe(true)
    expect(result.current.state.user?.accountId).toBe('a9')
    expect(result.current.state.accessToken).toBe('stored-at')
  })

  it('clears the stored session on logout', () => {
    const { result } = renderHook(() => useAuth(), { wrapper })
    act(() => {
      result.current.login({ userId: 'u1', accountId: 'a1', email: 'user@example.com' }, 'at', 'rt')
    })
    expect(sessionStorage.getItem('neo.bank.session')).not.toBeNull()
    act(() => {
      result.current.logout()
    })
    expect(sessionStorage.getItem('neo.bank.session')).toBeNull()
  })

  it('throws when useAuth is used outside of AuthProvider', () => {
    expect(() => renderHook(() => useAuth())).toThrow('useAuth must be used within AuthProvider')
  })
})
