/** Number of random bytes in a generated embed secret (256 bits). */
export const EMBED_SECRET_BYTES = 32

/**
 * Generates a new embed-token signing secret using the Web Crypto CSPRNG.
 *
 * The secret is returned as lowercase hex (64 characters). It is used as the
 * HMAC key that customers' servers sign embed tokens with, so it must come from
 * a cryptographically secure source — never Math.random().
 */
export function generateEmbedSecret(
  cryptoImpl: Pick<Crypto, 'getRandomValues'> = globalThis.crypto,
): string {
  if (!cryptoImpl || typeof cryptoImpl.getRandomValues !== 'function') {
    throw new Error('Secure random number generator is not available')
  }
  const bytes = new Uint8Array(EMBED_SECRET_BYTES)
  cryptoImpl.getRandomValues(bytes)
  return Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('')
}
