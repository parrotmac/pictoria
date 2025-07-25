<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-content">
      <button class="modal-close" @click="$emit('close')">&times;</button>
      <h2>Upload Photo</h2>
      
      <div v-if="!uploadSuccess">
        <div 
          class="upload-area"
          :class="{ dragover: isDragging, disabled: uploading }"
          @dragover.prevent="isDragging = true"
          @dragleave.prevent="isDragging = false"
          @drop.prevent="handleDrop"
        >
          <p>Drag and drop your photos here or</p>
          <input 
            type="file" 
            ref="fileInput" 
            class="file-input"
            @change="handleFileSelect"
            accept="image/*"
            multiple
            :disabled="uploading"
          />
          <button class="btn" @click="(($refs.fileInput as HTMLInputElement).click())" :disabled="uploading">
            Choose Files
          </button>
        </div>

        <div v-if="uploading" class="upload-progress">
          <div class="spinner"></div>
          <p v-if="uploadProgress.total > 1">
            Uploading {{ uploadProgress.current }} of {{ uploadProgress.total }} files...
          </p>
          <p v-else-if="selectedFiles.length > 0">
            Uploading {{ selectedFiles[0].name }}...
          </p>
        </div>

        <div v-if="error" class="error-message">
          {{ error }}
        </div>
      </div>

      <div v-else class="success-message">
        <p>✓ Photos uploaded successfully!</p>
        <button class="btn btn-success" @click="viewInGallery">
          View in Gallery
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { usePhotosStore } from '../stores/photos'

const emit = defineEmits<{
  close: []
  uploaded: [photoId: string]
}>()

const photosStore = usePhotosStore()

const isDragging = ref(false)
const selectedFiles = ref<File[]>([])
const uploading = ref(false)
const error = ref('')
const uploadSuccess = ref(false)
const uploadProgress = ref({ current: 0, total: 0 })
const lastUploadedPhotoId = ref<string>('')

function handleDrop(e: DragEvent) {
  isDragging.value = false
  const files = Array.from(e.dataTransfer?.files || [])
  if (files.length > 0) {
    selectedFiles.value = files
    uploadFiles()
  }
}

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
    try {
      const result = await photosStore.uploadPhoto(file)
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

function viewInGallery() {
  emit('uploaded', lastUploadedPhotoId.value)
  emit('close')
}
</script>

<style scoped>
.upload-area {
  border: 2px dashed var(--button-primary);
  border-radius: 10px;
  padding: 40px;
  text-align: center;
  margin: 20px 0;
  background: var(--upload-area-bg);
  transition: all 0.3s ease;
}

.upload-area.dragover {
  background: var(--upload-area-hover);
  border-color: var(--border-hover);
}

.upload-area.disabled {
  opacity: 0.5;
  pointer-events: none;
}

.file-input {
  display: none;
}

.upload-progress {
  margin-top: 20px;
  text-align: center;
}

.error-message {
  color: var(--button-danger);
  margin-top: 10px;
}

.success-message {
  margin-top: 20px;
  padding: 15px;
  background: #d4edda;
  border: 1px solid #c3e6cb;
  border-radius: 5px;
  text-align: center;
}

.success-message p {
  color: #155724;
  margin: 0 0 10px 0;
  font-weight: 500;
}
</style>
