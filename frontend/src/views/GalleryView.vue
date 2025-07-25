<template>
  <div>
    <div class="sticky-nav">
      <div class="sticky-nav-content">
        <router-link to="/" class="btn">
          Return to Upload
        </router-link>
      </div>
    </div>
    
    <div class="container">
      <div v-if="photosStore.loading" class="loading">
        <div class="spinner"></div>
        <p>Loading photos...</p>
      </div>

      <div v-else-if="photosStore.error" class="error-message">
        {{ photosStore.error }}
      </div>

      <div v-else class="photo-grid">
        <PhotoCard
          v-for="photo in photosStore.photos"
          :key="photo.id"
          :photo="photo"
          :id="`photo-${photo.id}`"
        />
      </div>

      <div v-if="!photosStore.loading && photosStore.photos.length === 0" class="empty-state">
        <p>
          No photos yet. <router-link to="/" class="text-link">Start uploading</router-link> or capture your first photo!
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { usePhotosStore } from '../stores/photos'
import PhotoCard from '../components/PhotoCard.vue'

const route = useRoute()
const photosStore = usePhotosStore()
let endpointUpdateInterval: number | null = null

onMounted(async () => {
  await photosStore.fetchPhotos()
  
  // Check for photo ID in hash
  const hash = route.hash
  if (hash && hash.startsWith('#photo-')) {
    const photoId = hash.replace('#photo-', '')
    await nextTick()
    
    // Scroll to the specific photo
    setTimeout(() => {
      const element = document.getElementById(`photo-${photoId}`)
      if (element) {
        element.scrollIntoView({ behavior: 'smooth', block: 'center' })
      }
    }, 100)
  }

  // Update active endpoint periodically to ensure photo URLs work correctly
  endpointUpdateInterval = window.setInterval(() => {
    photosStore.updateActiveEndpoint()
  }, 5000) // Check every 5 seconds
})

onUnmounted(() => {
  if (endpointUpdateInterval) {
    clearInterval(endpointUpdateInterval)
  }
})
</script>

<style scoped>
.sticky-nav {
  position: sticky;
  top: 0;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  padding: 15px 0;
  z-index: 100;
  box-shadow: 0 2px 4px var(--shadow);
}

.sticky-nav-content {
  display: flex;
  align-items: center;
  justify-content: center;
}

.photo-grid {
  display: flex;
  flex-direction: column;
  gap: 20px;
  margin-top: 40px;
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-secondary);
}

.empty-state p {
  font-size: 1.1rem;
  line-height: 1.6;
}

.text-link {
  color: var(--button-primary);
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s ease;
}

.text-link:hover {
  color: var(--button-primary-hover);
  text-decoration: underline;
}

.error-message {
  text-align: center;
  color: var(--button-danger);
  padding: 40px 20px;
}

@media (max-width: 600px) {
  .photo-grid {
    max-width: 100%;
    padding: 0 10px;
  }
}
</style>