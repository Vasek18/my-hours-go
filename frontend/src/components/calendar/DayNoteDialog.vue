<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'
import AppButton from '@/components/AppButton.vue'

const props = withDefaults(
  defineProps<{
    open: boolean
    date: Date | null
    content: string
    saving?: boolean
  }>(),
  { saving: false },
)

const emit = defineEmits<{ (e: 'save', value: string): void; (e: 'close'): void }>()

const draft = ref('')
const textarea = ref<HTMLTextAreaElement | null>(null)
const fmt = new Intl.DateTimeFormat(undefined, {
  weekday: 'long',
  month: 'long',
  day: 'numeric',
})
const label = () => (props.date ? fmt.format(props.date) : '')

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      draft.value = props.content
      nextTick(() => textarea.value?.focus())
    }
  },
)

function close() {
  if (!props.saving) emit('close')
}
function save() {
  emit('save', draft.value)
}
function onKeydown(event: KeyboardEvent) {
  if (!props.open) return
  if (event.key === 'Escape') close()
  if (event.key === 'Enter' && (event.metaKey || event.ctrlKey)) save()
}

onMounted(() => document.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => document.removeEventListener('keydown', onKeydown))
</script>

<template>
  <Teleport to="body">
    <transition
      enter-active-class="transition duration-100 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-75 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="open"
        class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4"
        @mousedown.self="close"
      >
        <div
          role="dialog"
          aria-modal="true"
          class="flex w-full max-w-lg flex-col rounded-2xl bg-white p-6 shadow-xl ring-1 ring-slate-200 sm:p-7"
        >
          <div class="mb-3 flex items-center justify-between gap-3">
            <h2 class="text-lg font-semibold text-slate-900">Notes · {{ label() }}</h2>
            <button
              type="button"
              class="rounded-md p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
              aria-label="Close"
              @click="close"
            >
              ✕
            </button>
          </div>

          <textarea
            ref="textarea"
            v-model="draft"
            rows="8"
            placeholder="How was your day?"
            class="w-full resize-y rounded-lg border-0 bg-white px-3.5 py-2.5 text-sm text-slate-900 ring-1 ring-inset ring-slate-300 placeholder:text-slate-400 focus:ring-2 focus:ring-inset focus:ring-brand-500"
          />

          <div class="mt-4 flex items-center justify-end gap-3">
            <AppButton variant="secondary" :disabled="saving" @click="close">Cancel</AppButton>
            <AppButton :loading="saving" @click="save">Save</AppButton>
          </div>
        </div>
      </div>
    </transition>
  </Teleport>
</template>
