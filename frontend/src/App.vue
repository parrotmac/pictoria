<template>
  <div id="app">
    <!-- Offline indicator -->
    <OfflineIndicator />
    
    <router-view v-slot="{ Component }">
      <transition name="fade" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>

    <!-- Global auth modal -->
    <AuthModal v-if="showAuthModal" @close="showAuthModal = false" />
    <NetworkRoutingHint />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, provide } from 'vue'
import { useAuthStore } from './stores/auth'
import AuthModal from './components/AuthModal.vue'
import OfflineIndicator from './components/OfflineIndicator.vue'
import NetworkRoutingHint from './components/NetworkRoutingHint.vue'

const authStore = useAuthStore()
const showAuthModal = ref(false)

// Provide auth modal control to child components
provide('showAuthModal', (show: boolean) => {
  showAuthModal.value = show
})

onMounted(async () => {
  // Check authentication on app load
  await authStore.checkAuth()
  
  // Show auth modal if not authenticated
  if (!authStore.isAuthenticated) {
    showAuthModal.value = true
  }
})
</script>

<style>
@import './assets/styles/main.css';

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
