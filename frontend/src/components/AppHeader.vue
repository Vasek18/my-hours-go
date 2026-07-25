<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import AppLogo from '@/components/AppLogo.vue'
import AppButton from '@/components/AppButton.vue'
import UserMenu from '@/components/UserMenu.vue'

const auth = useAuthStore()
const { isAuthenticated } = storeToRefs(auth)
</script>

<template>
  <header class="sticky top-0 z-10 border-b border-slate-200 bg-white/80 backdrop-blur">
    <div class="mx-auto flex h-16 max-w-5xl items-center justify-between px-4 sm:px-6">
      <RouterLink
        :to="isAuthenticated ? { name: 'dashboard' } : { name: 'home' }"
        class="rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-brand-500"
      >
        <AppLogo />
      </RouterLink>

      <nav class="flex items-center gap-1 sm:gap-2">
        <template v-if="isAuthenticated">
          <AppButton :to="{ name: 'dashboard' }" variant="ghost">Calendar</AppButton>
          <UserMenu />
        </template>
        <template v-else>
          <AppButton :to="{ name: 'login' }" variant="ghost">Sign in</AppButton>
          <AppButton :to="{ name: 'register' }" variant="primary">Get started</AppButton>
        </template>
      </nav>
    </div>
  </header>
</template>
