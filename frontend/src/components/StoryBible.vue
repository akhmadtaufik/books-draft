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

      <!-- Tabs -->
      <div class="tab-bar">
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'character' }"
          @click="activeTab = 'character'"
        >
          👤 Characters
        </button>
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'worldbuilding' }"
          @click="activeTab = 'worldbuilding'"
        >
          🌍 Worldbuilding
        </button>
      </div>

      <!-- Note Editing View -->
      <div v-if="editingNote" class="note-editor">
        <button class="btn-back" @click="editingNote = null">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"/></svg>
          Back
        </button>
        <input
          v-model="editingNote.title"
          @input="onNoteEdit"
          class="note-title-input"
          placeholder="Note title..."
        />
        <textarea
          v-model="editingNoteText"
          @input="onNoteEdit"
          class="note-content-textarea"
          placeholder="Write your notes here..."
        ></textarea>
        <div class="note-save-status">
          <span v-if="noteSaving" class="status-saving">Saving...</span>
          <span v-else-if="noteSaved" class="status-saved">Saved ✓</span>
        </div>
      </div>

      <!-- Note List View -->
      <div v-else class="note-list-area">
        <div v-if="isLoading" class="notes-loading">Loading notes...</div>
        <div v-else-if="filteredNotes.length === 0" class="notes-empty">
          <p>No {{ activeTab === 'character' ? 'character' : 'worldbuilding' }} notes yet.</p>
        </div>
        <div v-else class="note-cards">
          <div
            v-for="note in filteredNotes"
            :key="note.id"
            class="note-card"
            @click="openNote(note)"
          >
            <div class="note-card-content">
              <span class="note-card-icon">{{ activeTab === 'character' ? '👤' : '🌍' }}</span>
              <div class="note-card-info">
                <span class="note-card-title">{{ note.title }}</span>
                <span class="note-card-date">{{ formatDate(note.updatedAt) }}</span>
              </div>
            </div>
            <button class="btn-delete-note" @click.stop="confirmDelete(note)" title="Delete note">
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"/><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"/><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"/><line x1="10" x2="10" y1="11" y2="17"/><line x1="14" x2="14" y1="11" y2="17"/></svg>
            </button>
          </div>
        </div>

        <button class="btn-add-note" @click="createNote" :disabled="isCreating">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
          {{ activeTab === 'character' ? 'New Character' : 'New Lore' }}
        </button>
      </div>
    </div>

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
import { ref, computed, watch, onUnmounted } from 'vue'
import { get, post, put, del } from '../composables/useApi.js'
import DeleteModal from './DeleteModal.vue'

const props = defineProps({
  bookId: { type: String, required: true },
  isOpen: { type: Boolean, default: false }
})

const emit = defineEmits(['toggle'])

const activeTab = ref('character')
const notes = ref([])
const isLoading = ref(false)
const isCreating = ref(false)

// Editing state
const editingNote = ref(null)
const editingNoteText = ref('')
const noteSaving = ref(false)
const noteSaved = ref(false)

// Delete state
const isDeleteModalOpen = ref(false)
const noteToDelete = ref(null)

// Debounce timer
let saveTimer = null
const SAVE_DEBOUNCE_MS = 1500

const filteredNotes = computed(() => {
  return notes.value.filter(n => n.type === activeTab.value)
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

async function createNote() {
  isCreating.value = true
  const defaultTitle = activeTab.value === 'character' ? 'Untitled Character' : 'New Lore'
  try {
    const note = await post(`/api/books/${props.bookId}/notes`, {
      title: defaultTitle,
      type: activeTab.value,
      content: { text: '' }
    })
    notes.value.unshift(note)
    openNote(note)
  } catch (err) {
    console.error('Failed to create note:', err)
  } finally {
    isCreating.value = false
  }
}

function openNote(note) {
  editingNote.value = { ...note }
  // Extract text from JSONB content
  const content = note.content || {}
  editingNoteText.value = content.text || ''
  noteSaved.value = false
}

function onNoteEdit() {
  noteSaved.value = false
  clearTimeout(saveTimer)
  saveTimer = setTimeout(saveNote, SAVE_DEBOUNCE_MS)
}

async function saveNote() {
  if (!editingNote.value) return
  noteSaving.value = true
  try {
    const updated = await put(`/api/notes/${editingNote.value.id}`, {
      title: editingNote.value.title,
      content: { text: editingNoteText.value }
    })
    // Update local state
    const idx = notes.value.findIndex(n => n.id === updated.id)
    if (idx !== -1) {
      notes.value[idx] = updated
    }
    noteSaved.value = true
  } catch (err) {
    console.error('Failed to save note:', err)
  } finally {
    noteSaving.value = false
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
    // If we were editing this note, go back to list
    if (editingNote.value && editingNote.value.id === note.id) {
      editingNote.value = null
    }
  } catch (err) {
    console.error('Failed to delete note:', err)
  }
  noteToDelete.value = null
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

onUnmounted(() => {
  clearTimeout(saveTimer)
})
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

/* Tab bar */
.tab-bar {
  display: flex;
  border-bottom: 1px solid #27272a;
}

.tab-btn {
  flex: 1;
  padding: 0.75rem 0.5rem;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: #71717a;
  font-size: 0.8125rem;
  font-family: 'Inter', sans-serif;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.tab-btn:hover {
  color: #d4d4d8;
  background: rgba(255,255,255,0.02);
}

.tab-btn.active {
  color: #e4e4e7;
  border-bottom-color: #8b5cf6;
}

/* Note list */
.note-list-area {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.notes-loading, .notes-empty {
  text-align: center;
  color: #52525b;
  font-size: 0.8125rem;
  padding: 2rem 1rem;
}

.notes-empty p {
  margin: 0;
}

.note-cards {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.note-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.625rem 0.75rem;
  background: #18181b;
  border: 1px solid #27272a;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.note-card:hover {
  border-color: #3f3f46;
  background: #1c1c20;
  transform: translateX(-2px);
}

.note-card-content {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  min-width: 0;
  flex: 1;
}

.note-card-icon {
  font-size: 1.25rem;
  flex-shrink: 0;
}

.note-card-info {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  min-width: 0;
}

.note-card-title {
  font-size: 0.875rem;
  font-weight: 500;
  color: #e4e4e7;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.note-card-date {
  font-size: 0.6875rem;
  color: #52525b;
}

.btn-delete-note {
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
  flex-shrink: 0;
}

.note-card:hover .btn-delete-note {
  opacity: 1;
}

.btn-delete-note:hover {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.1);
}

.btn-add-note {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  background: transparent;
  border: 1px dashed #3f3f46;
  color: #a1a1aa;
  padding: 0.75rem;
  margin-top: 0.5rem;
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.8125rem;
  font-family: 'Inter', sans-serif;
  transition: all 0.2s;
}

.btn-add-note:hover:not(:disabled) {
  border-color: #8b5cf6;
  color: #c4b5fd;
  background: rgba(139, 92, 246, 0.05);
}

.btn-add-note:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Note editor */
.note-editor {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 1rem;
  gap: 0.75rem;
  overflow: hidden;
}

.btn-back {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  background: transparent;
  border: none;
  color: #71717a;
  font-size: 0.8125rem;
  font-family: 'Inter', sans-serif;
  cursor: pointer;
  padding: 0.25rem 0;
  width: fit-content;
  transition: color 0.2s;
}

.btn-back:hover {
  color: #c4b5fd;
}

.note-title-input {
  width: 100%;
  background: #18181b;
  border: 1px solid #27272a;
  border-radius: 6px;
  color: #f4f4f5;
  font-size: 1rem;
  font-weight: 600;
  font-family: 'Inter', sans-serif;
  padding: 0.625rem 0.75rem;
  outline: none;
  transition: border-color 0.2s;
}

.note-title-input:focus {
  border-color: #8b5cf6;
}

.note-content-textarea {
  flex: 1;
  width: 100%;
  background: #18181b;
  border: 1px solid #27272a;
  border-radius: 6px;
  color: #d4d4d8;
  font-size: 0.875rem;
  font-family: 'Lora', 'Georgia', serif;
  line-height: 1.7;
  padding: 0.75rem;
  outline: none;
  resize: none;
  transition: border-color 0.2s;
}

.note-content-textarea:focus {
  border-color: #8b5cf6;
}

.note-save-status {
  font-size: 0.75rem;
  font-family: 'Inter', sans-serif;
  height: 1rem;
}

.status-saving {
  color: #fbbf24;
}

.status-saved {
  color: #34d399;
}
</style>
