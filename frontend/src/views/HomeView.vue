<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import AppButton from '@/components/AppButton.vue'

const auth = useAuthStore()
const { isAuthenticated } = storeToRefs(auth)
</script>

<template>
  <section class="mx-auto max-w-5xl px-4 py-24 sm:px-6 sm:py-32">
    <div class="mx-auto max-w-2xl text-center">
      <span
        class="inline-flex items-center rounded-full bg-brand-50 px-3 py-1 text-xs font-medium text-brand-700 ring-1 ring-inset ring-brand-200"
      >
        Go · Vue · Postgres · Redis
      </span>

      <h1 class="mt-6 text-5xl font-bold tracking-tight text-slate-900 sm:text-6xl">My hours</h1>

      <p class="mt-5 text-lg leading-relaxed text-slate-600">
        Track where your time goes — log activities on a simple weekly calendar and see your hours
        by type at a glance.
      </p>

      <div class="mt-10 flex items-center justify-center gap-3">
        <template v-if="isAuthenticated">
          <AppButton :to="{ name: 'dashboard' }" variant="primary">Open calendar</AppButton>
        </template>
        <template v-else>
          <AppButton :to="{ name: 'register' }" variant="primary">Get started</AppButton>
          <AppButton :to="{ name: 'login' }" variant="secondary">Sign in</AppButton>
        </template>
      </div>
    </div>
  </section>
</template>
