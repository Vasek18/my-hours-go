<script setup lang="ts">
withDefaults(
  defineProps<{
    id: string
    label: string
    modelValue: string
    type?: string
    autocomplete?: string
    placeholder?: string
    error?: string
    required?: boolean
  }>(),
  {
    type: 'text',
    autocomplete: 'off',
    placeholder: '',
    error: '',
    required: false,
  },
)

defineEmits<{ (e: 'update:modelValue', value: string): void }>()
</script>

<template>
  <div>
    <label :for="id" class="mb-1.5 block text-sm font-medium text-slate-700">
      {{ label }}
    </label>
    <input
      :id="id"
      :name="id"
      :type="type"
      :value="modelValue"
      :autocomplete="autocomplete"
      :placeholder="placeholder"
      :required="required"
      :aria-invalid="!!error"
      class="block w-full rounded-lg border-0 bg-white px-3.5 py-2.5 text-slate-900 shadow-sm ring-1 ring-inset ring-slate-300 placeholder:text-slate-400 focus:ring-2 focus:ring-inset focus:ring-brand-500"
      :class="error ? 'ring-red-400 focus:ring-red-500' : ''"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
    <p v-if="error" class="mt-1.5 text-sm text-red-600">{{ error }}</p>
  </div>
</template>
