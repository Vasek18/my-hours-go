<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useActivityTypesStore, type ActivityType } from '@/stores/activityTypes'
import { ApiError } from '@/lib/api'
import AppButton from '@/components/AppButton.vue'
import AppAlert from '@/components/AppAlert.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'

const store = useActivityTypesStore()
const { items } = storeToRefs(store)

const loading = ref(true)
const loadError = ref('')

// --- Delete flow ---
const target = ref<ActivityType | null>(null)
const deleting = ref(false)
const deleteError = ref('')

onMounted(async () => {
  try {
    await store.fetchAll()
  } catch {
    loadError.value = 'Could not load your activity types. Please try again.'
  } finally {
    loading.value = false
  }
})

function askDelete(item: ActivityType) {
  deleteError.value = ''
  target.value = item
}

function cancelDelete() {
  target.value = null
}

async function confirmDelete() {
  if (!target.value) return
  deleting.value = true
  deleteError.value = ''
  try {
    await store.remove(target.value.id)
    target.value = null
  } catch (err) {
    deleteError.value =
      err instanceof ApiError ? err.message : 'Could not delete the activity type.'
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <section class="mx-auto max-w-5xl px-4 py-12 sm:px-6">
    <header class="mb-8 flex items-end justify-between gap-4">
      <div>
        <h1 class="text-3xl font-semibold tracking-tight text-slate-900">Activity types</h1>
        <p class="mt-1.5 text-slate-500">Organize your activities with a name and color.</p>
      </div>
      <AppButton :to="{ name: 'activity-type-new' }" variant="primary">New type</AppButton>
    </header>

    <AppAlert v-if="loadError" variant="error" class="mb-6">{{ loadError }}</AppAlert>
    <AppAlert v-if="deleteError" variant="error" class="mb-6">{{ deleteError }}</AppAlert>

    <div class="rounded-2xl bg-white shadow-sm ring-1 ring-slate-200">
      <p v-if="loading" class="px-6 py-10 text-center text-sm text-slate-500">Loading…</p>

      <div v-else-if="items.length === 0" class="px-6 py-12 text-center">
        <p class="text-sm text-slate-500">You don't have any activity types yet.</p>
        <AppButton :to="{ name: 'activity-type-new' }" variant="primary" class="mt-4">
          Create your first type
        </AppButton>
      </div>

      <ul v-else class="divide-y divide-slate-100">
        <li v-for="item in items" :key="item.id" class="flex items-center gap-4 px-6 py-4">
          <span
            class="inline-block h-9 w-9 shrink-0 rounded-full ring-1 ring-inset ring-slate-900/10"
            :style="{ backgroundColor: item.color }"
            aria-hidden="true"
          />
          <RouterLink
            :to="{ name: 'activity-type-edit', params: { id: item.id } }"
            class="min-w-0 flex-1 truncate font-medium text-slate-900 hover:text-brand-600"
          >
            {{ item.name }}
          </RouterLink>
          <div class="flex items-center gap-2">
            <AppButton
              :to="{ name: 'activity-type-edit', params: { id: item.id } }"
              variant="secondary"
            >
              Edit
            </AppButton>
            <AppButton variant="danger" @click="askDelete(item)">Delete</AppButton>
          </div>
        </li>
      </ul>
    </div>

    <ConfirmDialog
      :open="target !== null"
      title="Delete activity type"
      :message="target ? `Delete “${target.name}”? This can't be undone.` : ''"
      confirm-label="Delete"
      :loading="deleting"
      @confirm="confirmDelete"
      @cancel="cancelDelete"
    />
  </section>
</template>
