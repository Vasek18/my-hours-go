import { ref, computed, nextTick, onMounted, onBeforeUnmount, type ComputedRef } from 'vue'
import { useActivitiesStore, type Activity } from '@/stores/activities'
import type { ActivityType } from '@/stores/activityTypes'
import { ApiError } from '@/lib/api'
import { textColorFor } from '@/lib/contrast'
import { MAX_SLICES, planReconcile, type Slice } from '@/lib/slices'

interface EditorOptions {
  activities: ReturnType<typeof useActivitiesStore>
  hourActivities: (d: Date, hour: number) => Activity[]
  range: ComputedRef<{ start: Date; end: Date }>
  typeById: ComputedRef<Record<string, ActivityType>>
  neutralColor: string
}

function cellStart(d: Date, hour: number): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate(), hour)
}

// Editing state for one hour: a working set of equal slices, saved by diffing
// against the stored activities. The popover and hour grid drive it.
export function useHourEditor({
  activities,
  hourActivities,
  range,
  typeById,
  neutralColor,
}: EditorOptions) {
  const active = ref<{ date: Date; hour: number } | null>(null)
  const slices = ref<Slice[]>([])
  const focusIndex = ref(0)
  const original = ref<Activity[]>([])
  const draft = ref('')
  const busy = ref(false)
  const error = ref('')
  const internalShift = ref(false) // suppress commit while moving focus between slices

  const inputRef = ref<HTMLInputElement | null>(null)
  const anchorEl = ref<HTMLElement | null>(null)
  const panelEl = ref<HTMLElement | null>(null)
  const panelPos = ref<{ top: number; left: number; width: number } | null>(null)

  function setInputRef(el: unknown) {
    if (el) inputRef.value = el as HTMLInputElement
  }

  function updatePanelPos() {
    const el = anchorEl.value
    if (!el) return
    const r = el.getBoundingClientRect()
    const width = 288
    const gap = 6
    const h = panelEl.value?.offsetHeight ?? 160
    const left = Math.max(8, Math.min(r.left, window.innerWidth - width - 8))
    let top = r.bottom + gap
    if (top + h > window.innerHeight - 8) {
      const above = r.top - gap - h
      top = above >= 8 ? above : Math.max(8, window.innerHeight - 8 - h)
    }
    panelPos.value = { top, left, width }
  }
  // Position, then re-measure once the popover has rendered at its real height.
  function positionPanelSoon() {
    updatePanelPos()
    nextTick(updatePanelPos)
  }
  function onReposition() {
    if (active.value) updatePanelPos()
  }

  const focusedHasText = computed(() => draft.value.trim() !== '')
  const canCopyUp = computed(() => {
    if (!active.value) return false
    const h = active.value.hour
    return h < 23 && hourActivities(active.value.date, h + 1).length < MAX_SLICES
  })

  function isActiveCell(d: Date, hour: number): boolean {
    return (
      !!active.value && active.value.date.getTime() === d.getTime() && active.value.hour === hour
    )
  }
  function sliceStyle(s: Slice) {
    const bg = (s.typeId && typeById.value[s.typeId]?.color) || neutralColor
    return { backgroundColor: bg, color: textColorFor(bg) }
  }

  function openCell(event: MouseEvent, d: Date, hour: number, index = 0) {
    if (busy.value) return
    anchorEl.value = (event.currentTarget as HTMLElement).closest(
      '[data-cell]',
    ) as HTMLElement | null
    const list = hourActivities(d, hour)
    active.value = { date: d, hour }
    original.value = list
    slices.value = list.length
      ? list.map((a) => ({ id: a.id, description: a.description, typeId: a.type_id }))
      : [{ id: null, description: '', typeId: null }]
    focusIndex.value = Math.min(Math.max(index, 0), slices.value.length - 1)
    draft.value = slices.value[focusIndex.value]!.description
    nextTick(() => {
      inputRef.value?.focus()
      positionPanelSoon()
    })
  }
  function closeEditor() {
    active.value = null
    slices.value = []
    original.value = []
    focusIndex.value = 0
    anchorEl.value = null
    panelPos.value = null
  }
  const cancel = closeEditor

  function shiftFocus(to: number) {
    internalShift.value = true
    focusIndex.value = to
    draft.value = slices.value[to]!.description
    nextTick(() => {
      inputRef.value?.focus()
      internalShift.value = false
      positionPanelSoon()
    })
  }
  function focusSlice(i: number) {
    if (i === focusIndex.value) return
    slices.value[focusIndex.value]!.description = draft.value.trim()
    shiftFocus(i)
  }
  function addSlice(offset: number) {
    if (slices.value.length >= MAX_SLICES) return
    slices.value[focusIndex.value]!.description = draft.value.trim()
    const at = focusIndex.value + offset
    slices.value.splice(at, 0, { id: null, description: '', typeId: null })
    draft.value = ''
    shiftFocus(at)
  }
  const addBefore = () => addSlice(0)
  const addAfter = () => addSlice(1)

  function copyAfter() {
    if (slices.value.length >= MAX_SLICES) return
    const cur = slices.value[focusIndex.value]!
    const text = draft.value.trim() || cur.description
    if (!text) return
    cur.description = text
    slices.value.splice(focusIndex.value + 1, 0, {
      id: null,
      description: text,
      typeId: cur.typeId,
    })
    nextTick(updatePanelPos)
  }

  async function applyPlan(date: Date, hour: number, sliceList: Slice[], originalList: Activity[]) {
    const plan = planReconcile(cellStart(date, hour), sliceList, originalList)
    for (const id of plan.deletes) await activities.remove(id)
    for (const u of plan.updates) await activities.update(u.id, u.payload)
    for (const p of plan.creates) await activities.create(p)
  }

  async function persist(): Promise<boolean> {
    if (!active.value) return false
    slices.value[focusIndex.value]!.description = draft.value.trim()
    busy.value = true
    error.value = ''
    try {
      await applyPlan(active.value.date, active.value.hour, slices.value, original.value)
      await activities.fetchRange(range.value.start, range.value.end)
      return true
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Could not save the activity.'
      return false
    } finally {
      busy.value = false
    }
  }

  async function commit() {
    if (!active.value || busy.value || internalShift.value) return
    if (await persist()) closeEditor()
  }

  async function applyType(type: ActivityType) {
    if (!active.value || busy.value) return
    const s = slices.value[focusIndex.value]!
    s.typeId = type.id
    if (!draft.value.trim()) draft.value = type.name
    if (await persist()) closeEditor()
  }

  // Copy the focused activity into the hour above (hour + 1), appended and re-split.
  async function copyUp() {
    if (!active.value || busy.value) return
    const cur = slices.value[focusIndex.value]!
    const text = draft.value.trim() || cur.description
    if (!text) return
    const { date, hour } = active.value
    const targetHour = hour + 1
    if (targetHour > 23) return
    const targetList = hourActivities(date, targetHour)
    if (targetList.length >= MAX_SLICES) return

    cur.description = text
    busy.value = true
    error.value = ''
    try {
      await applyPlan(date, hour, slices.value, original.value)
      const combined: Slice[] = [
        ...targetList.map((a) => ({ id: a.id, description: a.description, typeId: a.type_id })),
        { id: null, description: text, typeId: cur.typeId },
      ]
      await applyPlan(date, targetHour, combined, targetList)
      await activities.fetchRange(range.value.start, range.value.end)
      closeEditor()
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Could not copy the activity.'
    } finally {
      busy.value = false
    }
  }

  onMounted(() => {
    // capture = true catches the inner scroll container (its scroll doesn't bubble)
    window.addEventListener('scroll', onReposition, true)
    window.addEventListener('resize', onReposition)
  })
  onBeforeUnmount(() => {
    window.removeEventListener('scroll', onReposition, true)
    window.removeEventListener('resize', onReposition)
  })

  return {
    active,
    slices,
    focusIndex,
    draft,
    busy,
    error,
    panelPos,
    panelEl,
    setInputRef,
    focusedHasText,
    canCopyUp,
    isActiveCell,
    sliceStyle,
    openCell,
    closeEditor,
    cancel,
    focusSlice,
    addBefore,
    addAfter,
    copyAfter,
    copyUp,
    commit,
    applyType,
    MAX_SLICES,
  }
}
