<template>
  <div class="preview-container">
    <div class="preview-header">
      <div class="preview-header-left">
        <h2>{{ previewData?.title || 'Book Preview' }}</h2>
        <div v-if="previewData && !isLoading" class="preview-stats">
          ⏱️ Est. Reading Time: {{ totalReadingTimeStr }} <span class="divider">|</span> 
          <div class="pagination-controls">
            📄 Pages 
            <form @submit.prevent="jumpToPage" class="page-jump-form">
              <input 
                type="number" 
                v-model.number="targetPageInput" 
                @blur="jumpToPage"
                class="page-input" 
                min="1" 
                :max="totalPages" 
                aria-label="Jump to page"
              />
            </form>
            - {{ pageRight }} of {{ totalPages }}
          </div>
        </div>
      </div>
      
      <div class="header-controls">
        <select 
          class="chapter-select" 
          v-model="selectedChapterIndex" 
          @change="jumpToChapter" 
          v-if="previewData?.chapters?.length > 0"
        >
          <option value="" disabled>Jump to Chapter...</option>
          <option v-for="(chapter, index) in previewData.chapters" :key="index" :value="index">
            {{ chapter.title || `Chapter ${index + 1}` }}
          </option>
        </select>

        <button @click="$emit('close')" class="btn-close">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
          <span>Close Preview</span>
        </button>
      </div>
    </div>

    <div v-if="isLoading" class="loading-state">Generating preview...</div>
    <div v-else-if="error" class="error-state">{{ error }}</div>
    
    <div v-else class="preview-content">
      <div v-if="previewData?.chapters.length === 0" class="empty-state">
        This book has no chapters yet.
      </div>
      
      <div v-else class="book-simulator-wrapper">
        <button class="nav-btn prev" @click="prevSpread" :disabled="currentSpreadIndex <= 0" aria-label="Previous Page">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"></polyline></svg>
        </button>

        <div class="book-viewport" ref="viewportContainer">
          <div 
            class="book-page-flow" 
            ref="flowContainer"
            :style="{ transform: `translateX(calc(${currentSpreadIndex} * (-100% - 8vw)))` }"
          >
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

        <button class="nav-btn next" @click="nextSpread" :disabled="currentSpreadIndex >= totalSpreads - 1" aria-label="Next Page">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"></polyline></svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, computed, nextTick, watch } from 'vue'
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

const viewportContainer = ref(null)
const flowContainer = ref(null)

const currentSpreadIndex = ref(0)
const totalSpreads = ref(1)

const selectedChapterIndex = ref('')
const chapterSpreadMap = ref([])

const totalPages = computed(() => totalSpreads.value * 2)
const pageLeft = computed(() => (currentSpreadIndex.value * 2) + 1)
const pageRight = computed(() => Math.min((currentSpreadIndex.value * 2) + 2, totalPages.value))

const targetPageInput = ref(1)

// Keep input synced with left page when navigating via Next/Prev buttons
watch(pageLeft, (newVal) => {
  targetPageInput.value = newVal
})

const jumpToPage = () => {
  let p = targetPageInput.value
  if (typeof p !== 'number' || isNaN(p)) p = 1
  if (p < 1) p = 1
  if (p > totalPages.value) p = totalPages.value
  targetPageInput.value = p // clamp to valid range

  const targetSpread = Math.floor((p - 1) / 2)
  currentSpreadIndex.value = targetSpread
}

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

const calculatePages = () => {
  if (!flowContainer.value || !viewportContainer.value) return
  
  const flowEl = flowContainer.value
  const vwWidth = viewportContainer.value.clientWidth
  
  // Dynamically calculate gap in pixels to handle window resizing properly
  const style = window.getComputedStyle(flowEl)
  const gapPx = parseFloat(style.columnGap) || 0
  const stride = vwWidth + gapPx
  
  // Total spreads is the required width divided by a single spread stride width
  totalSpreads.value = Math.max(1, Math.ceil((flowEl.scrollWidth + gapPx) / stride))
  
  // Bound the current index to the calculated spreads
  if (currentSpreadIndex.value >= totalSpreads.value) {
    currentSpreadIndex.value = Math.max(0, totalSpreads.value - 1)
  }

  // Calculate chapter starting spreads
  const chapterElements = flowEl.querySelectorAll('.preview-chapter')
  chapterSpreadMap.value = Array.from(chapterElements).map(el => {
    // offsetLeft gives the horizontal position of the column the element starts in
    return Math.floor(el.offsetLeft / stride)
  })
}

const jumpToChapter = () => {
  if (selectedChapterIndex.value === '' || chapterSpreadMap.value.length === 0) return
  const spreadIndex = chapterSpreadMap.value[selectedChapterIndex.value]
  
  if (spreadIndex !== undefined) {
    currentSpreadIndex.value = spreadIndex
  }
  // Optional: reset the select back to default, or leave it showing current chapter
  selectedChapterIndex.value = ''
}

const nextSpread = () => {
  if (currentSpreadIndex.value < totalSpreads.value - 1) {
    currentSpreadIndex.value++
  }
}

const prevSpread = () => {
  if (currentSpreadIndex.value > 0) {
    currentSpreadIndex.value--
  }
}

// Recalculate pages occasionally as TiTap finishes rendering images/content
let resizeObserver = null

onMounted(async () => {
  try {
    previewData.value = await get(`/api/books/${props.bookId}/preview`)
    
    // Wait for initial render
    await nextTick()
    
    // Calculate pages after a short delay to allow TiTap to fully mount
    setTimeout(() => {
      calculatePages()
      
      // Setup ResizeObserver to recalculate if the container size changes
      if (viewportContainer.value) {
        resizeObserver = new ResizeObserver(() => {
          calculatePages()
        })
        resizeObserver.observe(viewportContainer.value)
      }
    }, 300)
    
  } catch (err) {
    error.value = err.message
  } finally {
    isLoading.value = false
  }
})

onBeforeUnmount(() => {
  if (resizeObserver && viewportContainer.value) {
    resizeObserver.unobserve(viewportContainer.value)
    resizeObserver.disconnect()
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
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-family: 'Inter', sans-serif;
  font-size: 0.85rem;
  color: #52525b;
}

.divider {
  color: #d4d4d8;
  margin: 0 0.25rem;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.page-jump-form {
  display: inline-block;
  margin: 0;
}

.page-input {
  width: 45px;
  padding: 0.15rem;
  font-size: 0.85rem;
  font-family: 'Inter', sans-serif;
  text-align: center;
  border: 1px solid #e4e4e7;
  border-radius: 4px;
  background: #f4f4f5;
  color: #18181b;
  transition: all 0.2s;
}

.page-input:focus {
  outline: none;
  border-color: #a1a1aa;
  background: #fff;
  box-shadow: 0 0 0 2px rgba(161, 161, 170, 0.2);
}

/* Hide number input spinners for cleaner look */
.page-input::-webkit-outer-spin-button,
.page-input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
.page-input[type=number] {
  -moz-appearance: textfield;
}

.header-controls {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.chapter-select {
  padding: 0.4rem 0.8rem;
  font-family: 'Inter', sans-serif;
  font-size: 0.875rem;
  border: 1px solid #e4e4e7;
  border-radius: 6px;
  background-color: #fafafa;
  color: #18181b;
  cursor: pointer;
  outline: none;
  transition: border-color 0.2s;
}

.chapter-select:focus {
  border-color: #a1a1aa;
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
  display: flex;
  overflow: hidden;
  background: #fff;
}

.book-simulator-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  box-sizing: border-box;
  padding: 3rem 6vw;
  display: flex;
  align-items: center;
  justify-content: center;
}

.book-viewport {
  width: 100%;
  height: 100%;
  overflow: hidden;
  position: relative;
  /* Add perspective for potential 3D effects if needed */
  perspective: 2000px;
}

/* CSS Multi-Column layout */
.book-page-flow {
  display: block;
  height: 100%;
  width: 100%;
  column-count: 2;         /* Forces exactly 2 pages on screen */
  column-gap: 8vw;         /* Responsive middle book spine gap */
  column-rule: 1px solid rgba(0, 0, 0, 0.05); /* Subtle center line for the spine */
  column-fill: auto;
  /* Hardware accelerated transform sliding replaces native scroll */
  transition: transform 0.6s cubic-bezier(0.25, 1, 0.5, 1);
  font-family: 'Georgia', 'Lora', serif;
  box-sizing: border-box;
}

.preview-chapter {
  break-inside: auto;
}

.chapter-heading {
  font-family: 'Inter', sans-serif;
  font-size: 2.2rem;
  text-align: center;
  margin-top: 2rem;
  margin-bottom: 2.5rem;
  color: #18181b;
  break-after: avoid; 
  break-before: page; 
}

/* Ensure the first chapter doesn't have an empty page before it */
.preview-chapter:first-child .chapter-heading {
  break-before: auto;
}

.chapter-divider {
  text-align: center;
  margin: 3rem 0;
  color: #a1a1aa;
  letter-spacing: 0.5em;
  break-before: avoid; 
  break-after: page;
}

/* Navigation buttons */
.nav-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  background: #fff;
  border: 1px solid #e4e4e7;
  border-radius: 50%;
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 4px 6px rgba(0,0,0,0.05);
  z-index: 10;
  transition: all 0.2s;
  color: #52525b;
}

.nav-btn:hover:not(:disabled) {
  background: #f4f4f5;
  color: #18181b;
  transform: translateY(-50%) scale(1.05);
}

.nav-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.nav-btn.prev {
  left: 2vw; /* Safely placed outside the viewport on the padding */
}

.nav-btn.next {
  right: 2vw; /* Safely placed outside the viewport on the padding */
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
