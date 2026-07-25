<script setup lang="ts">
import type { ViewMode } from '@/composables/useCalendarView'

defineProps<{ modelValue: ViewMode; modes: { key: ViewMode; label: string }[] }>()
defineEmits<{ (e: 'update:modelValue', value: ViewMode): void }>()
</script>

<template>
  <div class="inline-flex rounded-lg bg-slate-100 p-0.5 text-sm">
    <button
      v-for="m in modes"
      :key="m.key"
      type="button"
      class="rounded-md px-3 py-1.5 font-medium transition focus:outline-none"
      :class="modelValue === m.key ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500'"
      @click="$emit('update:modelValue', m.key)"
    >
      {{ m.label }}
    </button>
  </div>
</template>
