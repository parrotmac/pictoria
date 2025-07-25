<template>
  <div class="container">
    <div class="header">
      <h1>Pictoria 📸</h1>
      <small>Upload or capture your photos</small>
      <p v-if="authStore.user" class="welcome-text">
        Welcome, {{ authStore.user.name }}!
      </p>
    </div>

    <div class="action-buttons">
      <div class="action-button" @click="showUploadModal = true">
        <Upload :size="50" />
        <h2>Upload Photo</h2>
        <p>Choose from your device</p>
      </div>

      <div class="action-button" @click="showCameraModal = true">
        <Camera :size="50" />
        <h2>Capture Photo</h2>
        <p>Use your camera</p>
      </div>
    </div>

    <div class="gallery-button">
      <router-link to="/gallery" class="view-gallery-button">
        <h3>View Gallery</h3>
      </router-link>
    </div>

    <!-- Upload Modal -->
    <PhotoUpload 
      v-if="showUploadModal" 
      @close="showUploadModal = false"
      @uploaded="handlePhotoUploaded"
    />

    <!-- Camera Modal -->
    <PhotoCapture 
      v-if="showCameraModal" 
      @close="showCameraModal = false"
      @captured="handlePhotoUploaded"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { Upload, Camera } from 'lucide-vue-next'
import PhotoUpload from '../components/PhotoUpload.vue'
import PhotoCapture from '../components/PhotoCapture.vue'

const router = useRouter()
const authStore = useAuthStore()

const showUploadModal = ref(false)
const showCameraModal = ref(false)

function handlePhotoUploaded(photoId: string) {
  // Navigate to gallery with the photo ID
  router.push({ 
    path: '/gallery', 
    hash: `#photo-${photoId}` 
  })
}
</script>

<style scoped>
.header {
  text-align: center;
  padding: 20px 0 15px 0;
}

.header h1 {
  font-size: 2rem;
  color: var(--header-color);
  margin-bottom: 5px;
}

.welcome-text {
  margin-top: 10px;
  color: var(--text-secondary);
  font-size: 14px;
}

.action-buttons {
  display: flex;
  gap: 20px;
  justify-content: center;
  margin: 25px 0;
  flex-wrap: wrap;
}

.action-button {
  background: var(--bg-secondary);
  border: 2px solid var(--border-color);
  border-radius: 20px;
  padding: 40px 30px;
  cursor: pointer;
  transition: all 0.3s ease;
  text-align: center;
  min-width: 220px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.action-button:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 30px var(--shadow);
  border-color: var(--border-hover);
}

.action-button svg {
  color: var(--button-primary);
  margin-bottom: 15px;
}

.action-button h2 {
  font-size: 1.3rem;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.action-button p {
  color: var(--text-secondary);
}

.gallery-button {
  display: flex;
  justify-content: center;
  margin: 25px 0;
}

.view-gallery-button {
  background: var(--bg-secondary);
  border: 2px solid var(--border-color);
  border-radius: 20px;
  padding: 20px 30px;
  cursor: pointer;
  transition: all 0.3s ease;
  text-align: center;
  min-width: 200px;
  text-decoration: none;
  display: inline-block;
  color: var(--text-primary);
}

.view-gallery-button:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 30px var(--shadow);
  border-color: var(--button-success);
}

.view-gallery-button h3 {
  font-size: 1.2rem;
  color: var(--button-success);
  margin: 0;
  font-weight: 600;
}

@media (max-width: 600px) {
  .action-buttons {
    flex-direction: column;
    align-items: stretch;
  }
  
  .action-button {
    min-width: unset;
    width: 100%;
    max-width: 400px;
    margin: 0 auto;
  }
}
</style>