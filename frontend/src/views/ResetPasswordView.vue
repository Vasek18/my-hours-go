<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ApiError } from '@/lib/api'
import AuthCard from '@/components/AuthCard.vue'
import FormInput from '@/components/FormInput.vue'
import AppButton from '@/components/AppButton.vue'
import AppAlert from '@/components/AppAlert.vue'

const auth = useAuthStore()
const route = useRoute()

const token = (route.query.token as string) || ''
const password = ref('')
const passwordConfirmation = ref('')
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
    successMessage.value = await auth.resetPassword({
      token,
      password: password.value,
      password_confirmation: passwordConfirmation.value,
    })
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
  <AuthCard title="Choose a new password" subtitle="Enter a new password for your account">
    <AppAlert v-if="!token" variant="error" class="mb-5">
      This reset link is missing its token. Please request a new one.
    </AppAlert>

    <AppAlert v-if="successMessage" variant="success" class="mb-5">{{ successMessage }}</AppAlert>
    <AppAlert v-if="formError" variant="error" class="mb-5">{{ formError }}</AppAlert>
    <AppAlert v-if="errors.token" variant="error" class="mb-5">{{ errors.token }}</AppAlert>

    <form v-if="!successMessage" class="space-y-5" @submit.prevent="submit">
      <FormInput
        id="password"
        v-model="password"
        label="New password"
        type="password"
        autocomplete="new-password"
        placeholder="At least 8 characters"
        :error="errors.password"
        required
      />
      <FormInput
        id="password_confirmation"
        v-model="passwordConfirmation"
        label="Confirm new password"
        type="password"
        autocomplete="new-password"
        :error="errors.password_confirmation"
        required
      />
      <AppButton type="submit" block :loading="loading" :disabled="!token">
        Reset password
      </AppButton>
    </form>

    <template #footer>
      <RouterLink :to="{ name: 'login' }" class="font-medium text-brand-600 hover:text-brand-700">
        Back to sign in
      </RouterLink>
    </template>
  </AuthCard>
</template>
