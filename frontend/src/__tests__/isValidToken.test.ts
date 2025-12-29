import { describe, it, expect } from 'vitest'
import { isValidToken } from '../App'

describe('isValidToken', () => {
  describe('returns false for invalid tokens', () => {
    it('rejects null', () => {
      expect(isValidToken(null)).toBe(false)
    })

    it('rejects empty string', () => {
      expect(isValidToken('')).toBe(false)
    })

    it('rejects "undefined" string', () => {
      expect(isValidToken('undefined')).toBe(false)
    })

    it('rejects "null" string', () => {
      expect(isValidToken('null')).toBe(false)
    })

    it('rejects non-JWT strings', () => {
      expect(isValidToken('abc')).toBe(false)
      expect(isValidToken('invalid_token')).toBe(false)
      expect(isValidToken('some.random.text.with.dots')).toBe(false)
    })

    it('rejects tokens with wrong number of parts', () => {
      expect(isValidToken('part1.part2')).toBe(false)
      expect(isValidToken('part1.part2.part3.part4')).toBe(false)
      expect(isValidToken('singlepart')).toBe(false)
    })

    it('rejects tokens with empty parts', () => {
      expect(isValidToken('...')).toBe(false)
      expect(isValidToken('.part2.part3')).toBe(false)
      expect(isValidToken('part1..part3')).toBe(false)
      expect(isValidToken('part1.part2.')).toBe(false)
    })

    it('rejects tokens with invalid base64url characters', () => {
      expect(isValidToken('abc!.def.ghi')).toBe(false)
      expect(isValidToken('abc.def$.ghi')).toBe(false)
      expect(isValidToken('abc.def.ghi#')).toBe(false)
    })

    it('rejects expired tokens', () => {
      // Create an expired JWT payload
      const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' })).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
      const expiredPayload = btoa(JSON.stringify({ 
        exp: Math.floor(Date.now() / 1000) - 3600, // Expired 1 hour ago
        sub: 'test-user' 
      })).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
      const signature = 'dummysignature'
      const expiredToken = `${header}.${expiredPayload}.${signature}`
      
      expect(isValidToken(expiredToken)).toBe(false)
    })

    it('rejects tokens with malformed JSON payload', () => {
      const header = btoa(JSON.stringify({ alg: 'HS256' })).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
      // Not valid JSON when decoded
      const badPayload = btoa('not valid json').replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
      const signature = 'sig'
      
      expect(isValidToken(`${header}.${badPayload}.${signature}`)).toBe(false)
    })
  })

  describe('returns true for valid tokens', () => {
    it('accepts a valid JWT format with future expiry', () => {
      const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' })).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
      const validPayload = btoa(JSON.stringify({ 
        exp: Math.floor(Date.now() / 1000) + 3600, // Expires in 1 hour
        sub: 'test-user',
        iat: Math.floor(Date.now() / 1000)
      })).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
      const signature = 'dummysignaturevalue'
      const validToken = `${header}.${validPayload}.${signature}`
      
      expect(isValidToken(validToken)).toBe(true)
    })

    it('accepts a valid JWT format without exp claim', () => {
      const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' })).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
      const payloadWithoutExp = btoa(JSON.stringify({ 
        sub: 'test-user',
        iat: Math.floor(Date.now() / 1000)
      })).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
      const signature = 'dummysignaturevalue'
      const validToken = `${header}.${payloadWithoutExp}.${signature}`
      
      expect(isValidToken(validToken)).toBe(true)
    })

    it('accepts tokens with base64url special characters', () => {
      const header = btoa(JSON.stringify({ alg: 'HS256' })).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
      const payload = btoa(JSON.stringify({ sub: 'test' })).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
      const signature = 'abc_def-123' // Valid base64url characters
      const token = `${header}.${payload}.${signature}`
      
      expect(isValidToken(token)).toBe(true)
    })
  })
})
