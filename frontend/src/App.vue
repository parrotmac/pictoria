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
import { ref, provide, watch } from 'vue'
import { nameStore } from './stores/auth'
import AuthModal from './components/AuthModal.vue'
import OfflineIndicator from './components/OfflineIndicator.vue'
import NetworkRoutingHint from './components/NetworkRoutingHint.vue'

const showAuthModal = ref(false)

// Provide auth modal control to child components
provide('showAuthModal', (show: boolean) => {
  showAuthModal.value = show
})

watch(nameStore, (newVal: string) => {
  if (!newVal) {
    showAuthModal.value = true
  }
}, { immediate: true })
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
