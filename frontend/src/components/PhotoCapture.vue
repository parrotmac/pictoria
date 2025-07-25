<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-content camera-modal">
      <button class="modal-close" @click="$emit('close')">&times;</button>
      
      <div class="camera-interface">
        <!-- Camera Roll -->
        <CameraRoll />
        
        <!-- Camera View -->
        <div class="camera-container">
          <video 
            ref="videoEl" 
            autoplay 
            playsinline
            class="camera-video"
          ></video>
          
          <canvas ref="canvasEl" style="display: none;"></canvas>
          
          <!-- Camera Info -->
          <div v-if="cameras.length > 1" class="camera-info">
            <button class="switch-camera-btn" @click="switchCamera">
              <RotateCw :size="20" />
            </button>
          </div>
          
          <!-- Capture Button -->
          <div class="camera-controls">
            <button 
              class="capture-button" 
              @click="capturePhoto"
              :disabled="!cameraReady"
            >
              <div class="capture-button-inner"></div>
            </button>
          </div>
          
          <!-- Quick Actions -->
          <div class="quick-actions">
            <button @click="navigateToGallery" class="quick-action-btn">
              <Images :size="20" />
              <span>Gallery</span>
            </button>
          </div>
        </div>

        <!-- Status Messages -->
        <div v-if="uploadQueue.pendingUploads.length > 0" class="upload-status">
          <Loader2 :size="16" class="spinning" />
          <span>Uploading {{ uploadQueue.pendingUploads.length }} photo(s)...</span>
        </div>

        <div v-if="error" class="error-message">
          {{ error }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUploadQueueStore } from '../stores/uploadQueue'
import CameraRoll from './CameraRoll.vue'
import { RotateCw, Images, Loader2 } from 'lucide-vue-next'

const emit = defineEmits<{
  close: []
}>()

const router = useRouter()
const uploadQueue = useUploadQueueStore()

const videoEl = ref<HTMLVideoElement>()
const canvasEl = ref<HTMLCanvasElement>()
const stream = ref<MediaStream | null>(null)
const error = ref('')
const cameras = ref<MediaDeviceInfo[]>([])
const currentCameraId = ref<string>('')
const cameraReady = computed(() => !!stream.value)

onMounted(() => {
  loadCameras()
})

onUnmounted(() => {
  stopCamera()
})

async function loadCameras() {
  try {
    const devices = await navigator.mediaDevices.enumerateDevices()
    cameras.value = devices.filter(device => device.kind === 'videoinput')
    if (cameras.value.length > 0 && !currentCameraId.value) {
      currentCameraId.value = cameras.value[0].deviceId
    }
    startCamera()
  } catch (err) {
    console.error('Failed to load cameras:', err)
    error.value = 'Failed to access camera'
  }
}

async function startCamera() {
  try {
    stopCamera()
    
    const constraints: MediaStreamConstraints = { 
      video: currentCameraId.value 
        ? { 
            deviceId: { exact: currentCameraId.value },
            width: { ideal: 3840 },
            height: { ideal: 2160 }
          }
        : { 
            facingMode: 'environment',
            width: { ideal: 3840 },
            height: { ideal: 2160 }
          }
    }
    
    stream.value = await navigator.mediaDevices.getUserMedia(constraints)
    if (videoEl.value) {
      videoEl.value.srcObject = stream.value
    }
  } catch (err) {
    error.value = 'Failed to access camera. Please check permissions.'
    console.error('Camera error:', err)
  }
}

async function switchCamera() {
  if (cameras.value.length <= 1) return
  
  const currentIndex = cameras.value.findIndex(cam => cam.deviceId === currentCameraId.value)
  const nextIndex = (currentIndex + 1) % cameras.value.length
  currentCameraId.value = cameras.value[nextIndex].deviceId
  
  await startCamera()
}

function stopCamera() {
  if (stream.value) {
    stream.value.getTracks().forEach(track => track.stop())
    stream.value = null
  }
}

async function capturePhoto() {
  if (!videoEl.value || !canvasEl.value || !cameraReady.value) return
  
  const video = videoEl.value
  const canvas = canvasEl.value
  canvas.width = video.videoWidth
  canvas.height = video.videoHeight
  
  const context = canvas.getContext('2d')
  if (context) {
    context.drawImage(video, 0, 0)
    
    // Convert to blob and create file
    canvas.toBlob(async (blob) => {
      if (!blob) {
        error.value = 'Failed to capture photo'
        return
      }
      
      const file = new File([blob], `capture-${Date.now()}.jpg`, { type: 'image/jpeg' })
      const preview = URL.createObjectURL(blob)
      
      // Add to upload queue for background upload
      uploadQueue.addToQueue(file, preview)
      
      // Visual feedback - brief flash effect
      if (videoEl.value) {
        videoEl.value.style.opacity = '0.3'
        setTimeout(() => {
          if (videoEl.value) {
            videoEl.value.style.opacity = '1'
          }
        }, 100)
      }
    }, 'image/jpeg', 0.9)
  }
}

function navigateToGallery() {
  emit('close')
  router.push('/gallery')
}
</script>

<style scoped>
.camera-modal {
  max-width: 90vw;
  width: 700px;
}

.camera-interface {
  position: relative;
}

.camera-container {
  position: relative;
  text-align: center;
}

.camera-video {
  width: 100%;
  max-width: 600px;
  height: auto;
  border-radius: 10px;
  transition: opacity 0.1s ease;
  background: #000;
}

.camera-info {
  position: absolute;
  top: 20px;
  right: 20px;
  z-index: 10;
}

.switch-camera-btn {
  background: rgba(0, 0, 0, 0.6);
  color: white;
  border: none;
  border-radius: 50%;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.2s ease;
}

.switch-camera-btn:hover {
  background: rgba(0, 0, 0, 0.8);
}

.camera-controls {
  position: absolute;
  bottom: 30px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 10;
}

.capture-button {
  width: 70px;
  height: 70px;
  border-radius: 50%;
  border: 4px solid white;
  background: transparent;
  cursor: pointer;
  padding: 5px;
  transition: all 0.2s ease;
}

.capture-button:hover:not(:disabled) {
  transform: scale(1.05);
}

.capture-button:active:not(:disabled) {
  transform: scale(0.95);
}

.capture-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.capture-button-inner {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background: white;
  transition: all 0.1s ease;
}

.capture-button:active:not(:disabled) .capture-button-inner {
  background: #e0e0e0;
}

.quick-actions {
  position: absolute;
  bottom: 40px;
  left: 20px;
  z-index: 10;
}

.quick-action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  border: none;
  border-radius: 20px;
  padding: 8px 16px;
  cursor: pointer;
  transition: background 0.2s ease;
  font-size: 14px;
}

.quick-action-btn:hover {
  background: rgba(0, 0, 0, 0.8);
}

.upload-status {
  display: flex;
  align-items: center;
  gap: 8px;
  justify-content: center;
  margin-top: 15px;
  color: var(--text-secondary);
  font-size: 14px;
}

.error-message {
  color: var(--button-danger);
  margin-top: 10px;
  text-align: center;
  font-size: 14px;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.spinning {
  animation: spin 1s linear infinite;
}
</style>