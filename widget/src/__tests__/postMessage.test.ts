/**
 * postMessage security helpers
 */
import { describe, it, expect, vi } from 'vitest'
import {
  DEFAULT_DASHBOARD_ORIGINS,
  DEV_DASHBOARD_ORIGINS,
  createConfigMessageHandler,
  getDashboardOrigins,
  getEventTargetOrigins,
  isAllowedOrigin,
  normalizeOrigin,
  parseOriginList,
  postWidgetEvent,
  validateApiBase,
} from '../utils/postMessage'

describe('normalizeOrigin', () => {
  it.each([
    ['https://botla.app', 'https://botla.app'],
    ['https://botla.app/', 'https://botla.app'],
    ['https://botla.app/some/path?q=1', 'https://botla.app'],
    ['HTTPS://Botla.App', 'https://botla.app'],
    ['http://localhost:5173', 'http://localhost:5173'],
    ['  https://botla.app  ', 'https://botla.app'],
    ['*', null],
    ['', null],
    [undefined, null],
    ['not a url', null],
    ['javascript:alert(1)', null],
    ['data:text/html,hi', null],
    ['file:///etc/passwd', null],
  ])('%s -> %s', (input, expected) => {
    expect(normalizeOrigin(input as string | undefined)).toBe(expected)
  })
})

describe('parseOriginList', () => {
  it('splits, normalizes, drops invalid entries and de-duplicates', () => {
    expect(
      parseOriginList('https://a.example/, https://b.example', 'https://a.example,*,bogus', undefined),
    ).toEqual(['https://a.example', 'https://b.example'])
  })
})

describe('isAllowedOrigin', () => {
  const allowed = ['https://botla.app']
  it.each([
    ['https://botla.app', allowed, true],
    ['https://botla.app.evil.com', allowed, false],
    ['https://botla.app:8443', allowed, false],
    ['http://botla.app', allowed, false],
    ['https://evil.com', allowed, false],
    ['https://sub.botla.app', allowed, false],
    ['null', allowed, false],
    ['', allowed, false],
    ['https://botla.app', [], false],
  ])('%s against %j -> %s', (origin, list, expected) => {
    expect(isAllowedOrigin(origin as string, list as string[])).toBe(expected)
  })
})

describe('getDashboardOrigins', () => {
  it('uses production defaults when nothing is configured', () => {
    expect(getDashboardOrigins('', false)).toEqual(DEFAULT_DASHBOARD_ORIGINS)
  })
  it('uses configured origins (normalized) instead of defaults', () => {
    expect(getDashboardOrigins('https://staging.botla.app/,https://x.example', false)).toEqual([
      'https://staging.botla.app',
      'https://x.example',
    ])
  })
  it('adds local dev-server origins only in development', () => {
    expect(getDashboardOrigins('', true)).toEqual([...DEFAULT_DASHBOARD_ORIGINS, ...DEV_DASHBOARD_ORIGINS])
    for (const o of DEV_DASHBOARD_ORIGINS) expect(getDashboardOrigins('', false)).not.toContain(o)
  })
})

describe('getEventTargetOrigins', () => {
  const dash = ['https://botla.app']
  it.each([
    [undefined, ['https://botla.app']],
    ['https://customer.example/page', ['https://botla.app', 'https://customer.example']],
    ['*', ['https://botla.app']],
    ['garbage', ['https://botla.app']],
    ['https://botla.app', ['https://botla.app']],
  ])('parent-origin %s -> %j', (parent, expected) => {
    expect(getEventTargetOrigins(parent as string | undefined, dash)).toEqual(expected)
  })
})

describe('validateApiBase', () => {
  it.each([
    ['https://api.botla.app', 'https://api.botla.app'],
    ['https://api.botla.app/', 'https://api.botla.app'],
    ['https://api.botla.app/v2/', 'https://api.botla.app/v2'],
    ['http://localhost:8080', 'http://localhost:8080'],
    ['http://127.0.0.1:8080', 'http://127.0.0.1:8080'],
    ['http://api.botla.app', null],
    ['http://evil.com', null],
    ['https://user:pass@api.botla.app', null],
    ['https://api.botla.app?x=1', null],
    ['https://api.botla.app#frag', null],
    ['javascript:alert(1)', null],
    ['/relative', null],
    ['', null],
    [42, null],
    [undefined, null],
  ])('%s -> %s', (input, expected) => {
    expect(validateApiBase(input)).toBe(expected)
  })
})

describe('postWidgetEvent', () => {
  function framedWindow() {
    const parent = { postMessage: vi.fn() }
    const self = { parent } as unknown as Window
    return { parent, self }
  }

  it('posts once per trusted origin and never with a wildcard target', () => {
    const { parent, self } = framedWindow()
    postWidgetEvent('message_sent', { content: 'hi' }, ['https://botla.app', '*', '', 'https://x.example'], self)
    expect(parent.postMessage).toHaveBeenCalledTimes(2)
    expect(parent.postMessage).toHaveBeenCalledWith(
      { type: 'WIDGET_EVENT_MESSAGE_SENT', payload: { content: 'hi' } },
      'https://botla.app',
    )
    expect(parent.postMessage).toHaveBeenCalledWith(expect.anything(), 'https://x.example')
    for (const call of parent.postMessage.mock.calls) expect(call[1]).not.toBe('*')
  })

  it('does nothing when there are no trusted origins', () => {
    const { parent, self } = framedWindow()
    postWidgetEvent('error', {}, [], self)
    expect(parent.postMessage).not.toHaveBeenCalled()
  })

  it('does nothing when not framed', () => {
    const self = {} as { parent: unknown; postMessage: ReturnType<typeof vi.fn> }
    self.parent = self
    self.postMessage = vi.fn()
    postWidgetEvent('error', {}, ['https://botla.app'], self as unknown as Window)
    expect(self.postMessage).not.toHaveBeenCalled()
  })
})

describe('createConfigMessageHandler', () => {
  const parentWin = {} as MessageEventSource
  const otherWin = {} as MessageEventSource

  function setup(allowedOrigins: string[] = ['https://botla.app']) {
    const onConfig = vi.fn()
    const onReject = vi.fn()
    const handler = createConfigMessageHandler({ allowedOrigins, expectedSource: parentWin, onConfig, onReject })
    const send = (data: unknown, origin: string, source: MessageEventSource | null = parentWin) =>
      handler({ data, origin, source } as MessageEvent)
    return { onConfig, onReject, send }
  }

  const cfg = { type: 'WIDGET_CONFIG', config: { 'api-base': 'https://api.botla.app' } }

  it('accepts config from an allowed origin and the parent window', () => {
    const { onConfig, send } = setup()
    send(cfg, 'https://botla.app')
    expect(onConfig).toHaveBeenCalledWith(cfg.config, expect.anything())
  })

  it.each([
    ['https://botla.app.evil.com'],
    ['https://evil.com'],
    ['http://botla.app'],
    ['null'],
  ])('rejects lookalike/foreign origin %s', (origin) => {
    const { onConfig, onReject, send } = setup()
    send(cfg, origin)
    expect(onConfig).not.toHaveBeenCalled()
    expect(onReject).toHaveBeenCalledWith('origin', expect.anything())
  })

  it('fails closed when the allowlist is empty', () => {
    const { onConfig, onReject, send } = setup([])
    send(cfg, 'https://botla.app')
    expect(onConfig).not.toHaveBeenCalled()
    expect(onReject).toHaveBeenCalledWith('origin', expect.anything())
  })

  it('rejects messages from a window other than the parent', () => {
    const { onConfig, onReject, send } = setup()
    send(cfg, 'https://botla.app', otherWin)
    expect(onConfig).not.toHaveBeenCalled()
    expect(onReject).toHaveBeenCalledWith('source', expect.anything())
  })

  it('rejects malformed config payloads', () => {
    const { onConfig, send } = setup()
    send({ type: 'WIDGET_CONFIG', config: 'x' }, 'https://botla.app')
    send({ type: 'WIDGET_CONFIG', config: ['x'] }, 'https://botla.app')
    send({ type: 'WIDGET_CONFIG' }, 'https://botla.app')
    expect(onConfig).not.toHaveBeenCalled()
  })

  it('ignores unrelated messages silently', () => {
    const { onConfig, onReject, send } = setup()
    send({ type: 'SOMETHING_ELSE' }, 'https://evil.com')
    send(null, 'https://evil.com')
    send('string', 'https://evil.com')
    expect(onConfig).not.toHaveBeenCalled()
    expect(onReject).not.toHaveBeenCalled()
  })
})
