<template>
  <div class="editor-layout">
    <template v-if="isPreviewMode">
      <BookPreview :bookId="bookId" @close="isPreviewMode = false" />
    </template>
    
    <template v-else>
      <header class="app-header">
        <div class="header-left">
          <button @click="$emit('close')" class="btn-back">⬅️ Back to Library</button>
          <h1>Novel Drafting App</h1>
          <span class="app-badge">v0.1 — MVP</span>
        </div>
        
        <div class="header-right">
          <button @click="isStoryBibleOpen = !isStoryBibleOpen" class="btn-story-bible" :class="{ active: isStoryBibleOpen }">
            📖 Story Bible
          </button>
          <button @click="handleEpubExport" class="btn-primary" :disabled="isExporting">
            {{ isExporting ? 'Exporting...' : 'Export EPUB' }}
          </button>
          <button @click="handlePdfExport" class="btn-primary" :disabled="isExportingPdf">
            {{ isExportingPdf ? 'Exporting...' : 'Export PDF' }}
          </button>
          <button @click="isPreviewMode = true" class="btn-primary">
            Preview Book
          </button>
        </div>
      </header>
      
      <main class="app-main">
        <ChapterSidebar 
          ref="sidebarRef"
          :bookId="bookId" 
          :activeChapterId="currentChapterId"
          @select="currentChapterId = $event"
          @preview="isPreviewMode = true"
        />
        
        <section class="editor-area">
          <TipTapEditor 
            v-if="currentChapterId" 
            :chapterId="currentChapterId" 
            :key="currentChapterId"
            @title-updated="onChapterTitleUpdated"
          />
          <div v-else class="empty-editor-state">
            <p>Select a chapter from the sidebar or create a new one to start writing.</p>
          </div>
        </section>

        <StoryBible
          :bookId="bookId"
          :isOpen="isStoryBibleOpen"
          @toggle="isStoryBibleOpen = !isStoryBibleOpen"
        />
      </main>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { get } from '../composables/useApi.js'
import ChapterSidebar from './ChapterSidebar.vue'
import TipTapEditor from './TipTapEditor.vue'
import BookPreview from './BookPreview.vue'
import StoryBible from './StoryBible.vue'
import { generatePdfExport, generateEpubExport } from '../utils/exportService.js'

const props = defineProps({
  bookId: {
    type: [String, Number],
    required: true
  }
})

const emit = defineEmits(['close'])

const currentChapterId = ref(null)
const isPreviewMode = ref(false)
const isStoryBibleOpen = ref(false)
const sidebarRef = ref(null)

const isExporting = ref(false)
const isExportingPdf = ref(false)

async function loadInitialChapter() {
  try {
    const chapters = await get(`/api/books/${props.bookId}/chapters`)
    if (chapters && chapters.length > 0) {
      currentChapterId.value = chapters[0].id
    } else {
      currentChapterId.value = null
    }
  } catch (err) {
    console.error('Failed to load chapters:', err)
  }
}

onMounted(() => {
  loadInitialChapter()
})

watch(() => props.bookId, () => {
  loadInitialChapter()
})

async function handlePdfExport() {
  isExportingPdf.value = true
  try {
    await generatePdfExport(props.bookId)
  } finally {
    isExportingPdf.value = false
  }
}

async function handleEpubExport() {
  isExporting.value = true
  try {
    await generateEpubExport(props.bookId)
  } finally {
    isExporting.value = false
  }
}

function onChapterTitleUpdated({ id, title }) {
  if (sidebarRef.value) {
    sidebarRef.value.updateChapterTitle(id, title)
  }
}
</script>

<style scoped>
.editor-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1.5rem;
  background: #18181b;
  border-bottom: 1px solid #27272a;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.btn-back {
  background: transparent;
  border: none;
  font-size: 1rem;
  color: #e4e4e7;
  cursor: pointer;
  padding: 0.5rem 1rem;
  display: flex;
  align-items: center;
  transition: transform 0.2s, background-color 0.2s;
  border-radius: 6px;
  font-weight: 500;
}

.btn-back:hover {
  background: #27272a;
  transform: translateX(-3px);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.app-header h1 {
  font-size: 1.125rem;
  font-weight: 600;
  margin: 0;
}

.app-badge {
  font-size: 0.7rem;
  padding: 0.2rem 0.6rem;
  border-radius: 999px;
  background: #3b0764;
  color: #c084fc;
}

.app-main {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.editor-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background-color: #0f0f12;
}

.empty-editor-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #71717a;
}
</style>
