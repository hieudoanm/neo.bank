type SessionTokens = {
  accessToken: string
  refreshToken: string
}

type TokenRefresher = () => Promise<SessionTokens | null>

let sessionTokens: SessionTokens | null = null
let refresher: TokenRefresher | null = null
let refreshPromise: Promise<SessionTokens | null> | null = null

export const setSessionTokens = (tokens: SessionTokens | null) => {
  sessionTokens = tokens
}

export const getSessionTokens = (): SessionTokens | null => sessionTokens

export const clearSessionTokens = () => {
  sessionTokens = null
}

export const setSessionRefresher = (fn: TokenRefresher | null) => {
  refresher = fn
}

// refreshSession dedupes concurrent callers onto a single in-flight refresh so
// parallel 401s cannot race the rotating refresh token.
export const refreshSession = (): Promise<SessionTokens | null> => {
  if (!refresher) return Promise.resolve(null)
  if (!refreshPromise) {
    refreshPromise = refresher().finally(() => {
      refreshPromise = null
    })
  }
  return refreshPromise
}
