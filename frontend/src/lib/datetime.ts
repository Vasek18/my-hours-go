// Local-timezone date helpers shared by the calendar. Pure and side-effect-free.

export const HOURS = Array.from({ length: 24 }, (_, h) => h)

export function startOfDay(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate())
}

export function startOfMonth(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), 1)
}

export function addDays(d: Date, n: number): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate() + n)
}

// Monday-based start of the week containing d.
export function weekStartOf(d: Date): Date {
  return addDays(d, -((d.getDay() + 6) % 7))
}

export function weekDays(weekStart: Date): Date[] {
  return Array.from({ length: 7 }, (_, i) => addDays(weekStart, i))
}

// 6-week (42-day) grid covering the month that d falls in.
export function monthMatrix(d: Date): Date[] {
  const gridStart = weekStartOf(startOfMonth(d))
  return Array.from({ length: 42 }, (_, i) => addDays(gridStart, i))
}

export function dayKey(d: Date): string {
  return `${d.getFullYear()}-${d.getMonth()}-${d.getDate()}`
}

// Calendar date as YYYY-MM-DD in local time (for the API's date-only fields).
export function toISODate(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

export function cellId(d: Date, hour: number): string {
  return `${dayKey(d)}-${hour}`
}

export function sameDay(a: Date, b: Date): boolean {
  return dayKey(a) === dayKey(b)
}
