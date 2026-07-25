// Pick a readable text color (near-black or white) for a given background,
// based on WCAG relative luminance. The crossover where black vs white gives
// equal contrast is L ≈ 0.179, so above it we use dark text, below it light.

const DARK = '#0f172a' // slate-900
const LIGHT = '#ffffff'

function channel(value: number): number {
  const v = value / 255
  return v <= 0.03928 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4)
}

export function textColorFor(background: string): string {
  const hex = background.replace('#', '')
  if (hex.length !== 6) return DARK

  const r = parseInt(hex.slice(0, 2), 16)
  const g = parseInt(hex.slice(2, 4), 16)
  const b = parseInt(hex.slice(4, 6), 16)
  if ([r, g, b].some(Number.isNaN)) return DARK

  const luminance = 0.2126 * channel(r) + 0.7152 * channel(g) + 0.0722 * channel(b)
  return luminance > 0.179 ? DARK : LIGHT
}
