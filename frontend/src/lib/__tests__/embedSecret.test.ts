import { describe, expect, it, vi } from 'vitest'
import { EMBED_SECRET_BYTES, generateEmbedSecret } from '../embedSecret'

describe('generateEmbedSecret', () => {
  it('returns 64 lowercase hex characters (256 bits)', () => {
    const secret = generateEmbedSecret()
    expect(secret).toMatch(/^[0-9a-f]{64}$/)
    expect(secret).toHaveLength(EMBED_SECRET_BYTES * 2)
  })

  it('produces different values on each call', () => {
    const seen = new Set(Array.from({ length: 50 }, () => generateEmbedSecret()))
    expect(seen.size).toBe(50)
  })

  it('draws its bytes from crypto.getRandomValues', () => {
    const getRandomValues = vi.fn(<T extends ArrayBufferView | null>(arr: T): T => {
      if (arr instanceof Uint8Array) arr.fill(0xab)
      return arr
    })
    const secret = generateEmbedSecret({ getRandomValues } as Pick<Crypto, 'getRandomValues'>)
    expect(getRandomValues).toHaveBeenCalledTimes(1)
    expect(secret).toBe('ab'.repeat(EMBED_SECRET_BYTES))
  })

  it('does not fall back to Math.random', () => {
    const spy = vi.spyOn(Math, 'random')
    generateEmbedSecret()
    expect(spy).not.toHaveBeenCalled()
    spy.mockRestore()
  })

  it('throws when no secure generator is available', () => {
    expect(() => generateEmbedSecret({} as Pick<Crypto, 'getRandomValues'>)).toThrow(
      /not available/,
    )
  })
})
