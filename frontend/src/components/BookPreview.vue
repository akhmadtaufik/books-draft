<template>
  <div class="preview-container">
    <div class="preview-header">
      <div class="preview-header-left">
        <h2>{{ previewData?.title || 'Book Preview' }}</h2>
        <div v-if="previewData && !isLoading" class="preview-stats">
          Total Book Stats: {{ totalWords.toLocaleString() }} words | ⏱️ Approx. {{ totalReadingTimeStr }} read
        </div>
      </div>
      <button @click="$emit('close')" class="btn-close">
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
        <span>Close Preview</span>
      </button>
    </div>

    <div v-if="isLoading" class="loading-state">Generating preview...</div>
    <div v-else-if="error" class="error-state">{{ error }}</div>
    
    <div v-else class="preview-content">
      <div v-if="previewData?.chapters.length === 0" class="empty-state">
        This book has no chapters yet.
      </div>
      
      <div 
        v-for="(chapter, index) in previewData.chapters" 
        :key="index" 
        class="preview-chapter"
      >
        <h1 class="chapter-heading">{{ chapter.title || 'Untitled Chapter' }}</h1>
        <!-- We use a read-only TipTap editor to render the JSON content -->
        <PreviewRenderer :content="chapter.content" />
        
        <div v-if="index < previewData.chapters.length - 1" class="chapter-divider">***</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { get } from '../composables/useApi.js'
import { calculateReadingTime } from '../composables/useReadingTime.js'
import PreviewRenderer from './PreviewRenderer.vue'

const props = defineProps({
  bookId: { type: String, required: true }
})

const emit = defineEmits(['close'])

const previewData = ref(null)
const isLoading = ref(true)
const error = ref(null)

const totalWords = computed(() => {
  if (!previewData.value || !previewData.value.chapters) return 0
  let total = 0
  for (const chapter of previewData.value.chapters) {
    const stats = calculateReadingTime(chapter.content)
    total += stats.wordCount
  }
  return total
})

const totalReadingTimeStr = computed(() => {
  const readingTime = Math.max(1, Math.round(totalWords.value / 200))
  const hours = Math.floor(readingTime / 60)
  const mins = readingTime % 60
  
  let formattedTime = ''
  if (hours > 0) {
    formattedTime += `${hours} hour${hours > 1 ? 's' : ''} `
  }
  if (mins > 0 || hours === 0) {
    formattedTime += `${mins} min${mins > 1 ? 's' : ''}`
  }
  return formattedTime.trim()
})

onMounted(async () => {
  try {
    previewData.value = await get(`/api/books/${props.bookId}/preview`)
  } catch (err) {
    error.value = err.message
  } finally {
    isLoading.value = false
  }
})
</script>

<style scoped>
.preview-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: #f4f4f5; /* Light reading theme */
  color: #18181b;
}

.preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 2rem;
  background-color: #fff;
  border-bottom: 1px solid #e4e4e7;
  position: sticky;
  top: 0;
  z-index: 10;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
}

.preview-header-left {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.preview-header h2 {
  margin: 0;
  font-family: 'Inter', sans-serif;
  font-size: 1.25rem;
  font-weight: 600;
}

.preview-stats {
  font-family: 'Inter', sans-serif;
  font-size: 0.85rem;
  color: #52525b;
}

.btn-close {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: transparent;
  border: none;
  color: #52525b;
  cursor: pointer;
  font-family: 'Inter', sans-serif;
  font-size: 0.875rem;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  transition: all 0.2s;
}

.btn-close:hover {
  background-color: #f4f4f5;
  color: #18181b;
}

.preview-content {
  flex: 1;
  overflow-y: auto;
  padding: 4rem 15%;
  font-family: 'Georgia', 'Lora', serif;
}

.preview-chapter {
  max-width: 800px;
  margin: 0 auto;
}

.chapter-heading {
  font-family: 'Inter', sans-serif;
  font-size: 2.5rem;
  text-align: center;
  margin-bottom: 3rem;
  color: #18181b;
}

.chapter-divider {
  text-align: center;
  margin: 5rem 0;
  color: #a1a1aa;
  letter-spacing: 0.5em;
}

.loading-state, .error-state, .empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #71717a;
  font-family: 'Inter', sans-serif;
}

.error-state {
  color: #ef4444;
}
</style>
