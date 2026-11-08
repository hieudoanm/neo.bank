import { createContext, useContext, useReducer, useCallback, ReactNode } from 'react'

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

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [state, dispatch] = useReducer(authReducer, initialState)

  const login = useCallback((user: User, accessToken: string, refreshToken: string) => {
    dispatch({
      type: 'LOGIN',
      payload: { user, accessToken, refreshToken },
    })
  }, [])

  const refresh = useCallback((accessToken: string, refreshToken: string) => {
    dispatch({ type: 'REFRESH', payload: { accessToken, refreshToken } })
  }, [])

  const logout = useCallback(() => {
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
