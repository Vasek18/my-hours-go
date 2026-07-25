<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'
import AppButton from '@/components/AppButton.vue'

const props = withDefaults(
  defineProps<{
    open: boolean
    title: string
    message?: string
    confirmLabel?: string
    cancelLabel?: string
    confirmVariant?: 'danger' | 'primary'
    loading?: boolean
  }>(),
  {
    message: '',
    confirmLabel: 'Confirm',
    cancelLabel: 'Cancel',
    confirmVariant: 'danger',
    loading: false,
  },
)

const emit = defineEmits<{ (e: 'confirm'): void; (e: 'cancel'): void }>()

const confirmButton = ref<InstanceType<typeof AppButton> | null>(null)

function cancel() {
  if (props.loading) return
  emit('cancel')
}

function onKeydown(event: KeyboardEvent) {
  if (props.open && event.key === 'Escape') cancel()
}

// Move focus onto the confirm action when the dialog opens.
watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      nextTick(() => (confirmButton.value?.$el as HTMLElement | undefined)?.focus())
    }
  },
)

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
        @mousedown.self="cancel"
      >
        <div
          role="dialog"
          aria-modal="true"
          aria-labelledby="confirm-dialog-title"
          class="w-full max-w-md rounded-2xl bg-white p-6 shadow-xl ring-1 ring-slate-200 sm:p-8"
        >
          <h2 id="confirm-dialog-title" class="text-lg font-semibold text-slate-900">
            {{ title }}
          </h2>
          <p v-if="message" class="mt-2 text-sm text-slate-600">{{ message }}</p>

          <div class="mt-6 flex justify-end gap-3">
            <AppButton variant="secondary" :disabled="loading" @click="cancel">
              {{ cancelLabel }}
            </AppButton>
            <AppButton
              ref="confirmButton"
              :variant="confirmVariant"
              :loading="loading"
              @click="emit('confirm')"
            >
              {{ confirmLabel }}
            </AppButton>
          </div>
        </div>
      </div>
    </transition>
  </Teleport>
</template>
