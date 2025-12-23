import { describe, it, expect } from 'vitest'
import { formatXAxisTick, formatYAxisTick } from '../DashboardPage'

describe('Dashboard formatters', () => {
  it('formats X axis tick to TR short month', () => {
    const iso = new Date(2024, 4, 9).toISOString()
    const out = formatXAxisTick(iso)
    expect(out).toMatch(/09\s*[A-Za-zÇĞİÖŞÜ]/)
  })

  it('formats Y axis tick as string', () => {
    expect(formatYAxisTick(10)).toBe('10')
    expect(formatYAxisTick(0)).toBe('0')
  })
})
