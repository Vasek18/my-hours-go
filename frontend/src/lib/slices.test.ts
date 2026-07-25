import { describe, it, expect } from 'vitest'
import { sliceBoundaries, planReconcile, type Slice } from './slices'
import type { Activity } from '@/stores/activities'

const base = new Date(2026, 6, 8, 9) // 09:00 local

function activity(over: Partial<Activity>): Activity {
  return {
    id: 'x',
    description: 'X',
    start_time: base.toISOString(),
    end_time: new Date(base.getTime() + 3_600_000).toISOString(),
    type_id: null,
    ...over,
  }
}

describe('sliceBoundaries', () => {
  it('splits the hour into equal blocks', () => {
    const { start, end } = sliceBoundaries(base, 3, 1) // middle of 3
    expect(start.getTime()).toBe(base.getTime() + 20 * 60_000)
    expect(end.getTime()).toBe(base.getTime() + 40 * 60_000)
  })
})

describe('planReconcile', () => {
  it('creates a single full-hour activity from an empty cell', () => {
    const slices: Slice[] = [{ id: null, description: 'Run', typeId: null }]
    const plan = planReconcile(base, slices, [])
    expect(plan.creates).toHaveLength(1)
    expect(plan.updates).toHaveLength(0)
    expect(plan.deletes).toHaveLength(0)
    expect(plan.creates[0]!.end_time).toBe(new Date(base.getTime() + 3_600_000).toISOString())
  })

  it('splitting one activity into two updates the original and creates the new', () => {
    const orig = activity({ id: 'a', description: 'Work' })
    const slices: Slice[] = [
      { id: 'a', description: 'Work', typeId: null },
      { id: null, description: 'Nap', typeId: null },
    ]
    const plan = planReconcile(base, slices, [orig])
    expect(plan.creates).toHaveLength(1)
    expect(plan.updates).toHaveLength(1) // original resized to 30m
    expect(plan.updates[0]!.payload.end_time).toBe(
      new Date(base.getTime() + 30 * 60_000).toISOString(),
    )
    expect(plan.deletes).toHaveLength(0)
  })

  it('clearing text deletes and the survivor is resized to the full hour', () => {
    const a = activity({
      id: 'a',
      description: 'Work',
      end_time: new Date(base.getTime() + 30 * 60_000).toISOString(), // was a 30m slice
    })
    const b = activity({
      id: 'b',
      description: 'Nap',
      start_time: new Date(base.getTime() + 30 * 60_000).toISOString(),
      end_time: new Date(base.getTime() + 60 * 60_000).toISOString(),
    })
    const slices: Slice[] = [
      { id: 'a', description: 'Work', typeId: null },
      { id: 'b', description: '', typeId: null }, // cleared
    ]
    const plan = planReconcile(base, slices, [a, b])
    expect(plan.deletes).toEqual(['b'])
    expect(plan.updates).toHaveLength(1)
    expect(plan.updates[0]!.payload.end_time).toBe(
      new Date(base.getTime() + 60 * 60_000).toISOString(),
    )
  })

  it('leaves an unchanged activity untouched', () => {
    const orig = activity({ id: 'a', description: 'Work' })
    const slices: Slice[] = [{ id: 'a', description: 'Work', typeId: null }]
    const plan = planReconcile(base, slices, [orig])
    expect(plan).toEqual({ deletes: [], updates: [], creates: [] })
  })
})
