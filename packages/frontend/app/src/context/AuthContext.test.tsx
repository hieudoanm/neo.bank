import { describe, it, expect } from 'vitest'
import { act, renderHook } from '@testing-library/react'
import type { ReactNode } from 'react'
import { AuthProvider, useAuth } from './AuthContext'

const wrapper = ({ children }: { children: ReactNode }) => <AuthProvider>{children}</AuthProvider>

describe('AuthContext', () => {
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

  it('throws when useAuth is used outside of AuthProvider', () => {
    expect(() => renderHook(() => useAuth())).toThrow('useAuth must be used within AuthProvider')
  })
})
