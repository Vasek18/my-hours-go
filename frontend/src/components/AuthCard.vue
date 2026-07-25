<script setup lang="ts">
defineProps<{ title: string; subtitle?: string }>()
</script>

<template>
  <div class="mx-auto flex w-full max-w-md flex-col px-4 py-12 sm:py-16">
    <div class="rounded-2xl bg-white p-8 shadow-sm ring-1 ring-slate-200">
      <h1 class="text-2xl font-semibold tracking-tight text-slate-900">{{ title }}</h1>
      <p v-if="subtitle" class="mt-1.5 text-sm text-slate-500">{{ subtitle }}</p>

      <div class="mt-6">
        <slot />
      </div>
    </div>

    <div v-if="$slots.footer" class="mt-6 text-center text-sm text-slate-500">
      <slot name="footer" />
    </div>
  </div>
</template>
