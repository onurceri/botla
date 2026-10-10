import { describe, it, expect } from 'vitest'
import { getMessages, normalizeLang, resolveLang, t } from '../i18n'

describe('i18n', () => {
  it('normalizes language codes to supported languages', () => {
    expect(normalizeLang('en')).toBe('en')
    expect(normalizeLang('en-US')).toBe('en')
    expect(normalizeLang('TR_tr')).toBe('tr')
    expect(normalizeLang('de-DE')).toBeUndefined()
    expect(normalizeLang('')).toBeUndefined()
    expect(normalizeLang(undefined)).toBeUndefined()
  })

  it('falls back to Turkish when the language is missing or unknown', () => {
    expect(resolveLang()).toBe('tr')
    expect(resolveLang(undefined, null, '')).toBe('tr')
    expect(resolveLang('fr')).toBe('tr')
  })

  it('uses the first supported candidate (override before bot language)', () => {
    expect(resolveLang('en', 'tr-TR')).toBe('en')
    expect(resolveLang(undefined, 'en-GB')).toBe('en')
    expect(resolveLang('xx', 'en')).toBe('en')
  })

  it('returns strings for both languages', () => {
    expect(t('inputPlaceholder')).toBe('Mesaj yazın...')
    expect(t('inputPlaceholder', 'tr')).toBe('Mesaj yazın...')
    expect(t('inputPlaceholder', 'en')).toBe('Type a message...')
    expect(getMessages('en').unread(3)).toBe('3 unread')
    expect(getMessages('tr').unread(3)).toBe('Okunmamış 3')
  })

  it('has the same keys and non-empty values in every language', () => {
    const trKeys = Object.keys(getMessages('tr')).sort()
    expect(Object.keys(getMessages('en')).sort()).toEqual(trKeys)
    for (const lang of ['tr', 'en'] as const) {
      for (const v of Object.values(getMessages(lang))) {
        if (typeof v === 'string') expect(v.length).toBeGreaterThan(0)
      }
    }
  })
})
