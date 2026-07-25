<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { RouterLink } from 'vue-router'
import { useActivitiesStore, type Activity } from '@/stores/activities'
import { useActivityTypesStore } from '@/stores/activityTypes'
import { useDayNotesStore } from '@/stores/dayNotes'
import { cellId, dayKey, toISODate } from '@/lib/datetime'
import { summarizeByType } from '@/lib/summary'
import { textColorFor } from '@/lib/contrast'
import { useMediaQuery } from '@/composables/useMediaQuery'
import { useCalendarView, VIEW_MODES } from '@/composables/useCalendarView'
import { useHourEditor } from '@/composables/useHourEditor'
import AppAlert from '@/components/AppAlert.vue'
import ViewSwitcher from '@/components/calendar/ViewSwitcher.vue'
import MonthGrid from '@/components/calendar/MonthGrid.vue'
import TimeSummary from '@/components/calendar/TimeSummary.vue'
import DayNoteDialog from '@/components/calendar/DayNoteDialog.vue'

const NEUTRAL = '#f1f5f9' // slate-100, used when an activity has no type

const activities = useActivitiesStore()
const types = useActivityTypesStore()
const dayNotes = useDayNotesStore()
const { items: activityItems } = storeToRefs(activities)
const { items: typeItems } = storeToRefs(types)
const { byDate: noteByDate } = storeToRefs(dayNotes)

const typeById = computed(() => Object.fromEntries(typeItems.value.map((t) => [t.id, t] as const)))
function cellStyle(a: Activity) {
  const bg = (a.type_id && typeById.value[a.type_id]?.color) || NEUTRAL
  return { backgroundColor: bg, color: textColorFor(bg) }
}
const hourMap = computed(() => {
  const m: Record<string, Activity[]> = {}
  for (const a of activityItems.value) {
    const s = new Date(a.start_time)
    ;(m[cellId(s, s.getHours())] ??= []).push(a)
  }
  for (const k in m) {
    m[k]!.sort((a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime())
  }
  return m
})
const activitiesByDay = computed(() => {
  const m: Record<string, Activity[]> = {}
  for (const a of activityItems.value) (m[dayKey(new Date(a.start_time))] ??= []).push(a)
  return m
})
function hourActivities(d: Date, hour: number): Activity[] {
  return hourMap.value[cellId(d, hour)] ?? []
}

const isMobile = useMediaQuery('(max-width: 639.98px)')

const view = useCalendarView(isMobile)
const editor = useHourEditor({
  activities,
  hourActivities,
  range: view.range,
  typeById,
  neutralColor: NEUTRAL,
})

const { mode, gridDays, monthDays, anchorMonth, hourRows, range, rangeKey, headerLabel, isToday, weekdayLabel, viewMode, prev, next, goToday, setViewMode, openDay } = view // prettier-ignore
const { active, slices, focusIndex, draft, panelPos, panelEl, setInputRef, focusedHasText, canCopyUp, isActiveCell, sliceStyle, openCell, closeEditor, cancel, focusSlice, addBefore, addAfter, copyAfter, copyUp, commit, applyType, error: saveError, MAX_SLICES } = editor // prettier-ignore

const typeSummary = computed(() => summarizeByType(activityItems.value, typeById.value, NEUTRAL))
const gridStyle = computed(() => ({
  gridTemplateColumns: `3.5rem repeat(${gridDays.value.length}, minmax(0, 1fr))`,
}))
const hourLabel = (hour: number) => `${String(hour).padStart(2, '0')}:00`

// --- day notes (diary) ---
const noteDate = ref<Date | null>(null)
const noteSaving = ref(false)
const noteError = ref('')
const noteContent = computed(() =>
  noteDate.value ? (noteByDate.value[toISODate(noteDate.value)] ?? '') : '',
)
const hasNote = (d: Date) => !!noteByDate.value[toISODate(d)]
function openNote(d: Date) {
  noteError.value = ''
  noteDate.value = d
}
async function saveNote(content: string) {
  if (!noteDate.value) return
  noteSaving.value = true
  noteError.value = ''
  try {
    await dayNotes.save(toISODate(noteDate.value), content)
    noteDate.value = null
  } catch {
    noteError.value = 'Could not save the note. Please try again.'
  } finally {
    noteSaving.value = false
  }
}

const loading = ref(true)
const loadError = ref('')
async function load() {
  loading.value = true
  loadError.value = ''
  try {
    await activities.fetchRange(range.value.start, range.value.end)
  } catch {
    loadError.value = 'Could not load your activities. Please try again.'
  } finally {
    loading.value = false
  }
  try {
    await dayNotes.fetchRange(toISODate(range.value.start), toISODate(range.value.end))
  } catch {
    // Notes are secondary; the calendar still works without them.
  }
}

// Any navigation changes the range; close the editor and reload for the new span.
watch(rangeKey, () => {
  closeEditor()
  load()
})
onMounted(async () => {
  try {
    await types.fetchAll()
  } catch {
    // Chips just won't show; the calendar still works.
  }
  await load()
})
</script>

<template>
  <section class="mx-auto max-w-7xl px-2 py-6 sm:px-4">
    <header class="mb-4 px-1">
      <div class="flex items-center justify-between gap-3">
        <h1 class="text-xl font-semibold tracking-tight text-slate-900 sm:text-2xl">
          {{ headerLabel }}
        </h1>
        <div class="flex items-center gap-1">
          <button
            type="button"
            class="rounded-lg px-2.5 py-1.5 text-slate-500 hover:bg-slate-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-brand-500"
            aria-label="Previous"
            @click="prev"
          >
            ‹
          </button>
          <button
            type="button"
            class="rounded-lg px-3 py-1.5 text-sm font-medium text-slate-700 ring-1 ring-slate-300 hover:bg-slate-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-brand-500"
            @click="goToday"
          >
            Today
          </button>
          <button
            type="button"
            class="rounded-lg px-2.5 py-1.5 text-slate-500 hover:bg-slate-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-brand-500"
            aria-label="Next"
            @click="next"
          >
            ›
          </button>
        </div>
      </div>

      <ViewSwitcher
        v-if="isMobile"
        class="mt-3"
        :model-value="viewMode"
        :modes="VIEW_MODES"
        @update:model-value="setViewMode"
      />
    </header>

    <AppAlert v-if="saveError || loadError || noteError" variant="error" class="mb-3">
      {{ saveError || loadError || noteError }}
    </AppAlert>

    <MonthGrid
      v-if="mode === 'month'"
      :days="monthDays"
      :month="anchorMonth"
      :activities-by-day="activitiesByDay"
      :type-by-id="typeById"
      :neutral-color="NEUTRAL"
      :weekday-label="weekdayLabel"
      :is-today="isToday"
      @open-day="openDay"
    />

    <div v-else class="overflow-hidden rounded-2xl bg-white shadow-sm ring-1 ring-slate-200">
      <div class="grid border-b border-slate-200" :style="gridStyle">
        <div class="py-2" />
        <div v-for="d in gridDays" :key="'h-' + dayKey(d)" class="px-1 py-2 text-center">
          <p class="text-xs font-medium uppercase tracking-wide text-slate-400">
            {{ weekdayLabel(d) }}
          </p>
          <div class="mt-0.5 flex items-center justify-center gap-1">
            <span
              class="text-sm font-semibold"
              :class="isToday(d) ? 'text-brand-600' : 'text-slate-900'"
            >
              {{ d.getDate() }}
            </span>
            <button
              type="button"
              class="rounded p-0.5 transition hover:bg-slate-100"
              :class="hasNote(d) ? 'text-brand-600' : 'text-slate-300 hover:text-slate-500'"
              :aria-label="hasNote(d) ? 'Edit day notes' : 'Add day notes'"
              @click="openNote(d)"
            >
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                class="h-4 w-4"
                aria-hidden="true"
              >
                <path d="M9 4H6a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-3" />
                <path d="M18.5 2.5a2.12 2.12 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      <div class="max-h-[68vh] overflow-y-auto">
        <div
          v-for="hour in hourRows"
          :key="hour"
          class="grid border-b border-slate-100 last:border-b-0"
          :style="gridStyle"
        >
          <div class="py-1 pr-2 text-right text-xs tabular-nums text-slate-400">
            {{ hourLabel(hour) }}
          </div>
          <div
            v-for="d in gridDays"
            :key="dayKey(d) + '-' + hour"
            data-cell
            class="min-h-[3rem] border-l border-slate-100 p-1"
          >
            <div v-if="isActiveCell(d, hour)" class="flex h-full min-h-[2.5rem] gap-1">
              <div v-for="(s, i) in slices" :key="i" class="relative min-w-0 flex-1">
                <input
                  v-if="i === focusIndex"
                  :ref="setInputRef"
                  v-model="draft"
                  type="text"
                  placeholder="Activity…"
                  class="h-full min-h-[2.5rem] w-full rounded-md border-0 bg-white px-2 py-1 text-sm text-slate-900 ring-1 ring-inset ring-brand-500 placeholder:text-slate-400 focus:outline-none"
                  @blur="commit"
                  @keydown.enter.prevent="commit"
                  @keydown.esc.prevent="cancel"
                />
                <button
                  v-else
                  type="button"
                  class="flex h-full min-h-[2.5rem] w-full items-center justify-center rounded-md px-1 text-xs shadow-sm"
                  :style="sliceStyle(s)"
                  @mousedown.prevent
                  @click="focusSlice(i)"
                >
                  <span class="truncate">{{ s.description }}</span>
                </button>
              </div>
            </div>

            <div
              v-else-if="hourActivities(d, hour).length"
              class="flex h-full min-h-[2.5rem] gap-1"
            >
              <button
                v-for="(a, i) in hourActivities(d, hour)"
                :key="a.id"
                type="button"
                class="flex min-w-0 flex-1 items-center justify-center rounded-md px-1 text-xs shadow-sm"
                :style="cellStyle(a)"
                @click="openCell($event, d, hour, i)"
              >
                <span class="truncate">{{ a.description }}</span>
              </button>
            </div>

            <button
              v-else
              type="button"
              class="h-full min-h-[2.5rem] w-full rounded-md transition hover:bg-slate-50"
              aria-label="Add activity"
              @click="openCell($event, d, hour, 0)"
            />
          </div>
        </div>
      </div>
    </div>

    <p v-if="!loading && mode !== 'month'" class="mt-2 px-1 text-xs text-slate-400">
      Click any slot to add an activity. Pick a type to color it. Clear the text to delete.
    </p>
    <p v-else-if="!loading" class="mt-2 px-1 text-xs text-slate-400">
      Tap a day to open it and add activities.
    </p>

    <TimeSummary :rows="typeSummary" :label="headerLabel" />

    <DayNoteDialog
      :open="noteDate !== null"
      :date="noteDate"
      :content="noteContent"
      :saving="noteSaving"
      @save="saveNote"
      @close="noteDate = null"
    />

    <Teleport to="body">
      <div
        v-if="active && panelPos"
        ref="panelEl"
        class="fixed z-40"
        :style="{
          top: panelPos.top + 'px',
          left: panelPos.left + 'px',
          width: panelPos.width + 'px',
        }"
        @mousedown.prevent
      >
        <div class="rounded-xl bg-white p-2 shadow-lg ring-1 ring-slate-200">
          <div
            v-if="focusedHasText"
            class="mb-2 grid grid-cols-2 gap-1.5 border-b border-slate-100 pb-2 [&>button]:rounded-lg [&>button]:px-2 [&>button]:py-1.5 [&>button]:text-xs [&>button]:font-medium [&>button]:text-slate-700 [&>button]:ring-1 [&>button]:ring-slate-300 [&>button:hover]:bg-slate-50 [&>button:disabled]:cursor-not-allowed [&>button:disabled]:opacity-40"
          >
            <button type="button" :disabled="slices.length >= MAX_SLICES" @click="addBefore">
              + Before
            </button>
            <button type="button" :disabled="slices.length >= MAX_SLICES" @click="addAfter">
              + After
            </button>
            <button type="button" :disabled="slices.length >= MAX_SLICES" @click="copyAfter">
              Copy →
            </button>
            <button type="button" :disabled="!canCopyUp" @click="copyUp">Copy ↑</button>
          </div>

          <div v-if="typeItems.length" class="flex max-h-56 flex-wrap gap-1.5 overflow-y-auto">
            <button
              v-for="t in typeItems"
              :key="t.id"
              type="button"
              class="rounded-full px-2.5 py-1 text-xs font-medium shadow-sm transition"
              :style="{ backgroundColor: t.color, color: textColorFor(t.color) }"
              @click="applyType(t)"
            >
              {{ t.name }}
            </button>
          </div>
          <p v-else class="px-1 py-0.5 text-xs text-slate-500">
            No activity types.
            <RouterLink :to="{ name: 'activity-types' }" class="font-medium text-brand-600">
              Create some
            </RouterLink>
          </p>
        </div>
      </div>
    </Teleport>
  </section>
</template>
