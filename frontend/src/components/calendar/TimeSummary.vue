<script setup lang="ts">
import { computed } from 'vue'
import { formatDuration, percentOfTotal, totalMs, type SummaryRow } from '@/lib/summary'

const props = defineProps<{ rows: SummaryRow[]; label: string }>()
const total = computed(() => totalMs(props.rows))
</script>

<template>
  <div v-if="rows.length" class="mt-4 rounded-2xl bg-white p-4 shadow-sm ring-1 ring-slate-200">
    <p class="mb-3 text-xs font-medium uppercase tracking-wide text-slate-400">
      Time by type ({{ label }})
    </p>
    <ul class="space-y-2">
      <li v-for="r in rows" :key="r.key" class="flex items-center gap-3 text-sm">
        <span
          class="inline-block h-3 w-3 shrink-0 rounded-full ring-1 ring-inset ring-slate-900/10"
          :style="{ backgroundColor: r.color }"
          aria-hidden="true"
        />
        <span class="min-w-0 flex-1 truncate font-medium text-slate-700">{{ r.name }}</span>
        <span class="tabular-nums text-slate-500">{{ formatDuration(r.ms) }}</span>
        <span class="w-12 text-right tabular-nums text-slate-400"
          >{{ percentOfTotal(r.ms, total) }}%</span
        >
      </li>
    </ul>
  </div>
</template>
