import { describe, it, expect } from 'vitest'
import {
  addDays,
  weekStartOf,
  weekDays,
  monthMatrix,
  dayKey,
  cellId,
  sameDay,
  toISODate,
} from './datetime'

describe('datetime', () => {
  it('weekStartOf returns the Monday of the week', () => {
    // 2026-07-08 is a Wednesday.
    const monday = weekStartOf(new Date(2026, 6, 8))
    expect(dayKey(monday)).toBe(dayKey(new Date(2026, 6, 6)))
    // A Monday maps to itself.
    expect(dayKey(weekStartOf(new Date(2026, 6, 6)))).toBe(dayKey(new Date(2026, 6, 6)))
    // A Sunday maps back to the preceding Monday.
    expect(dayKey(weekStartOf(new Date(2026, 6, 12)))).toBe(dayKey(new Date(2026, 6, 6)))
  })

  it('weekDays yields 7 consecutive days from the start', () => {
    const days = weekDays(new Date(2026, 6, 6))
    expect(days).toHaveLength(7)
    expect(dayKey(days[6]!)).toBe(dayKey(new Date(2026, 6, 12)))
  })

  it('monthMatrix covers 42 days aligned to Mondays', () => {
    const m = monthMatrix(new Date(2026, 6, 15))
    expect(m).toHaveLength(42)
    expect(m[0]!.getDay()).toBe(1) // Monday
    expect(dayKey(m[0]!)).toBe(dayKey(new Date(2026, 5, 29))) // Mon before Jul 1
  })

  it('addDays crosses month boundaries', () => {
    expect(dayKey(addDays(new Date(2026, 6, 31), 1))).toBe(dayKey(new Date(2026, 7, 1)))
  })

  it('toISODate zero-pads to YYYY-MM-DD', () => {
    expect(toISODate(new Date(2026, 6, 8))).toBe('2026-07-08')
    expect(toISODate(new Date(2026, 11, 31))).toBe('2026-12-31')
  })

  it('cellId and sameDay are consistent', () => {
    const d = new Date(2026, 6, 8, 9)
    expect(cellId(d, 9)).toBe(`${dayKey(d)}-9`)
    expect(sameDay(new Date(2026, 6, 8, 1), new Date(2026, 6, 8, 23))).toBe(true)
    expect(sameDay(new Date(2026, 6, 8), new Date(2026, 6, 9))).toBe(false)
  })
})
