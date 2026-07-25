<script setup lang="ts">
import { RouterView } from 'vue-router'
import AppHeader from '@/components/AppHeader.vue'
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <AppHeader />
    <main class="flex-1">
      <RouterView />
    </main>
    <footer class="border-t border-slate-200 py-6">
      <p class="text-center text-sm text-slate-400">My hours — built with Go &amp; Vue</p>
    </footer>
  </div>
</template>
