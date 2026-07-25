<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ApiError } from '@/lib/api'
import AuthCard from '@/components/AuthCard.vue'
import FormInput from '@/components/FormInput.vue'
import AppButton from '@/components/AppButton.vue'
import AppAlert from '@/components/AppAlert.vue'

const auth = useAuthStore()
const router = useRouter()

const name = ref('')
const email = ref('')
const password = ref('')
const passwordConfirmation = ref('')
const errors = ref<Record<string, string>>({})
const formError = ref('')
const loading = ref(false)

async function submit() {
  loading.value = true
  errors.value = {}
  formError.value = ''
  try {
    await auth.register({
      name: name.value,
      email: email.value,
      password: password.value,
      password_confirmation: passwordConfirmation.value,
    })
    router.push({ name: 'dashboard' })
  } catch (err) {
    if (err instanceof ApiError) {
      errors.value = err.errors
      formError.value = Object.keys(err.errors).length ? '' : err.message
    } else {
      formError.value = 'Unable to create your account. Please try again.'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AuthCard title="Create your account" subtitle="Start building in seconds">
    <AppAlert v-if="formError" variant="error" class="mb-5">{{ formError }}</AppAlert>

    <form class="space-y-5" @submit.prevent="submit">
      <FormInput
        id="name"
        v-model="name"
        label="Name"
        autocomplete="name"
        :error="errors.name"
        required
      />
      <FormInput
        id="email"
        v-model="email"
        label="Email"
        type="email"
        autocomplete="email"
        :error="errors.email"
        required
      />
      <FormInput
        id="password"
        v-model="password"
        label="Password"
        type="password"
        autocomplete="new-password"
        placeholder="At least 8 characters"
        :error="errors.password"
        required
      />
      <FormInput
        id="password_confirmation"
        v-model="passwordConfirmation"
        label="Confirm password"
        type="password"
        autocomplete="new-password"
        :error="errors.password_confirmation"
        required
      />

      <AppButton type="submit" block :loading="loading">Create account</AppButton>
    </form>

    <template #footer>
      Already have an account?
      <RouterLink :to="{ name: 'login' }" class="font-medium text-brand-600 hover:text-brand-700">
        Sign in
      </RouterLink>
    </template>
  </AuthCard>
</template>
