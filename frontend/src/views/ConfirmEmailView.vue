<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import { ApiError } from '@/lib/api'
import AuthCard from '@/components/AuthCard.vue'
import AppAlert from '@/components/AppAlert.vue'
import AppButton from '@/components/AppButton.vue'

const auth = useAuthStore()
const { isAuthenticated } = storeToRefs(auth)
const route = useRoute()

const status = ref<'loading' | 'success' | 'error'>('loading')
const message = ref('')

onMounted(async () => {
  const token = (route.query.token as string) || ''
  if (!token) {
    status.value = 'error'
    message.value = 'This confirmation link is missing its token.'
    return
  }
  try {
    message.value = await auth.confirmEmailChange(token)
    status.value = 'success'
  } catch (err) {
    status.value = 'error'
    if (err instanceof ApiError) {
      message.value = err.errors.token || err.message
    } else {
      message.value = 'Something went wrong confirming your email.'
    }
  }
})
</script>

<template>
  <AuthCard title="Confirm email change">
    <p v-if="status === 'loading'" class="text-sm text-slate-500">Confirming your new email…</p>

    <AppAlert v-else-if="status === 'success'" variant="success">{{ message }}</AppAlert>
    <AppAlert v-else variant="error">{{ message }}</AppAlert>

    <div v-if="status !== 'loading'" class="mt-6">
      <AppButton v-if="isAuthenticated" :to="{ name: 'profile' }" block>Back to profile</AppButton>
      <AppButton v-else :to="{ name: 'login' }" block>Go to sign in</AppButton>
    </div>

    <template #footer>
      <RouterLink to="/" class="font-medium text-brand-600 hover:text-brand-700">Home</RouterLink>
    </template>
  </AuthCard>
</template>
