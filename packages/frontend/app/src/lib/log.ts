export type LogLevel = 'debug' | 'info' | 'warn' | 'error'

export type LogContext = Record<string, unknown>

const LEVEL_PRIORITY: Record<LogLevel, number> = {
  debug: 10,
  info: 20,
  warn: 30,
  error: 40,
}

const isLogLevel = (value: string): value is LogLevel => value in LEVEL_PRIORITY

const resolveMinLevel = (): LogLevel => {
  const configured = import.meta.env.VITE_LOG_LEVEL
  if (typeof configured === 'string' && isLogLevel(configured.toLowerCase())) {
    return configured.toLowerCase() as LogLevel
  }
  return import.meta.env.DEV ? 'debug' : 'info'
}

let minLevel: LogLevel = resolveMinLevel()

export const getLogLevel = (): LogLevel => minLevel

export const setLogLevel = (level: LogLevel): void => {
  minLevel = level
}

const serializeContext = (context?: LogContext): LogContext | undefined => {
  if (!context) return undefined
  const serialized: LogContext = {}
  for (const [key, value] of Object.entries(context)) {
    serialized[key] =
      value instanceof Error
        ? { name: value.name, message: value.message, stack: value.stack }
        : value
  }
  return serialized
}

const write = (level: LogLevel, message: string, context?: LogContext): void => {
  if (LEVEL_PRIORITY[level] < LEVEL_PRIORITY[minLevel]) return
  const line = `${new Date().toISOString()} [${level.toUpperCase()}] ${message}`
  const serialized = serializeContext(context)
  if (serialized === undefined) {
    console[level](line)
    return
  }
  console[level](line, serialized)
}

export const logger = {
  debug: (message: string, context?: LogContext) => write('debug', message, context),
  info: (message: string, context?: LogContext) => write('info', message, context),
  warn: (message: string, context?: LogContext) => write('warn', message, context),
  error: (message: string, context?: LogContext) => write('error', message, context),
}

export type Logger = typeof logger

export const createLogger = (scope: string): Logger => ({
  debug: (message, context) => write('debug', `${scope}: ${message}`, context),
  info: (message, context) => write('info', `${scope}: ${message}`, context),
  warn: (message, context) => write('warn', `${scope}: ${message}`, context),
  error: (message, context) => write('error', `${scope}: ${message}`, context),
})
