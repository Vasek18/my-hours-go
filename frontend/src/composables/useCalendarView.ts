import { ref, computed, type Ref } from 'vue'
import {
  HOURS,
  startOfDay,
  addDays,
  weekStartOf,
  weekDays,
  monthMatrix,
  dayKey,
} from '@/lib/datetime'

export type ViewMode = 'day' | '3day' | 'week' | 'month'

export const VIEW_MODES: { key: ViewMode; label: string }[] = [
  { key: 'day', label: 'Day' },
  { key: '3day', label: '3 days' },
  { key: 'week', label: 'Week' },
  { key: 'month', label: 'Month' },
]

const fmtWeekday = new Intl.DateTimeFormat(undefined, { weekday: 'short' })
const fmtDayLong = new Intl.DateTimeFormat(undefined, {
  weekday: 'short',
  month: 'short',
  day: 'numeric',
})
const fmtMonthDay = new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric' })
const fmtMonthYear = new Intl.DateTimeFormat(undefined, { month: 'long', year: 'numeric' })

// Position, current view mode, visible range and labels for the calendar. On
// desktop the mode is always the full week; the switcher only applies on mobile.
export function useCalendarView(isMobile: Ref<boolean>) {
  const anchor = ref(startOfDay(new Date()))
  const viewMode = ref<ViewMode>('day')
  const mode = computed<ViewMode>(() => (isMobile.value ? viewMode.value : 'week'))

  const weekStart = computed(() => weekStartOf(anchor.value))
  const gridDays = computed(() => {
    switch (mode.value) {
      case 'day':
        return [startOfDay(anchor.value)]
      case '3day':
        return [0, 1, 2].map((i) => addDays(anchor.value, i))
      default:
        return weekDays(weekStart.value)
    }
  })
  const monthDays = computed(() => monthMatrix(anchor.value))
  const anchorMonth = computed(() => anchor.value.getMonth())

  // Hours run 23:00 → 00:00 so the usually-empty evening leads and sleep sinks.
  const hourRows = [...HOURS].reverse()

  const range = computed(() => {
    switch (mode.value) {
      case 'day':
        return { start: startOfDay(anchor.value), end: addDays(anchor.value, 1) }
      case '3day':
        return { start: startOfDay(anchor.value), end: addDays(anchor.value, 3) }
      case 'month': {
        const start = monthDays.value[0]!
        return { start, end: addDays(start, 42) }
      }
      default:
        return { start: weekStart.value, end: addDays(weekStart.value, 7) }
    }
  })
  const rangeKey = computed(() => `${range.value.start.getTime()}-${range.value.end.getTime()}`)

  const headerLabel = computed(() => {
    const a = startOfDay(anchor.value)
    switch (mode.value) {
      case 'day':
        return fmtDayLong.format(a)
      case '3day':
        return `${fmtMonthDay.format(a)} – ${fmtMonthDay.format(addDays(a, 2))}`
      case 'month':
        return fmtMonthYear.format(a)
      default:
        return `${fmtMonthDay.format(weekStart.value)} – ${fmtMonthDay.format(addDays(weekStart.value, 6))}`
    }
  })

  const todayKey = dayKey(new Date())
  const isToday = (d: Date) => dayKey(d) === todayKey
  const weekdayLabel = (d: Date) => fmtWeekday.format(d)

  function shift(dir: number) {
    if (mode.value === 'month') {
      anchor.value = new Date(anchor.value.getFullYear(), anchor.value.getMonth() + dir, 1)
      return
    }
    const step = mode.value === 'day' ? 1 : mode.value === '3day' ? 3 : 7
    anchor.value = addDays(anchor.value, step * dir)
  }
  const prev = () => shift(-1)
  const next = () => shift(1)
  const goToday = () => (anchor.value = startOfDay(new Date()))
  const setViewMode = (m: ViewMode) => (viewMode.value = m)
  // From the month overview, drill into a day.
  function openDay(d: Date) {
    anchor.value = startOfDay(d)
    viewMode.value = 'day'
  }

  return {
    anchor,
    viewMode,
    mode,
    gridDays,
    monthDays,
    anchorMonth,
    hourRows,
    range,
    rangeKey,
    headerLabel,
    isToday,
    weekdayLabel,
    prev,
    next,
    goToday,
    setViewMode,
    openDay,
  }
}
