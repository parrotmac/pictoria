import { ref } from 'vue'

const installPrompt = ref<any>(null)
const isInstallable = ref(false)

// Listen for the beforeinstallprompt event
if (typeof window !== 'undefined') {
  window.addEventListener('beforeinstallprompt', (e) => {
    // Prevent the default prompt
    e.preventDefault()
    // Store the event for later use
    installPrompt.value = e
    isInstallable.value = true
  })

  window.addEventListener('appinstalled', () => {
    // Clear the install prompt
    installPrompt.value = null
    isInstallable.value = false
  })
}

export function usePWA() {
  async function installApp() {
    if (!installPrompt.value) return false

    // Show the install prompt
    await installPrompt.value.prompt()
    
    // Wait for the user to respond
    const { outcome } = await installPrompt.value.userChoice
    
    if (outcome === 'accepted') {
      installPrompt.value = null
      isInstallable.value = false
      return true
    }
    
    return false
  }

  return {
    isInstallable,
    installApp
  }
}