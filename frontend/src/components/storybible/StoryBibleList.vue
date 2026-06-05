<template>
  <div class="note-list-area">
    <div v-if="isLoading" class="notes-loading">Loading notes...</div>
    <div v-else-if="notes.length === 0" class="notes-empty">
      <p>No notes in this category yet.</p>
    </div>
    <div v-else class="note-cards">
      <div
        v-for="note in notes"
        :key="note.id"
        class="note-card"
        @click="$emit('edit', note)"
      >
        <div class="note-card-content">
          <span class="note-card-icon">{{ getIcon(note.type) }}</span>
          <div class="note-card-info">
            <span class="note-card-title">{{ note.title }}</span>
            <span class="note-card-date">{{ formatDate(note.updatedAt) }}</span>
          </div>
        </div>
        <button class="btn-delete-note" @click.stop="$emit('delete', note)" title="Delete note">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"/><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"/><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"/><line x1="10" x2="10" y1="11" y2="17"/><line x1="14" x2="14" y1="11" y2="17"/></svg>
        </button>
      </div>
    </div>
    
    <button class="btn-add-note" @click="$emit('create')" :disabled="isCreating">
      <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
      New Note
    </button>
  </div>
</template>

<script setup>
defineProps({
  notes: { type: Array, default: () => [] },
  isLoading: Boolean,
  isCreating: Boolean,
  activeCategory: String
})
defineEmits(['edit', 'delete', 'create'])

function formatDate(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function getIcon(type) {
  if (type === 'character') return '👤'
  if (type === 'worldbuilding') return '🌍'
  return '📝'
}
</script>

<style scoped>
.note-list-area { flex: 1; overflow-y: auto; padding: 1rem; display: flex; flex-direction: column; gap: 0.5rem; }
.notes-loading, .notes-empty { text-align: center; color: #52525b; font-size: 0.8125rem; padding: 2rem 1rem; }
.notes-empty p { margin: 0; }
.note-cards { display: flex; flex-direction: column; gap: 0.375rem; }
.note-card { display: flex; align-items: center; justify-content: space-between; padding: 0.625rem 0.75rem; background: #18181b; border: 1px solid #27272a; border-radius: 8px; cursor: pointer; transition: all 0.2s; }
.note-card:hover { border-color: #3f3f46; background: #1c1c20; transform: translateX(-2px); }
.note-card-content { display: flex; align-items: center; gap: 0.625rem; min-width: 0; flex: 1; }
.note-card-icon { font-size: 1.25rem; flex-shrink: 0; }
.note-card-info { display: flex; flex-direction: column; gap: 0.125rem; min-width: 0; }
.note-card-title { font-size: 0.875rem; font-weight: 500; color: #e4e4e7; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.note-card-date { font-size: 0.6875rem; color: #52525b; }
.btn-delete-note { background: transparent; border: none; color: #52525b; cursor: pointer; padding: 0.25rem; border-radius: 4px; display: flex; align-items: center; opacity: 0; transition: all 0.2s; flex-shrink: 0; }
.note-card:hover .btn-delete-note { opacity: 1; }
.btn-delete-note:hover { color: #ef4444; background: rgba(239, 68, 68, 0.1); }
.btn-add-note { display: flex; align-items: center; justify-content: center; gap: 0.5rem; background: transparent; border: 1px dashed #3f3f46; color: #a1a1aa; padding: 0.75rem; margin-top: 0.5rem; border-radius: 8px; cursor: pointer; font-size: 0.8125rem; font-family: 'Inter', sans-serif; transition: all 0.2s; }
.btn-add-note:hover:not(:disabled) { border-color: #8b5cf6; color: #c4b5fd; background: rgba(139, 92, 246, 0.05); }
.btn-add-note:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
