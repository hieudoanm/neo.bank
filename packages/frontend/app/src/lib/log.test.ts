import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { logger, createLogger, setLogLevel, getLogLevel } from './log'

type Level = 'debug' | 'info' | 'warn' | 'error'

const FIXED_TIME = '2026-01-01T12:00:00.000Z'

describe('log', () => {
  let spies: Record<Level, ReturnType<typeof vi.spyOn>>

  beforeEach(() => {
    spies = {
      debug: vi.spyOn(console, 'debug').mockImplementation(() => {}),
      info: vi.spyOn(console, 'info').mockImplementation(() => {}),
      warn: vi.spyOn(console, 'warn').mockImplementation(() => {}),
      error: vi.spyOn(console, 'error').mockImplementation(() => {}),
    }
    setLogLevel('debug')
    vi.useFakeTimers()
    vi.setSystemTime(new Date(FIXED_TIME))
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
    setLogLevel('debug')
  })

  it('prefixes an ISO timestamp and uppercase level', () => {
    logger.info('hello')
    expect(spies.info).toHaveBeenCalledWith(`${FIXED_TIME} [INFO] hello`)
  })

  it('routes each level to the matching console method', () => {
    logger.debug('d')
    logger.info('i')
    logger.warn('w')
    logger.error('e')

    expect(spies.debug).toHaveBeenCalledWith(`${FIXED_TIME} [DEBUG] d`)
    expect(spies.info).toHaveBeenCalledWith(`${FIXED_TIME} [INFO] i`)
    expect(spies.warn).toHaveBeenCalledWith(`${FIXED_TIME} [WARN] w`)
    expect(spies.error).toHaveBeenCalledWith(`${FIXED_TIME} [ERROR] e`)
  })

  it('appends a structured context object', () => {
    logger.info('balance loaded', { accountId: 'acc-1' })
    expect(spies.info).toHaveBeenCalledWith(`${FIXED_TIME} [INFO] balance loaded`, {
      accountId: 'acc-1',
    })
  })

  it('serialises Error values in the context', () => {
    const err = new Error('boom')
    logger.error('request failed', { error: err })

    const [, context] = spies.error.mock.calls[0] as [string, Record<string, unknown>]
    expect(context).toEqual({ error: { name: 'Error', message: 'boom', stack: err.stack } })
  })

  it('suppresses levels below the configured minimum', () => {
    setLogLevel('warn')
    logger.debug('d')
    logger.info('i')
    logger.warn('w')
    logger.error('e')

    expect(spies.debug).not.toHaveBeenCalled()
    expect(spies.info).not.toHaveBeenCalled()
    expect(spies.warn).toHaveBeenCalledTimes(1)
    expect(spies.error).toHaveBeenCalledTimes(1)
  })

  it('exposes the configured minimum level', () => {
    setLogLevel('error')
    expect(getLogLevel()).toBe('error')
  })

  it('createLogger prefixes messages with a scope', () => {
    const log = createLogger('auth')
    log.warn('token expiring', { seconds: 30 })
    expect(spies.warn).toHaveBeenCalledWith(`${FIXED_TIME} [WARN] auth: token expiring`, {
      seconds: 30,
    })
  })
})
