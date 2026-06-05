<template>
  <aside class="story-bible" :class="{ open: isOpen }">
    <!-- Toggle Button (always visible) -->
    <button class="toggle-btn" @click="$emit('toggle')" :title="isOpen ? 'Close Story Bible' : 'Open Story Bible'">
      <svg v-if="isOpen" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
      <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"/></svg>
    </button>

    <!-- Panel Content (only when open) -->
    <div v-if="isOpen" class="panel-content">
      <div class="panel-header">
        <h2>Story Bible</h2>
      </div>

      <StoryBibleSidebar 
        :categories="categories" 
        :activeCategory="activeCategory" 
        @select="activeCategory = $event" 
      />

      <StoryBibleList 
        :notes="filteredNotes" 
        :isLoading="isLoading" 
        :isCreating="isCreating"
        :activeCategory="activeCategory"
        @create="openCreateModal" 
        @edit="openEditModal" 
        @delete="confirmDelete" 
      />
    </div>

    <!-- Note Modal -->
    <StoryBibleModal 
      :isOpen="isModalOpen" 
      :initialData="editingNote" 
      @close="isModalOpen = false" 
      @save="saveNote" 
    />

    <!-- Delete Confirmation Modal -->
    <DeleteModal
      :isOpen="isDeleteModalOpen"
      title="Delete Note"
      message="Are you sure you want to delete this note? This action cannot be undone."
      @confirm="executeDelete"
      @cancel="isDeleteModalOpen = false; noteToDelete = null"
    />
  </aside>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { get, post, put, del } from '../composables/useApi.js'
import StoryBibleSidebar from './storybible/StoryBibleSidebar.vue'
import StoryBibleList from './storybible/StoryBibleList.vue'
import StoryBibleModal from './storybible/StoryBibleModal.vue'
import DeleteModal from './DeleteModal.vue'

const props = defineProps({
  bookId: { type: [String, Number], required: true },
  isOpen: { type: Boolean, default: false }
})

const emit = defineEmits(['toggle'])

const categories = [
  { id: 'character', label: 'Characters', icon: '👤' },
  { id: 'worldbuilding', label: 'Worldbuilding', icon: '🌍' }
]
const activeCategory = ref('character')

const notes = ref([])
const isLoading = ref(false)
const isCreating = ref(false)

const isModalOpen = ref(false)
const editingNote = ref(null)

const isDeleteModalOpen = ref(false)
const noteToDelete = ref(null)

const filteredNotes = computed(() => {
  return notes.value.filter(n => n.type === activeCategory.value)
})

async function fetchNotes() {
  if (!props.bookId) return
  isLoading.value = true
  try {
    notes.value = await get(`/api/books/${props.bookId}/notes`) || []
  } catch (err) {
    console.error('Failed to fetch notes:', err)
  } finally {
    isLoading.value = false
  }
}

watch(() => props.bookId, fetchNotes)
watch(() => props.isOpen, (open) => {
  if (open && notes.value.length === 0) {
    fetchNotes()
  }
})

// Fetch on mount if already open
if (props.isOpen && props.bookId) {
  fetchNotes()
}

function openCreateModal() {
  editingNote.value = null
  isModalOpen.value = true
}

function openEditModal(note) {
  editingNote.value = { ...note }
  isModalOpen.value = true
}

async function saveNote(payload) {
  if (editingNote.value?.id) {
    // Update existing
    try {
      const updated = await put(`/api/notes/${editingNote.value.id}`, payload)
      const idx = notes.value.findIndex(n => n.id === updated.id)
      if (idx !== -1) notes.value[idx] = updated
      isModalOpen.value = false
    } catch (err) {
      console.error('Failed to update note:', err)
    }
  } else {
    // Create new
    isCreating.value = true
    try {
      const newNote = await post(`/api/books/${props.bookId}/notes`, {
        ...payload,
        type: activeCategory.value
      })
      notes.value.unshift(newNote)
      isModalOpen.value = false
    } catch (err) {
      console.error('Failed to create note:', err)
    } finally {
      isCreating.value = false
    }
  }
}

function confirmDelete(note) {
  noteToDelete.value = note
  isDeleteModalOpen.value = true
}

async function executeDelete() {
  const note = noteToDelete.value
  if (!note) return
  isDeleteModalOpen.value = false

  try {
    await del(`/api/notes/${note.id}`)
    notes.value = notes.value.filter(n => n.id !== note.id)
  } catch (err) {
    console.error('Failed to delete note:', err)
  }
  noteToDelete.value = null
}
</script>

<style scoped>
.story-bible {
  position: relative;
  display: flex;
  flex-direction: row;
  height: 100%;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  width: 0;
}

.story-bible.open {
  width: 320px;
}

/* Toggle button */
.toggle-btn {
  position: absolute;
  left: -32px;
  top: 50%;
  transform: translateY(-50%);
  width: 32px;
  height: 32px;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 6px 0 0 6px;
  color: #a1a1aa;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 50;
  transition: all 0.2s;
}

.toggle-btn:hover {
  color: #e4e4e7;
  background: rgba(39, 39, 42, 0.8);
}

/* Panel content */
.panel-content {
  margin-left: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #121214;
  border-left: 1px solid #27272a;
  overflow: hidden;
  width: 320px;
  min-width: 320px;
}

.panel-header {
  padding: 1rem 1.25rem;
  border-bottom: 1px solid #27272a;
}

.panel-header h2 {
  font-size: 0.875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #a1a1aa;
  margin: 0;
}
</style>
