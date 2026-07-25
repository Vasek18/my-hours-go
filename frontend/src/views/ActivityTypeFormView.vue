<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useActivityTypesStore } from '@/stores/activityTypes'
import { ApiError } from '@/lib/api'
import FormInput from '@/components/FormInput.vue'
import AppButton from '@/components/AppButton.vue'
import AppAlert from '@/components/AppAlert.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'

const store = useActivityTypesStore()
const route = useRoute()
const router = useRouter()

const id = computed(() => (route.params.id as string | undefined) ?? '')
const isEdit = computed(() => id.value !== '')

const name = ref('')
const color = ref('#4f46e5')
const sort = ref('500')

const errors = ref<Record<string, string>>({})
const formError = ref('')
const loading = ref(false)
const ready = ref(false)

// --- Delete flow (edit only) ---
const confirmOpen = ref(false)
const deleting = ref(false)

onMounted(async () => {
  if (isEdit.value) {
    try {
      const item = await store.fetchOne(id.value)
      name.value = item.name
      color.value = item.color
      sort.value = String(item.sort)
    } catch {
      // Missing or not owned — send the user back to the list.
      router.replace({ name: 'activity-types' })
      return
    }
  }
  ready.value = true
})

function buildPayload() {
  const trimmed = sort.value.trim()
  const sortNum = trimmed === '' ? 500 : Math.trunc(Number(trimmed))
  return { name: name.value, color: color.value, sort: sortNum }
}

async function submit() {
  loading.value = true
  errors.value = {}
  formError.value = ''
  try {
    if (isEdit.value) {
      await store.update(id.value, buildPayload())
    } else {
      await store.create(buildPayload())
    }
    await router.push({ name: 'activity-types' })
  } catch (err) {
    if (err instanceof ApiError) {
      errors.value = err.errors
      formError.value = Object.keys(err.errors).length ? '' : err.message
    } else {
      formError.value = 'Something went wrong. Please try again.'
    }
  } finally {
    loading.value = false
  }
}

async function confirmDelete() {
  deleting.value = true
  formError.value = ''
  try {
    await store.remove(id.value)
    await router.push({ name: 'activity-types' })
  } catch (err) {
    confirmOpen.value = false
    formError.value = err instanceof ApiError ? err.message : 'Could not delete the activity type.'
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <section class="mx-auto max-w-2xl px-4 py-12 sm:px-6">
    <header class="mb-8">
      <h1 class="text-3xl font-semibold tracking-tight text-slate-900">
        {{ isEdit ? 'Edit activity type' : 'New activity type' }}
      </h1>
      <p class="mt-1.5 text-slate-500">Give it a name and pick a color.</p>
    </header>

    <div v-if="ready" class="rounded-2xl bg-white p-6 shadow-sm ring-1 ring-slate-200 sm:p-8">
      <AppAlert v-if="formError" variant="error" class="mb-5">{{ formError }}</AppAlert>

      <form class="space-y-5" @submit.prevent="submit">
        <FormInput
          id="name"
          v-model="name"
          label="Name"
          placeholder="e.g. Deep work"
          :error="errors.name"
          required
        />

        <div>
          <label for="color" class="mb-1.5 block text-sm font-medium text-slate-700">Color</label>
          <div class="flex items-center gap-3">
            <input
              id="color"
              v-model="color"
              type="color"
              class="h-11 w-14 cursor-pointer rounded-lg border-0 bg-white p-1 shadow-sm ring-1 ring-inset ring-slate-300 focus:ring-2 focus:ring-inset focus:ring-brand-500"
            />
            <span class="font-mono text-sm uppercase text-slate-600">{{ color }}</span>
          </div>
          <p v-if="errors.color" class="mt-1.5 text-sm text-red-600">{{ errors.color }}</p>
        </div>

        <FormInput
          id="sort"
          v-model="sort"
          label="Sort order"
          type="number"
          placeholder="500"
          :error="errors.sort"
        />
        <p class="-mt-3 text-xs text-slate-400">
          Lower numbers appear first in the list. Defaults to 500.
        </p>

        <div class="flex items-center justify-between pt-2">
          <AppButton
            v-if="isEdit"
            type="button"
            variant="danger"
            :disabled="loading"
            @click="confirmOpen = true"
          >
            Delete
          </AppButton>
          <span v-else />

          <div class="flex items-center gap-3">
            <AppButton :to="{ name: 'activity-types' }" variant="secondary">Cancel</AppButton>
            <AppButton type="submit" :loading="loading">
              {{ isEdit ? 'Save changes' : 'Create' }}
            </AppButton>
          </div>
        </div>
      </form>
    </div>

    <ConfirmDialog
      :open="confirmOpen"
      title="Delete activity type"
      :message="`Delete “${name}”? This can't be undone.`"
      confirm-label="Delete"
      :loading="deleting"
      @confirm="confirmDelete"
      @cancel="confirmOpen = false"
    />
  </section>
</template>
