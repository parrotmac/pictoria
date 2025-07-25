<template>
  <div class="container">
    <div class="header">
      <h1>Parker's Wedding 📸</h1>
      <small>Upload or capture your photos</small>
      <p v-if="nameStore" class="welcome-text">
        Welcome, {{ nameStore }}!
        <span @click="nameStore = ''" style="text-decoration:underline;cursor:pointer">Change</span>
      </p>
    </div>

    <div class="action-buttons">
      <div class="action-button" @click="showUploadModal = true">
        <Upload :size="50" />
        <h2>Upload Photo</h2>
        <p>Choose from your device</p>
      </div>

      <input 
        type="file" 
        ref="fileInput" 
        class="file-input"
        @change="handleFileSelect"
        accept="image/*"
        style="display:none"
        multiple
        capture
        :disabled="uploading"
      />

      <div class="action-button" @click="(($refs.fileInput as HTMLInputElement).click())" :disabled="uploading">
        <Camera :size="50" />
        <h2>Capture Photo</h2>
        <p>Use your camera</p>
      </div>
    </div>

    <div class="action-buttons">
      <div v-if="uploading" class="upload-progress">
        <div class="spinner"></div>
        <p v-if="uploadProgress.total > 1">
          Uploading {{ uploadProgress.current }} of {{ uploadProgress.total }} files...
        </p>
        <p v-else-if="selectedFiles.length > 0">
          Uploading {{ selectedFiles[0].name }}...
        </p>
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
import { nameStore } from '../stores/auth'
import { Upload, Camera } from 'lucide-vue-next'
import PhotoUpload from '../components/PhotoUpload.vue'
import PhotoCapture from '../components/PhotoCapture.vue'
import { usePhotosStore } from '../stores/photos'

const router = useRouter()

const showUploadModal = ref(false)
const showCameraModal = ref(false)

function handlePhotoUploaded(photoId: string) {
  // Navigate to gallery with the photo ID
  router.push({ 
    path: '/gallery', 
    hash: `#photo-${photoId}` 
  })
}

const uploading = ref(false)
const error = ref('')
const selectedFiles = ref<File[]>([])
const uploadSuccess = ref(false)
const uploadProgress = ref({ current: 0, total: 0 })
const photosStore = usePhotosStore()
const lastUploadedPhotoId = ref<string>('')

function handleFileSelect(e: Event) {
  const target = e.target as HTMLInputElement
  const files = Array.from(target.files || [])
  if (files.length > 0) {
    selectedFiles.value = files
    uploadFiles()
  }
}

async function uploadFiles() {
  if (selectedFiles.value.length === 0) return

  uploading.value = true
  error.value = ''
  uploadSuccess.value = false
  uploadProgress.value.total = selectedFiles.value.length
  uploadProgress.value.current = 0

  const failedUploads: string[] = []

  for (const file of selectedFiles.value) {
      
    const a = document.createElement(`a`)

    var reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = function () {
      a.href = reader.result as string
      console.log(a.href)
      a.download = file.name
      a.dispatchEvent(new MouseEvent(`click`))
    };

    try {
      const result = await photosStore.uploadPhoto(nameStore.value, file)
      uploadProgress.value.current++
      if (result?.id) {
        lastUploadedPhotoId.value = result.id
      }
    } catch (err) {
      failedUploads.push(file.name)
      console.error('Upload error for', file.name, ':', err)
    }
  }

  if (failedUploads.length > 0) {
    error.value = `Failed to upload: ${failedUploads.join(', ')}`
  } else {
    uploadSuccess.value = true
  }
  
  uploading.value = false
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
