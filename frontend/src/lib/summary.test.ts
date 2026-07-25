import { describe, it, expect } from 'vitest'
import { summarizeByType, totalMs, formatDuration, percentOfTotal } from './summary'
import type { Activity } from '@/stores/activities'
import type { ActivityType } from '@/stores/activityTypes'

const HOUR = 3_600_000
function act(typeId: string | null, hours: number, offset = 0): Activity {
  const start = new Date(2026, 6, 8, 0).getTime() + offset * HOUR
  return {
    id: Math.random().toString(),
    description: 'x',
    type_id: typeId,
    start_time: new Date(start).toISOString(),
    end_time: new Date(start + hours * HOUR).toISOString(),
  }
}
const types: Record<string, ActivityType> = {
  sleep: { id: 'sleep', name: 'Sleep', color: '#111111', sort: 1 },
  work: { id: 'work', name: 'Work', color: '#222222', sort: 2 },
}

describe('summarizeByType', () => {
  it('sums per type, most spent first, untyped grouped', () => {
    const rows = summarizeByType(
      [act('work', 2, 0), act('sleep', 3, 2), act('work', 1, 5), act(null, 1, 6)],
      types,
      '#f1f5f9',
    )
    expect(rows.map((r) => [r.name, r.ms / HOUR])).toEqual([
      ['Work', 3],
      ['Sleep', 3],
      ['No type', 1],
    ])
    // Ties keep insertion-ish order; the untyped row uses the neutral color.
    expect(rows[2]!.color).toBe('#f1f5f9')
  })

  it('ignores non-positive durations', () => {
    const rows = summarizeByType([act('work', 0, 0)], types, '#f1f5f9')
    expect(rows).toHaveLength(0)
  })
})

describe('formatDuration', () => {
  it('formats hours and minutes', () => {
    expect(formatDuration(8 * HOUR)).toBe('8h')
    expect(formatDuration(90 * 60_000)).toBe('1h 30m')
    expect(formatDuration(45 * 60_000)).toBe('45m')
  })
})

describe('percentOfTotal', () => {
  it('rounds and guards divide-by-zero', () => {
    expect(percentOfTotal(HOUR, 4 * HOUR)).toBe(25)
    expect(percentOfTotal(HOUR, 0)).toBe(0)
  })
})

describe('totalMs', () => {
  it('sums row durations', () => {
    expect(
      totalMs([
        { key: 'a', name: 'A', color: '#000', ms: HOUR },
        { key: 'b', name: 'B', color: '#000', ms: 2 * HOUR },
      ]),
    ).toBe(3 * HOUR)
  })
})
