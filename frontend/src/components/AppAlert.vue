<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{ variant?: 'error' | 'success' | 'info' }>(), {
  variant: 'info',
})

const classes = computed(() => {
  const variants: Record<string, string> = {
    error: 'bg-red-50 text-red-700 ring-red-200',
    success: 'bg-emerald-50 text-emerald-700 ring-emerald-200',
    info: 'bg-brand-50 text-brand-700 ring-brand-200',
  }
  return variants[props.variant]
})
</script>

<template>
  <div class="rounded-lg px-4 py-3 text-sm ring-1 ring-inset" :class="classes" role="alert">
    <slot />
  </div>
</template>
