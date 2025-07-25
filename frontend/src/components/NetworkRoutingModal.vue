<template>
    <div class="modal-overlay" @click.self="emit('close')">
        <div class="modal-content" style="max-width: 400px;">
            <h2>Connect to Wi-Fi</h2>
            <p style="margin: 20px 0; color: var(--text-secondary);">
                Please connect to the Wi-Fi network for optimal performance.
            </p>
            <div v-if="hint?.wifi" class="wifi-details">
                <div class="wifi-field">
                    <p><strong>Name</strong> {{ hint.wifi.ssid }}</p>
                </div>
                <div v-if="hint.wifi.password" class="wifi-field">
                    <p><strong>Password</strong> <span class="monospaced">{{ hint.wifi.password }}</span></p>
                    <button 
                        class="btn btn-copy" 
                        @click="copyToClipboard(hint.wifi.password, 'Wi-Fi password')"
                        title="Copy Wi-Fi password"
                    >
                        📋
                    </button>
                </div>
            </div>
            <!-- <details v-if="networkStore.debugErrors">
                <summary>Debug Information</summary>
                <pre>{{ networkStore.debugErrors }}</pre>
                <pre>{{ networkStore.directNetworkingConnectivityDetails }}</pre>
            </details> -->
            <div class="modal-actions">
                <button class="btn btn-primary" @click="emit('close')">Close</button>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, watch } from 'vue'
import type { NetworkRoutingHint } from '../types';
import { useNetworkingStore } from '../stores/network'

defineProps<{
    hint: NetworkRoutingHint | null
}>()

const emit = defineEmits<{
    close: []
}>()

const networkStore = useNetworkingStore()

const copyToClipboard = async (text: string, fieldName: string) => {
    try {
        await navigator.clipboard.writeText(text);
        // You could add a toast notification here if you have one
        console.log(`${fieldName} copied to clipboard`);
    } catch (err) {
        console.error('Failed to copy to clipboard:', err);
        // Fallback for older browsers
        const textArea = document.createElement('textarea');
        textArea.value = text;
        document.body.appendChild(textArea);
        textArea.select();
        document.execCommand('copy');
        document.body.removeChild(textArea);
        console.log(`${fieldName} copied to clipboard (fallback)`);
    }
}

// Start polling when component mounts
onMounted(() => {
    networkStore.startPolling(3000) // Poll every 3 seconds
})

// Stop polling when component unmounts
onUnmounted(() => {
    networkStore.stopPolling()
})

// Watch for hint changes and auto-close if hint disappears
watch(() => networkStore.directNetworkingConnectivityDetails, (connectionDetails) => {
    if (!connectionDetails) {
        emit('close')
    }
})
</script>

<style scoped>
.wifi-field {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 10px 0;
}

.wifi-field p {
    margin: 0;
    flex: 1;
}

.btn-copy {
    background: var(--bg-secondary, #f5f5f5);
    border: 1px solid var(--border-color, #ddd);
    border-radius: 4px;
    padding: 4px 8px;
    cursor: pointer;
    font-size: 14px;
    transition: background-color 0.2s;
    min-width: 32px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.btn-copy:hover {
    background: var(--bg-hover, #e9e9e9);
}

.btn-copy:active {
    background: var(--bg-active, #d4d4d4);
}

.wifi-details {
    background: var(--bg-secondary, #f9f9f9);
    border: 1px solid var(--border-color, #e1e1e1);
    border-radius: 8px;
    padding: 16px;
    margin: 16px 0;
}

.monospaced {
    font-family: monospace, monospace;
}
</style>