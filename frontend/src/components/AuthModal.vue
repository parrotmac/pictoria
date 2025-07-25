<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-content" style="max-width: 400px;">
      <h2>Welcome to Pictoria!</h2>
      <p style="margin: 20px 0; color: var(--text-secondary);">
        Please enter your name to continue
      </p>

      <input
        type="text"
        v-model="userName"
        placeholder="Your name"
        @keyup.enter="handleSubmit"
        class="auth-input"
      />

      <div v-if="error" class="error-message">
        {{ error }}
      </div>

      <button
        class="btn"
        @click="handleSubmit"
        :disabled="!userName.trim() || loading"
        style="margin-top: 20px; width: 100%;"
      >
        {{ loading ? 'Creating...' : 'Continue' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { nameStore } from '../stores/auth';

const emit = defineEmits<{
  close: []
}>()

const userName = ref('')
const error = ref('')
const loading = ref(false)

async function handleSubmit() {
  if (!userName.value.trim()) return

  console.log(`new name`, userName.value.trim())
  nameStore.value = userName.value.trim()
  console.log(`new name`, nameStore.value)
  emit('close')
}
</script>

<style scoped>
.auth-input {
  width: 100%;
  padding: 10px;
  border: 1px solid var(--border-color);
  border-radius: 5px;
  font-size: 16px;
  background: var(--bg-primary);
  color: var(--text-primary);
}

.auth-input:focus {
  outline: none;
  border-color: var(--border-hover);
}

.error-message {
  color: var(--button-danger);
  margin-top: 10px;
}
</style>
