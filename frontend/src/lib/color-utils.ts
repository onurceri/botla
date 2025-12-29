export type RgbaColor = {
  r: number
  g: number
  b: number
  a: number
}

/**
 * Parse a color string (HEX or RGBA) to an RgbaColor object
 */
export function parseColor(value: string): RgbaColor {
  // Handle RGBA format: rgba(255, 255, 255, 0.5)
  const rgbaMatch = value.match(/rgba?\s*\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*(?:,\s*([\d.]+))?\s*\)/)
  if (rgbaMatch) {
    return {
      r: parseInt(rgbaMatch[1], 10),
      g: parseInt(rgbaMatch[2], 10),
      b: parseInt(rgbaMatch[3], 10),
      a: rgbaMatch[4] !== undefined ? parseFloat(rgbaMatch[4]) : 1,
    }
  }

  // Handle HEX format: #RRGGBB or #RRGGBBAA or #RGB
  if (value.startsWith('#')) {
    let hex = value.slice(1)

    // Handle shorthand hex (#RGB -> #RRGGBB)
    if (hex.length === 3) {
      hex = hex[0] + hex[0] + hex[1] + hex[1] + hex[2] + hex[2]
    }

    // Handle 8-char hex with alpha (#RRGGBBAA)
    if (hex.length === 8) {
      return {
        r: parseInt(hex.slice(0, 2), 16),
        g: parseInt(hex.slice(2, 4), 16),
        b: parseInt(hex.slice(4, 6), 16),
        a: parseInt(hex.slice(6, 8), 16) / 255,
      }
    }

    // Handle 6-char hex (#RRGGBB)
    if (hex.length === 6) {
      return {
        r: parseInt(hex.slice(0, 2), 16),
        g: parseInt(hex.slice(2, 4), 16),
        b: parseInt(hex.slice(4, 6), 16),
        a: 1,
      }
    }
  }

  // Default to black if parsing fails
  return { r: 0, g: 0, b: 0, a: 1 }
}

/**
 * Format an RgbaColor object to HEX string.
 * Returns #RRGGBB for fully opaque, #RRGGBBAA for transparent.
 */
export function formatRgba(color: RgbaColor): string {
  // Round RGB to integers and alpha to 2 decimal places
  const r = Math.round(color.r)
  const g = Math.round(color.g)
  const b = Math.round(color.b)
  const a = Math.round(color.a * 100) / 100

  return `rgba(${r}, ${g}, ${b}, ${a})`
}

/**
 * Convert an RgbaColor to HEX format (ignores alpha if 1)
 */
export function rgbaToHex(color: RgbaColor): string {
  const toHex = (n: number) => n.toString(16).padStart(2, '0')
  const hex = `#${toHex(color.r)}${toHex(color.g)}${toHex(color.b)}`
  if (color.a < 1) {
    return `${hex}${toHex(Math.round(color.a * 255))}`
  }
  return hex
}

/**
 * Check if a color string is valid (HEX or RGBA)
 */
export function isValidColor(value: string): boolean {
  // Check RGBA format
  if (/rgba?\s*\(\s*\d+\s*,\s*\d+\s*,\s*\d+\s*(?:,\s*[\d.]+)?\s*\)/.test(value)) {
    return true
  }
  // Check HEX format
  if (/^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$/.test(value)) {
    return true
  }
  return false
}
