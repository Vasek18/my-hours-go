<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ApiError } from '@/lib/api'
import AuthCard from '@/components/AuthCard.vue'
import FormInput from '@/components/FormInput.vue'
import AppButton from '@/components/AppButton.vue'
import AppAlert from '@/components/AppAlert.vue'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const email = ref('')
const password = ref('')
const errors = ref<Record<string, string>>({})
const formError = ref('')
const loading = ref(false)

async function submit() {
  loading.value = true
  errors.value = {}
  formError.value = ''
  try {
    await auth.login(email.value, password.value)
    const redirect = (route.query.redirect as string) || '/dashboard'
    router.push(redirect)
  } catch (err) {
    if (err instanceof ApiError) {
      errors.value = err.errors
      formError.value = Object.keys(err.errors).length ? '' : err.message
    } else {
      formError.value = 'Unable to sign in. Please try again.'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AuthCard title="Welcome back" subtitle="Sign in to your account">
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
      <FormInput
        id="password"
        v-model="password"
        label="Password"
        type="password"
        autocomplete="current-password"
        :error="errors.password"
        required
      />

      <div class="flex justify-end">
        <RouterLink
          :to="{ name: 'forgot-password' }"
          class="text-sm font-medium text-brand-600 hover:text-brand-700"
        >
          Forgot password?
        </RouterLink>
      </div>

      <AppButton type="submit" block :loading="loading">Sign in</AppButton>
    </form>

    <template #footer>
      Don't have an account?
      <RouterLink
        :to="{ name: 'register' }"
        class="font-medium text-brand-600 hover:text-brand-700"
      >
        Create one
      </RouterLink>
    </template>
  </AuthCard>
</template>
