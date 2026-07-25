import { describe, it, expect } from 'vitest'
import { textColorFor } from './contrast'

describe('textColorFor', () => {
  it('uses dark text on light backgrounds', () => {
    expect(textColorFor('#ffffff')).toBe('#0f172a')
    expect(textColorFor('#f1f5f9')).toBe('#0f172a')
    expect(textColorFor('#facc15')).toBe('#0f172a') // yellow
  })

  it('uses light text on dark backgrounds', () => {
    expect(textColorFor('#000000')).toBe('#ffffff')
    expect(textColorFor('#4f46e5')).toBe('#ffffff') // brand indigo
  })

  it('falls back to dark text on malformed input', () => {
    expect(textColorFor('nope')).toBe('#0f172a')
    expect(textColorFor('#abc')).toBe('#0f172a')
  })
})
