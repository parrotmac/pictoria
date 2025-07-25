<template>
    <transition name="slide">
        <div v-if="networkStore.showBanner && !hideConnectionHint"
            class="network-routing-hint">
            <WifiOff :size="20" />
            <span>
                Join the Wi-Fi network for best experience!
            </span>
            <button @click="showNetworkConnectionDialog = true" class="btn btn-primary">
                Connect to Wi-Fi
            </button>
            <button @click="hideConnectionHint = true" class="btn btn-secondary">
                Dismiss
            </button>
        </div>
    </transition>
    <NetworkRoutingModal
        v-if="showNetworkConnectionDialog && !!networkStore.showBanner && !hideConnectionHint"
        :hint="networkStore.directNetworkingConnectivityDetails" @close="showNetworkConnectionDialog = false" />
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { WifiOff } from 'lucide-vue-next'
import NetworkRoutingModal from './NetworkRoutingModal.vue';
import { useNetworkingStore } from '../stores/network';

const networkStore = useNetworkingStore()
const showNetworkConnectionDialog = ref(false)
const hideConnectionHint = ref(false)

onMounted(async () => {
    await networkStore.fetchDirectionConnectionDetails()
})

onUnmounted(() => {
    // Ensure polling stops when component is destroyed
    networkStore.stopPolling()
})

watch(() => networkStore.directApiAvailable, (isAvailable) => {
    if (!isAvailable) {
        hideConnectionHint.value = false
        console.info('Direct API is not available - showing connection hint')
    } else {
        hideConnectionHint.value = true
        console.info('Direct API is available - hiding connection hint')
    }
})

// Watch for hint disappearing and auto-close dialog/reset banner state
watch(() => networkStore.directNetworkingConnectivityDetails, (connectionDetails) => {
    if (connectionDetails) {
        showNetworkConnectionDialog.value = true
        console.info('Network connection hint available - showing dialog')
    } else {
        showNetworkConnectionDialog.value = false
        console.info('Network connection hint disappeared - hiding dialog')
    }
})

</script>
<style scoped>
.network-routing-hint {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    background-color: var(--button-danger);
    color: white;
    padding: 6px;
    text-align: center;
    z-index: 9999;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    font-weight: 500;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
}

.slide-enter-active,
.slide-leave-active {
    transition: transform 0.3s ease;
}

.slide-enter-from {
    transform: translateY(-100%);
}

.slide-leave-to {
    transform: translateY(-100%);
}
</style>
