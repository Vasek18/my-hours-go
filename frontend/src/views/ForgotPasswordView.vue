<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ApiError } from '@/lib/api'
import AuthCard from '@/components/AuthCard.vue'
import FormInput from '@/components/FormInput.vue'
import AppButton from '@/components/AppButton.vue'
import AppAlert from '@/components/AppAlert.vue'

const auth = useAuthStore()

const email = ref('')
const errors = ref<Record<string, string>>({})
const formError = ref('')
const successMessage = ref('')
const loading = ref(false)

async function submit() {
  loading.value = true
  errors.value = {}
  formError.value = ''
  successMessage.value = ''
  try {
    successMessage.value = await auth.forgotPassword(email.value)
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
</script>

<template>
  <AuthCard title="Reset your password" subtitle="We'll send you a link to choose a new one">
    <AppAlert v-if="successMessage" variant="success" class="mb-5">
      {{ successMessage }}
      <span class="mt-1 block text-emerald-600/80">
        In local development no real email is sent — open the reset link printed in the backend
        logs.
      </span>
    </AppAlert>
    <AppAlert v-if="formError" variant="error" class="mb-5">{{ formError }}</AppAlert>

    <form class="space-y-5" @submit.prevent="submit">
      <FormInput
        id="email"
        v-model="email"
        label="Email"
        type="email"
        autocomplete="email"
        :error="errors.email"
        required
      />
      <AppButton type="submit" block :loading="loading">Send reset link</AppButton>
    </form>

    <template #footer>
      <RouterLink :to="{ name: 'login' }" class="font-medium text-brand-600 hover:text-brand-700">
        Back to sign in
      </RouterLink>
    </template>
  </AuthCard>
</template>
