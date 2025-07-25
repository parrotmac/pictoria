import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../api'
import type { Photo } from '../types'

export const usePhotosStore = defineStore('photos', () => {
  const photos = ref<Photo[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const activeEndpoint = ref<string | null>(null)

  async function fetchPhotos() {
    loading.value = true
    error.value = null
    try {
      const response = await api.get('/api/photos')
      photos.value = response.data.reverse() // Latest first
      // Update active endpoint after successful request
      activeEndpoint.value = await api.getActiveEndpoint()
    } catch (err: any) {
      if (!navigator.onLine) {
        error.value = 'You are offline. Photos cannot be loaded.'
      } else {
        error.value = 'Failed to load photos'
      }
      console.error('Failed to load photos:', err)
    } finally {
      loading.value = false
    }
  }

  async function uploadPhoto(name: string, file: File) {
    if (!navigator.onLine) {
      throw new Error('You are offline. Please connect to the internet to upload photos.')
    }

    const formData = new FormData()
    formData.append('photo', file)

    const response = await api.post(`/api/upload?username=${encodeURIComponent(name)}`, formData, {
      config: {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      },
    })

    // Update active endpoint after successful upload
    activeEndpoint.value = await api.getActiveEndpoint()

    // Refresh photos after upload
    await fetchPhotos()
    return response.data
  }

  function getPhotoUrl(photo: Photo): string {
    // Build the path part
    let path: string
    const originalName = photo.originalName.toLowerCase()
    if (originalName.endsWith('.heic') || originalName.endsWith('.heif')) {
      path = `/uploads/${photo.id}.jpg`
    } else {
      const lastDot = photo.originalName.lastIndexOf('.')
      const ext = lastDot !== -1 ? photo.originalName.substring(lastDot) : '.jpg'
      path = `/uploads/${photo.id}${ext}`
    }

    // Use the active endpoint if available
    if (activeEndpoint.value) {
      return `${activeEndpoint.value}${path}`
    }

    // Fallback to relative URL
    return path
  }

  // Periodically update the active endpoint
  async function updateActiveEndpoint() {
    activeEndpoint.value = await api.getActiveEndpoint()
  }

  return {
    photos,
    loading,
    error,
    activeEndpoint,
    fetchPhotos,
    uploadPhoto,
    getPhotoUrl,
    updateActiveEndpoint
  }
})
