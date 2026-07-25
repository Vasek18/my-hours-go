// Aggregating activities into a per-type time summary. Pure logic.

import type { Activity } from '@/stores/activities'
import type { ActivityType } from '@/stores/activityTypes'

export interface SummaryRow {
  key: string
  name: string
  color: string
  ms: number
}

// Total time per type across `activities`, most spent first. Untyped activities
// are grouped under one row using `neutralColor`.
export function summarizeByType(
  activities: Activity[],
  typeById: Record<string, ActivityType>,
  neutralColor: string,
): SummaryRow[] {
  const totals: Record<string, number> = {}
  for (const a of activities) {
    const ms = new Date(a.end_time).getTime() - new Date(a.start_time).getTime()
    if (ms <= 0) continue
    const key = a.type_id ?? ''
    totals[key] = (totals[key] ?? 0) + ms
  }
  return Object.entries(totals)
    .map(([key, ms]) => {
      const type = key ? typeById[key] : undefined
      return {
        key: key || 'none',
        name: type?.name ?? 'No type',
        color: type?.color ?? neutralColor,
        ms,
      }
    })
    .sort((a, b) => b.ms - a.ms)
}

export function totalMs(rows: SummaryRow[]): number {
  return rows.reduce((sum, r) => sum + r.ms, 0)
}

export function formatDuration(ms: number): string {
  const totalMin = Math.round(ms / 60_000)
  const h = Math.floor(totalMin / 60)
  const m = totalMin % 60
  if (h && m) return `${h}h ${m}m`
  if (h) return `${h}h`
  return `${m}m`
}

export function percentOfTotal(ms: number, total: number): number {
  return total ? Math.round((ms / total) * 100) : 0
}
