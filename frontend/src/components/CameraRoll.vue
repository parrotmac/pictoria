<template>
  <div class="camera-roll" v-if="uploadQueue.recentCaptures.length > 0">
    <div class="camera-roll-header">
      <h3>Recent Captures</h3>
      <span class="count">{{ uploadQueue.recentCaptures.length }}</span>
    </div>
    
    <div class="camera-roll-items">
      <div 
        v-for="item in uploadQueue.recentCaptures" 
        :key="item.id"
        class="roll-item"
        :class="{ 
          'uploading': item.status === 'uploading',
          'failed': item.status === 'failed'
        }"
      >
        <img :src="item.preview" :alt="item.file.name" />
        
        <!-- Upload status overlay -->
        <div v-if="item.status === 'uploading'" class="status-overlay">
          <div class="progress-ring">
            <svg viewBox="0 0 36 36">
              <path
                class="progress-ring-bg"
                d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
              />
              <path
                class="progress-ring-bar"
                :stroke-dasharray="`${item.progress}, 100`"
                d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
              />
            </svg>
          </div>
        </div>
        
        <div v-else-if="item.status === 'pending'" class="status-overlay">
          <Loader2 :size="24" class="spinning" />
        </div>
        
        <div v-else-if="item.status === 'failed'" class="status-overlay failed">
          <AlertCircle :size="20" />
          <button @click="uploadQueue.retryUpload(item.id)" class="retry-btn">
            Retry
          </button>
        </div>
        
        <div v-else-if="item.status === 'completed'" class="status-overlay completed">
          <Check :size="24" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useUploadQueueStore } from '../stores/uploadQueue'
import { Loader2, AlertCircle, Check } from 'lucide-vue-next'

const uploadQueue = useUploadQueueStore()
</script>

<style scoped>
.camera-roll {
  background: var(--bg-secondary);
  border-radius: 10px;
  padding: 15px;
  margin-bottom: 20px;
}

.camera-roll-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.camera-roll-header h3 {
  font-size: 1rem;
  color: var(--text-primary);
  margin: 0;
}

.count {
  background: var(--button-primary);
  color: white;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 0.8rem;
}

.camera-roll-items {
  display: flex;
  gap: 10px;
  overflow-x: auto;
  padding-bottom: 5px;
}

.camera-roll-items::-webkit-scrollbar {
  height: 6px;
}

.camera-roll-items::-webkit-scrollbar-track {
  background: var(--border-color);
  border-radius: 3px;
}

.camera-roll-items::-webkit-scrollbar-thumb {
  background: var(--button-primary);
  border-radius: 3px;
}

.roll-item {
  position: relative;
  width: 80px;
  height: 80px;
  flex-shrink: 0;
  border-radius: 8px;
  overflow: hidden;
  border: 2px solid transparent;
  transition: all 0.2s ease;
}

.roll-item.uploading {
  border-color: var(--button-primary);
}

.roll-item.failed {
  border-color: var(--button-danger);
}

.roll-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.status-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: white;
}

.status-overlay.completed {
  background: rgba(46, 204, 113, 0.8);
  animation: fadeOut 1s ease-out forwards;
  animation-delay: 0.5s;
}

@keyframes fadeOut {
  to {
    opacity: 0;
  }
}

.progress-ring {
  width: 36px;
  height: 36px;
}

.progress-ring svg {
  transform: rotate(-90deg);
}

.progress-ring-bg {
  fill: none;
  stroke: rgba(255, 255, 255, 0.3);
  stroke-width: 3;
}

.progress-ring-bar {
  fill: none;
  stroke: white;
  stroke-width: 3;
  stroke-linecap: round;
  transition: stroke-dasharray 0.3s ease;
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.retry-btn {
  margin-top: 5px;
  padding: 2px 8px;
  font-size: 0.7rem;
  background: var(--button-danger);
  color: white;
  border: none;
  border-radius: 3px;
  cursor: pointer;
}

.retry-btn:hover {
  background: var(--button-danger-hover);
}
</style>