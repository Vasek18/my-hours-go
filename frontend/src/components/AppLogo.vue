<script setup lang="ts">
withDefaults(defineProps<{ size?: number }>(), { size: 32 })
</script>

<template>
  <span class="inline-flex items-center gap-2">
    <svg :width="size" :height="size" viewBox="0 0 32 32" aria-hidden="true" class="shrink-0">
      <rect width="32" height="32" rx="8" class="fill-brand-600" />
      <path
        d="M9 16a7 7 0 1 1 12.5 4.3"
        fill="none"
        stroke="#fff"
        stroke-width="3"
        stroke-linecap="round"
      />
      <circle cx="22" cy="21" r="2.6" class="fill-brand-300" />
    </svg>
    <span class="text-lg font-semibold tracking-tight text-slate-900">My hours</span>
  </span>
</template>
