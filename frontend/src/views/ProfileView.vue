<script setup lang="ts">
import { ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import { ApiError } from '@/lib/api'
import FormInput from '@/components/FormInput.vue'
import AppButton from '@/components/AppButton.vue'
import AppAlert from '@/components/AppAlert.vue'

const auth = useAuthStore()
const { user } = storeToRefs(auth)

// --- Profile (name) form ---
const name = ref(user.value?.name ?? '')
const profileErrors = ref<Record<string, string>>({})
const profileError = ref('')
const profileSuccess = ref('')
const profileLoading = ref(false)

// Keep the field in sync once the session has hydrated.
watch(user, (u) => {
  if (u && !name.value) name.value = u.name
})

async function saveProfile() {
  profileLoading.value = true
  profileErrors.value = {}
  profileError.value = ''
  profileSuccess.value = ''
  try {
    await auth.updateProfile(name.value)
    profileSuccess.value = 'Your profile has been updated.'
  } catch (err) {
    if (err instanceof ApiError) {
      profileErrors.value = err.errors
      profileError.value = Object.keys(err.errors).length ? '' : err.message
    } else {
      profileError.value = 'Something went wrong. Please try again.'
    }
  } finally {
    profileLoading.value = false
  }
}

// --- Email form ---
const newEmail = ref('')
const emailCurrentPassword = ref('')
const emailErrors = ref<Record<string, string>>({})
const emailError = ref('')
const emailSuccess = ref('')
const emailLoading = ref(false)

async function saveEmail() {
  emailLoading.value = true
  emailErrors.value = {}
  emailError.value = ''
  emailSuccess.value = ''
  try {
    emailSuccess.value = await auth.changeEmail({
      new_email: newEmail.value,
      current_password: emailCurrentPassword.value,
    })
    newEmail.value = ''
    emailCurrentPassword.value = ''
  } catch (err) {
    if (err instanceof ApiError) {
      emailErrors.value = err.errors
      emailError.value = Object.keys(err.errors).length ? '' : err.message
    } else {
      emailError.value = 'Something went wrong. Please try again.'
    }
  } finally {
    emailLoading.value = false
  }
}

// --- Password form ---
const currentPassword = ref('')
const password = ref('')
const passwordConfirmation = ref('')
const passwordErrors = ref<Record<string, string>>({})
const passwordError = ref('')
const passwordSuccess = ref('')
const passwordLoading = ref(false)

async function savePassword() {
  passwordLoading.value = true
  passwordErrors.value = {}
  passwordError.value = ''
  passwordSuccess.value = ''
  try {
    passwordSuccess.value = await auth.changePassword({
      current_password: currentPassword.value,
      password: password.value,
      password_confirmation: passwordConfirmation.value,
    })
    currentPassword.value = ''
    password.value = ''
    passwordConfirmation.value = ''
  } catch (err) {
    if (err instanceof ApiError) {
      passwordErrors.value = err.errors
      passwordError.value = Object.keys(err.errors).length ? '' : err.message
    } else {
      passwordError.value = 'Something went wrong. Please try again.'
    }
  } finally {
    passwordLoading.value = false
  }
}
</script>

<template>
  <section class="mx-auto max-w-2xl px-4 py-12 sm:px-6">
    <header class="mb-8">
      <h1 class="text-3xl font-semibold tracking-tight text-slate-900">Profile</h1>
      <p class="mt-1.5 text-slate-500">Manage your account details and password.</p>
    </header>

    <div class="space-y-8">
      <!-- Name -->
      <div class="rounded-2xl bg-white p-6 shadow-sm ring-1 ring-slate-200 sm:p-8">
        <h2 class="text-lg font-semibold text-slate-900">Your details</h2>
        <p class="mt-1 text-sm text-slate-500">Update the name shown on your account.</p>

        <AppAlert v-if="profileSuccess" variant="success" class="mt-5">{{
          profileSuccess
        }}</AppAlert>
        <AppAlert v-if="profileError" variant="error" class="mt-5">{{ profileError }}</AppAlert>

        <form class="mt-6 space-y-5" @submit.prevent="saveProfile">
          <FormInput
            id="name"
            v-model="name"
            label="Name"
            autocomplete="name"
            :error="profileErrors.name"
            required
          />
          <div class="flex justify-end">
            <AppButton type="submit" :loading="profileLoading">Save changes</AppButton>
          </div>
        </form>
      </div>

      <!-- Email -->
      <div class="rounded-2xl bg-white p-6 shadow-sm ring-1 ring-slate-200 sm:p-8">
        <h2 class="text-lg font-semibold text-slate-900">Email address</h2>
        <p class="mt-1 text-sm text-slate-500">
          Your current email is <span class="font-medium text-slate-700">{{ user?.email }}</span
          >. Changing it requires confirming the new address.
        </p>

        <AppAlert v-if="emailSuccess" variant="success" class="mt-5">
          {{ emailSuccess }}
          <span class="mt-1 block text-emerald-600/80">
            In local development no real email is sent — open the confirmation link printed in the
            backend logs.
          </span>
        </AppAlert>
        <AppAlert v-if="emailError" variant="error" class="mt-5">{{ emailError }}</AppAlert>

        <form class="mt-6 space-y-5" @submit.prevent="saveEmail">
          <FormInput
            id="new_email"
            v-model="newEmail"
            label="New email"
            type="email"
            autocomplete="email"
            :error="emailErrors.new_email"
            required
          />
          <FormInput
            id="email_current_password"
            v-model="emailCurrentPassword"
            label="Current password"
            type="password"
            autocomplete="current-password"
            :error="emailErrors.current_password"
            required
          />
          <div class="flex justify-end">
            <AppButton type="submit" :loading="emailLoading">Send confirmation link</AppButton>
          </div>
        </form>
      </div>

      <!-- Password -->
      <div class="rounded-2xl bg-white p-6 shadow-sm ring-1 ring-slate-200 sm:p-8">
        <h2 class="text-lg font-semibold text-slate-900">Change password</h2>
        <p class="mt-1 text-sm text-slate-500">Enter your current password to choose a new one.</p>

        <AppAlert v-if="passwordSuccess" variant="success" class="mt-5">{{
          passwordSuccess
        }}</AppAlert>
        <AppAlert v-if="passwordError" variant="error" class="mt-5">{{ passwordError }}</AppAlert>

        <form class="mt-6 space-y-5" @submit.prevent="savePassword">
          <FormInput
            id="current_password"
            v-model="currentPassword"
            label="Current password"
            type="password"
            autocomplete="current-password"
            :error="passwordErrors.current_password"
            required
          />
          <FormInput
            id="new_password"
            v-model="password"
            label="New password"
            type="password"
            autocomplete="new-password"
            placeholder="At least 8 characters"
            :error="passwordErrors.password"
            required
          />
          <FormInput
            id="new_password_confirmation"
            v-model="passwordConfirmation"
            label="Confirm new password"
            type="password"
            autocomplete="new-password"
            :error="passwordErrors.password_confirmation"
            required
          />
          <div class="flex justify-end">
            <AppButton type="submit" :loading="passwordLoading">Update password</AppButton>
          </div>
        </form>
      </div>
    </div>
  </section>
</template>
