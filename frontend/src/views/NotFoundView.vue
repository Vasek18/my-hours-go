<script setup lang="ts">
import AppButton from '@/components/AppButton.vue'
</script>

<template>
  <section class="mx-auto max-w-5xl px-4 py-24 text-center sm:px-6">
    <p class="text-sm font-semibold text-brand-600">404</p>
    <h1 class="mt-2 text-4xl font-bold tracking-tight text-slate-900">Page not found</h1>
    <p class="mt-3 text-slate-600">The page you're looking for doesn't exist.</p>
    <div class="mt-8">
      <AppButton :to="{ name: 'home' }" variant="primary">Back home</AppButton>
    </div>
  </section>
</template>
