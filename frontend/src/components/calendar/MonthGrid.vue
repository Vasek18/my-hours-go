<script setup lang="ts">
import type { Activity } from '@/stores/activities'
import type { ActivityType } from '@/stores/activityTypes'
import { dayKey } from '@/lib/datetime'
import { textColorFor } from '@/lib/contrast'

const props = defineProps<{
  days: Date[]
  month: number
  activitiesByDay: Record<string, Activity[]>
  typeById: Record<string, ActivityType>
  neutralColor: string
  weekdayLabel: (d: Date) => string
  isToday: (d: Date) => boolean
  chipMax?: number
}>()
defineEmits<{ (e: 'open-day', d: Date): void }>()

const max = props.chipMax ?? 3

function dayActivities(d: Date): Activity[] {
  return props.activitiesByDay[dayKey(d)] ?? []
}
function chipStyle(a: Activity) {
  const bg = (a.type_id && props.typeById[a.type_id]?.color) || props.neutralColor
  return { backgroundColor: bg, color: textColorFor(bg) }
}
</script>

<template>
  <div class="overflow-hidden rounded-2xl bg-white shadow-sm ring-1 ring-slate-200">
    <div class="grid grid-cols-7 border-b border-slate-200">
      <div
        v-for="d in days.slice(0, 7)"
        :key="'wl-' + d.getDay()"
        class="px-1 py-2 text-center text-xs font-medium uppercase tracking-wide text-slate-400"
      >
        {{ weekdayLabel(d) }}
      </div>
    </div>
    <div class="grid grid-cols-7">
      <button
        v-for="d in days"
        :key="dayKey(d)"
        type="button"
        class="min-h-[5rem] border-b border-l border-slate-100 p-1 text-left align-top transition hover:bg-slate-50 focus:outline-none"
        @click="$emit('open-day', d)"
      >
        <span
          class="inline-flex h-6 w-6 items-center justify-center rounded-full text-xs"
          :class="
            isToday(d)
              ? 'bg-brand-600 font-semibold text-white'
              : d.getMonth() === month
                ? 'text-slate-900'
                : 'text-slate-300'
          "
        >
          {{ d.getDate() }}
        </span>
        <div class="mt-1 space-y-0.5">
          <span
            v-for="a in dayActivities(d).slice(0, max)"
            :key="a.id"
            class="block truncate rounded px-1 py-0.5 text-[11px] leading-tight"
            :style="chipStyle(a)"
          >
            {{ a.description }}
          </span>
          <span v-if="dayActivities(d).length > max" class="block px-1 text-[11px] text-slate-400">
            +{{ dayActivities(d).length - max }} more
          </span>
        </div>
      </button>
    </div>
  </div>
</template>
