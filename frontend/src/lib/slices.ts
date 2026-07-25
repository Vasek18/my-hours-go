// Splitting an hour into equal slices, and diffing a working set of slices against
// the saved activities into a set of create/update/delete operations. Pure logic —
// the caller executes the plan against the API.

import type { Activity, ActivityPayload } from '@/stores/activities'

export const MAX_SLICES = 6 // 60 / 6 = 10-minute minimum slice

export interface Slice {
  id: string | null
  description: string
  typeId: string | null
}

export interface ReconcilePlan {
  deletes: string[]
  updates: { id: string; payload: ActivityPayload }[]
  creates: ActivityPayload[]
}

// Start/end of slice `index` when an hour starting at `base` is split into `count`
// equal blocks.
export function sliceBoundaries(
  base: Date,
  count: number,
  index: number,
): { start: Date; end: Date } {
  const sliceMs = count ? (60 / count) * 60_000 : 0
  return {
    start: new Date(base.getTime() + index * sliceMs),
    end: new Date(base.getTime() + (index + 1) * sliceMs),
  }
}

// Diff the desired slices (empty ones dropped) against the saved activities.
export function planReconcile(base: Date, sliceList: Slice[], original: Activity[]): ReconcilePlan {
  const finalSlices = sliceList.filter((s) => s.description.trim() !== '')
  const byId = new Map(original.map((a) => [a.id, a]))
  const keptIds = new Set(finalSlices.map((s) => s.id).filter((id): id is string => !!id))

  const plan: ReconcilePlan = {
    deletes: original.filter((a) => !keptIds.has(a.id)).map((a) => a.id),
    updates: [],
    creates: [],
  }

  finalSlices.forEach((s, i) => {
    const { start, end } = sliceBoundaries(base, finalSlices.length, i)
    const payload: ActivityPayload = {
      description: s.description.trim(),
      start_time: start.toISOString(),
      end_time: end.toISOString(),
      type_id: s.typeId,
    }
    if (!s.id) {
      plan.creates.push(payload)
      return
    }
    const orig = byId.get(s.id)
    const unchanged =
      orig &&
      orig.description === payload.description &&
      (orig.type_id ?? null) === s.typeId &&
      new Date(orig.start_time).getTime() === start.getTime() &&
      new Date(orig.end_time).getTime() === end.getTime()
    if (!unchanged) plan.updates.push({ id: s.id, payload })
  })

  return plan
}
