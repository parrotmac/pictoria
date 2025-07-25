import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { usePhotosStore } from './photos'

export interface QueuedUpload {
  id: string
  file: File
  preview: string
  status: 'pending' | 'uploading' | 'completed' | 'failed'
  progress: number
  error?: string
  uploadedPhotoId?: string
  timestamp: number
}

export const useUploadQueueStore = defineStore('uploadQueue', () => {
  const photosStore = usePhotosStore()
  
  const queue = ref<QueuedUpload[]>([])
  const isProcessing = ref(false)
  const maxConcurrentUploads = 3

  // Get user's recent captures (pending + completed in last 5 minutes)
  const recentCaptures = computed(() => {
    const fiveMinutesAgo = Date.now() - 5 * 60 * 1000
    return queue.value
      .filter(item => item.timestamp > fiveMinutesAgo)
      .sort((a, b) => b.timestamp - a.timestamp)
  })

  const pendingUploads = computed(() => 
    queue.value.filter(item => item.status === 'pending')
  )

  const uploadingCount = computed(() =>
    queue.value.filter(item => item.status === 'uploading').length
  )

  function addToQueue(file: File, preview: string): string {
    const id = `upload-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    
    queue.value.push({
      id,
      file,
      preview,
      status: 'pending',
      progress: 0,
      timestamp: Date.now()
    })

    // Start processing if not already running
    if (!isProcessing.value) {
      processQueue()
    }

    return id
  }

  async function processQueue() {
    if (isProcessing.value) return
    isProcessing.value = true

    while (pendingUploads.value.length > 0 && uploadingCount.value < maxConcurrentUploads) {
      const nextUpload = pendingUploads.value[0]
      if (nextUpload) {
        uploadFile(nextUpload)
      }
    }

    // Continue processing if there are more pending uploads
    if (pendingUploads.value.length > 0) {
      setTimeout(() => {
        isProcessing.value = false
        processQueue()
      }, 1000)
    } else {
      isProcessing.value = false
    }
  }

  async function uploadFile(item: QueuedUpload) {
    const index = queue.value.findIndex(q => q.id === item.id)
    if (index === -1) return

    // Update status to uploading
    queue.value[index].status = 'uploading'
    queue.value[index].progress = 0

    try {
      // Check if online
      if (!navigator.onLine) {
        throw new Error('You are offline')
      }

      // Create upload with progress tracking
      const formData = new FormData()
      formData.append('photo', item.file)

      const xhr = new XMLHttpRequest()
      
      // Track upload progress
      xhr.upload.addEventListener('progress', (e) => {
        if (e.lengthComputable) {
          const progress = Math.round((e.loaded / e.total) * 100)
          const idx = queue.value.findIndex(q => q.id === item.id)
          if (idx !== -1) {
            queue.value[idx].progress = progress
          }
        }
      })

      // Handle completion
      const uploadPromise = new Promise((resolve, reject) => {
        xhr.addEventListener('load', () => {
          if (xhr.status >= 200 && xhr.status < 300) {
            try {
              const response = JSON.parse(xhr.responseText)
              resolve(response)
            } catch (err) {
              reject(new Error('Invalid response'))
            }
          } else {
            reject(new Error(`Upload failed: ${xhr.status}`))
          }
        })

        xhr.addEventListener('error', () => {
          reject(new Error('Network error'))
        })

        xhr.addEventListener('abort', () => {
          reject(new Error('Upload cancelled'))
        })
      })

      // Send request
      xhr.open('POST', '/api/upload')
      xhr.withCredentials = true
      xhr.send(formData)

      const response = await uploadPromise as any

      // Update item status
      const idx = queue.value.findIndex(q => q.id === item.id)
      if (idx !== -1) {
        queue.value[idx].status = 'completed'
        queue.value[idx].progress = 100
        queue.value[idx].uploadedPhotoId = response.id
      }

      // Refresh photos in the background
      photosStore.fetchPhotos()

    } catch (error: any) {
      // Mark as failed
      const idx = queue.value.findIndex(q => q.id === item.id)
      if (idx !== -1) {
        queue.value[idx].status = 'failed'
        queue.value[idx].error = error.message
        queue.value[idx].progress = 0
      }
    }
  }

  function retryUpload(id: string) {
    const item = queue.value.find(q => q.id === id)
    if (item && item.status === 'failed') {
      item.status = 'pending'
      item.error = undefined
      item.progress = 0
      
      if (!isProcessing.value) {
        processQueue()
      }
    }
  }

  function removeFromQueue(id: string) {
    const index = queue.value.findIndex(q => q.id === id)
    if (index !== -1) {
      // Revoke the preview URL to free memory
      URL.revokeObjectURL(queue.value[index].preview)
      queue.value.splice(index, 1)
    }
  }

  function clearOldUploads() {
    const oneHourAgo = Date.now() - 60 * 60 * 1000
    queue.value = queue.value.filter(item => {
      if (item.timestamp < oneHourAgo && item.status === 'completed') {
        URL.revokeObjectURL(item.preview)
        return false
      }
      return true
    })
  }

  // Clean up old uploads periodically
  setInterval(clearOldUploads, 5 * 60 * 1000) // Every 5 minutes

  return {
    queue,
    recentCaptures,
    pendingUploads,
    uploadingCount,
    addToQueue,
    retryUpload,
    removeFromQueue
  }
})