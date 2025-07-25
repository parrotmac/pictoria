import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../api'
import type { NetworkRoutingHint } from '../types'


export const useNetworkingStore = defineStore('networking', () => {
    const directNetworkingConnectivityDetails = ref<NetworkRoutingHint | null>(null)
    const loading = ref(false)
    const error = ref<string | null>(null)
    const isPolling = ref(false)
    let pollingInterval: number | null = null
    const debugErrors = ref<string[]>([]);
    const directApiAvailable = ref(false)

    const showBanner = computed(() => {
        return directNetworkingConnectivityDetails && !directApiAvailable
    })

    async function fetchDirectionConnectionDetails(): Promise<NetworkRoutingHint | null> {
        try {
            loading.value = true
            error.value = null

            const response = await api.get<NetworkRoutingHint>('/api/direct-networking', {
                routingMethod: 'global',
            });
            directNetworkingConnectivityDetails.value = response.data || null

            if (directNetworkingConnectivityDetails.value) {
                startPolling()
            } else {
                stopPolling()
            }

            return directNetworkingConnectivityDetails.value
        } catch (err: any) {
            error.value = `Failed to fetch network routing hint: ${err.message}`
            debugErrors.value = [...debugErrors.value, `Fetch Error: ${err.message}`].slice(-10)
            return null
        } finally {
            loading.value = false
        }
    }

    async function pollDirectConnectivity() {
        // Same as fetchNetworkingHints but without setting loading state
        try {
            const directApiAvailableResponse = await api.checkDirectApiAvailability();
            directApiAvailable.value = directApiAvailableResponse;
            if (!directApiAvailableResponse) {
                debugErrors.value = [...debugErrors.value, 'Direct API is not available'].slice(-10)
                return
            }
        } catch (err: any) {
            debugErrors.value = [...debugErrors.value, `Direct Connection Poll Error: ${err.message}`].slice(-10)
        }
    }

    function startPolling(intervalMs: number = 5000) {
        if (isPolling.value) {
            return
        }
        isPolling.value = true
        pollingInterval = window.setInterval(pollDirectConnectivity, intervalMs)
    }

    function stopPolling() {
        if (pollingInterval) {
            clearInterval(pollingInterval)
            pollingInterval = null
        }
        isPolling.value = false
    }

    return {
        fetchDirectionConnectionDetails,
        directNetworkingConnectivityDetails,
        directApiAvailable,
        loading,
        error,
        debugErrors,
        isPolling,
        startPolling,
        stopPolling,
        showBanner,
    }
})
