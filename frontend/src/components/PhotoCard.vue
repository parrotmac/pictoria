<template>
  <div class="photo-card">
    <div class="photo-content">
      <img 
        :src="photoUrl" 
        :alt="photo.originalName"
        @load="imageLoaded = true"
        v-show="imageLoaded"
      />
      <div v-if="!imageLoaded" class="image-placeholder">
        <div class="spinner"></div>
      </div>
      <button 
        v-if="canDelete" 
        class="delete-button"
        @click="handleDelete"
        :disabled="isDeleting"
        aria-label="Delete photo"
      >
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M3 6h18M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2m3 0v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6h14zM10 11v6M14 11v6"/>
        </svg>
      </button>
    </div>
    <div class="photo-info" v-if="photo.uploaderName">
      <p>Uploaded by {{ photo.uploaderName }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { usePhotosStore } from '../stores/photos'
import { useAuthStore } from '../stores/auth'
import type { Photo } from '../types'

const props = defineProps<{
  photo: Photo
}>()

const emit = defineEmits<{
  deleted: []
}>()

const photosStore = usePhotosStore()
const authStore = useAuthStore()
const imageLoaded = ref(false)
const isDeleting = ref(false)

const photoUrl = computed(() => photosStore.getPhotoUrl(props.photo))
const canDelete = computed(() => {
  return authStore.user && authStore.user.id === props.photo.uploadedBy
})

async function handleDelete() {
  if (!confirm('Are you sure you want to delete this photo?')) {
    return
  }
  
  isDeleting.value = true
  try {
    await photosStore.deletePhoto(props.photo.id)
    emit('deleted')
  } catch (error) {
    console.error('Failed to delete photo:', error)
    alert('Failed to delete photo. Please try again.')
  } finally {
    isDeleting.value = false
  }
}
</script>

<style scoped>
.photo-card {
  background: var(--bg-secondary);
  border-radius: 10px;
  overflow: hidden;
  box-shadow: 0 2px 10px var(--shadow);
  transition: transform 0.3s ease;
}

.photo-card:hover {
  transform: translateY(-5px);
}

.photo-content {
  position: relative;
}

.photo-card img {
  width: 100%;
  height: auto;
  display: block;
}

.image-placeholder {
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-primary);
}

.delete-button {
  position: absolute;
  top: 8px;
  right: 8px;
  background: rgba(255, 255, 255, 0.9);
  border: none;
  border-radius: 8px;
  padding: 8px;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s ease, background-color 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.photo-card:hover .delete-button {
  opacity: 1;
}

.delete-button:hover {
  background: rgba(255, 255, 255, 1);
}

.delete-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.delete-button svg {
  color: #dc3545;
}

.photo-info {
  padding: 12px;
}

.photo-info p {
  color: var(--button-primary);
  font-size: 14px;
  margin: 0;
  font-weight: 500;
}
</style>