<template>
  <aside class="sidebar">
    <div class="sidebar-header">
      <h2>Chapters</h2>
      <button @click="$emit('preview')" class="btn-preview" title="Preview Book">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"/><circle cx="12" cy="12" r="3"/></svg>
      </button>
    </div>

    <div v-if="isLoading" class="loading-state">Loading chapters...</div>
    <div v-else-if="error" class="error-state">{{ error }}</div>
    
    <div v-else class="chapter-list-container">
      <VueDraggable
        v-model="chapters"
        :animation="150"
        handle=".drag-handle"
        @end="onReorder"
        class="chapter-list"
      >
        <div
          v-for="chapter in chapters"
          :key="chapter.id"
          class="chapter-item"
          :class="{ active: chapter.id === activeChapterId }"
          @click="$emit('select', chapter.id)"
        >
          <div class="drag-handle" title="Drag to reorder" @click.stop>
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="8" x2="8" y1="6" y2="6"/><line x1="16" x2="16" y1="6" y2="6"/><line x1="8" x2="8" y1="12" y2="12"/><line x1="16" x2="16" y1="12" y2="12"/><line x1="8" x2="8" y1="18" y2="18"/><line x1="16" x2="16" y1="18" y2="18"/></svg>
          </div>
          <span class="chapter-title">{{ chapter.title || 'Untitled' }}</span>
          <button class="btn-delete" @click.stop="confirmDelete(chapter.id)" title="Delete Chapter">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"/><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"/><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"/><line x1="10" x2="10" y1="11" y2="17"/><line x1="14" x2="14" y1="11" y2="17"/></svg>
          </button>
        </div>
      </VueDraggable>

      <button class="btn-add-chapter" @click="createChapter" :disabled="isCreating">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
        New Chapter
      </button>
    </div>

    <!-- Delete Confirmation Modal -->
    <DeleteModal
      :isOpen="isDeleteModalOpen"
      title="Delete Chapter"
      message="Are you sure you want to delete this chapter? This cannot be undone."
      @confirm="executeDelete"
      @cancel="isDeleteModalOpen = false; itemToDeleteId = null"
    />
  </aside>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'
import { get, post, put, del } from '../composables/useApi.js'
import DeleteModal from './DeleteModal.vue'

const props = defineProps({
  bookId: { type: String, required: true },
  activeChapterId: { type: String, default: null }
})

const emit = defineEmits(['select', 'preview'])

const chapters = ref([])
const isLoading = ref(false)
const error = ref(null)
const isCreating = ref(false)

const isDeleteModalOpen = ref(false)
const itemToDeleteId = ref(null)

async function fetchChapters() {
  if (!props.bookId) return
  isLoading.value = true
  error.value = null
  try {
    chapters.value = await get(`/api/books/${props.bookId}/chapters`) || []
  } catch (err) {
    error.value = err.message
  } finally {
    isLoading.value = false
  }
}

watch(() => props.bookId, fetchChapters)
onMounted(fetchChapters)

async function onReorder() {
  const orderedIds = chapters.value.map(c => c.id)
  try {
    await put('/api/chapters/reorder', {
      bookId: props.bookId,
      orderedIds
    })
  } catch (err) {
    console.error('Failed to reorder chapters:', err)
    // Revert logic could be added here on failure
  }
}

async function createChapter() {
  isCreating.value = true
  try {
    const newChapter = await post(`/api/books/${props.bookId}/chapters`, {
      title: 'Untitled Chapter'
    })
    chapters.value.push({
      id: newChapter.id,
      title: newChapter.title,
      positionIndex: newChapter.positionIndex
    })
    emit('select', newChapter.id)
  } catch (err) {
    console.error('Failed to create chapter:', err)
  } finally {
    isCreating.value = false
  }
}

function confirmDelete(chapterId) {
  itemToDeleteId.value = chapterId
  isDeleteModalOpen.value = true
}

async function executeDelete() {
  const chapterId = itemToDeleteId.value
  if (!chapterId) return
  isDeleteModalOpen.value = false
  
  try {
    await del(`/api/chapters/${chapterId}`)
    chapters.value = chapters.value.filter(c => c.id !== chapterId)
    if (props.activeChapterId === chapterId) {
      const nextChapter = chapters.value[0]
      emit('select', nextChapter ? nextChapter.id : null)
    }
  } catch (err) {
    console.error('Failed to delete chapter:', err)
  }
}

defineExpose({
  updateChapterTitle: (id, newTitle) => {
    const chapter = chapters.value.find(c => c.id === id)
    if (chapter) chapter.title = newTitle
  }
})
</script>

<style scoped>
.sidebar {
  display: flex;
  flex-direction: column;
  background-color: #121214;
  border-right: 1px solid #27272a;
  color: #e4e4e7;
  height: 100%;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid #27272a;
}

.sidebar-header h2 {
  font-size: 0.875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #a1a1aa;
  margin: 0;
}

.btn-preview {
  background: transparent;
  border: none;
  color: #a1a1aa;
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.btn-preview:hover {
  color: #fff;
  background-color: #27272a;
}

.chapter-list-container {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.chapter-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 0.5rem;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
  user-select: none;
}

.chapter-item:hover {
  background-color: #18181b;
}

.chapter-item.active {
  background-color: #27272a;
  color: #fff;
  font-weight: 500;
}

.drag-handle {
  cursor: grab;
  color: #52525b;
  display: flex;
  align-items: center;
  padding: 0.25rem;
}

.drag-handle:hover {
  color: #a1a1aa;
}

.drag-handle:active {
  cursor: grabbing;
}

.chapter-title {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 0.9375rem;
}

.btn-delete {
  background: transparent;
  border: none;
  color: #52525b;
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 4px;
  display: flex;
  align-items: center;
  opacity: 0;
  transition: all 0.2s;
}

.chapter-item:hover .btn-delete {
  opacity: 1;
}

.btn-delete:hover {
  color: #ef4444;
  background-color: rgba(239, 68, 68, 0.1);
}

.btn-add-chapter {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  background-color: transparent;
  border: 1px dashed #3f3f46;
  color: #a1a1aa;
  padding: 0.75rem;
  margin-top: 0.5rem;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.btn-add-chapter:hover:not(:disabled) {
  border-color: #71717a;
  color: #fff;
  background-color: #18181b;
}

.btn-add-chapter:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.loading-state, .error-state {
  padding: 1.5rem;
  text-align: center;
  font-size: 0.875rem;
  color: #71717a;
}

.error-state {
  color: #ef4444;
}
</style>
