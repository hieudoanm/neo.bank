import { createContext, useContext, useEffect, useReducer, useCallback, ReactNode } from 'react'
import { authApi } from '@/lib/api/client'
import {
  setSessionTokens,
  getSessionTokens,
  clearSessionTokens,
  setSessionRefresher,
} from '@/lib/authSession'

export type User = {
  userId: string
  accountId: string
  email: string
}

type AuthState = {
  user: User | null
  accessToken: string | null
  refreshToken: string | null
  isAuthenticated: boolean
}

type AuthAction =
  | { type: 'LOGIN'; payload: { user: User; accessToken: string; refreshToken: string } }
  | { type: 'REFRESH'; payload: { accessToken: string; refreshToken: string } }
  | { type: 'LOGOUT' }

const initialState: AuthState = {
  user: null,
  accessToken: null,
  refreshToken: null,
  isAuthenticated: false,
}

function authReducer(state: AuthState, action: AuthAction): AuthState {
  switch (action.type) {
    case 'LOGIN':
      return {
        ...state,
        user: action.payload.user,
        accessToken: action.payload.accessToken,
        refreshToken: action.payload.refreshToken,
        isAuthenticated: true,
      }
    case 'REFRESH':
      return {
        ...state,
        accessToken: action.payload.accessToken,
        refreshToken: action.payload.refreshToken,
      }
    case 'LOGOUT':
      return initialState
  }
}

type AuthContextValue = {
  state: AuthState
  dispatch: React.Dispatch<AuthAction>
  login: (user: User, accessToken: string, refreshToken: string) => void
  refresh: (accessToken: string, refreshToken: string) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

// Session tokens are persisted in sessionStorage so a page reload does not
// silently log the user out; the tab session still ends when the tab closes.
const SESSION_STORAGE_KEY = 'neo.bank.session'

type StoredSession = { user: User; accessToken: string; refreshToken: string }

const readStoredSession = (): StoredSession | null => {
  try {
    const raw = sessionStorage.getItem(SESSION_STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as StoredSession
    if (!parsed?.user?.userId || !parsed.accessToken || !parsed.refreshToken) return null
    return parsed
  } catch {
    return null
  }
}

const initialStateFromStorage = (): AuthState => {
  const stored = readStoredSession()
  if (!stored) return initialState
  return {
    user: stored.user,
    accessToken: stored.accessToken,
    refreshToken: stored.refreshToken,
    isAuthenticated: true,
  }
}

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [state, dispatch] = useReducer(authReducer, initialState, initialStateFromStorage)

  useEffect(() => {
    setSessionRefresher(async () => {
      const tokens = getSessionTokens()
      if (!tokens?.refreshToken) return null
      try {
        const res = await authApi.refresh({ refreshToken: tokens.refreshToken })
        const next = { accessToken: res.accessToken, refreshToken: res.refreshToken }
        setSessionTokens(next)
        dispatch({ type: 'REFRESH', payload: next })
        return next
      } catch {
        clearSessionTokens()
        dispatch({ type: 'LOGOUT' })
        return null
      }
    })
    return () => setSessionRefresher(null)
  }, [])

  useEffect(() => {
    if (state.isAuthenticated && state.user && state.accessToken && state.refreshToken) {
      try {
        sessionStorage.setItem(
          SESSION_STORAGE_KEY,
          JSON.stringify({
            user: state.user,
            accessToken: state.accessToken,
            refreshToken: state.refreshToken,
          }),
        )
      } catch {
        // storage may be unavailable (private mode); session still works in memory
      }
      setSessionTokens({ accessToken: state.accessToken, refreshToken: state.refreshToken })
    } else {
      try {
        sessionStorage.removeItem(SESSION_STORAGE_KEY)
      } catch {
        // ignore storage errors
      }
      clearSessionTokens()
    }
  }, [state])

  const login = useCallback((user: User, accessToken: string, refreshToken: string) => {
    setSessionTokens({ accessToken, refreshToken })
    dispatch({
      type: 'LOGIN',
      payload: { user, accessToken, refreshToken },
    })
  }, [])

  const refresh = useCallback((accessToken: string, refreshToken: string) => {
    setSessionTokens({ accessToken, refreshToken })
    dispatch({ type: 'REFRESH', payload: { accessToken, refreshToken } })
  }, [])

  const logout = useCallback(() => {
    const tokens = getSessionTokens()
    if (tokens?.refreshToken) {
      void authApi.logout({ refreshToken: tokens.refreshToken }).catch(() => {
        // best-effort server-side revocation; local logout proceeds regardless
      })
    }
    clearSessionTokens()
    dispatch({ type: 'LOGOUT' })
  }, [])

  return (
    <AuthContext.Provider value={{ state, dispatch, login, refresh, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}
