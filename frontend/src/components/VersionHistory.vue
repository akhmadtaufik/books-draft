<template>
  <Teleport to="body">
    <div class="vh-overlay" @click.self="$emit('close')">
      <div class="vh-panel">
        <!-- Header -->
        <div class="vh-header">
          <div class="vh-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <polyline points="12,6 12,12 16,14"/>
            </svg>
            Revision History
          </div>
          <button class="vh-close" @click="$emit('close')" aria-label="Close">✕</button>
        </div>

        <div class="vh-body">
          <!-- Versions List -->
          <div class="vh-list-pane">
            <div v-if="isLoading" class="vh-status">Loading versions...</div>
            <div v-else-if="versions.length === 0" class="vh-status vh-empty">
              <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                <polyline points="14,2 14,8 20,8"/>
              </svg>
              <p>No versions saved yet.</p>
              <p class="vh-hint">Use "Save Milestone" or navigate away to create a session snapshot.</p>
            </div>
            <div
              v-for="version in versions"
              :key="version.id"
              class="vh-version-item"
              :class="{ 'is-selected': selectedVersion?.id === version.id }"
              @click="selectVersion(version)"
            >
              <div class="vh-version-item-top">
                <div class="vh-version-badge" :class="version.snapshotType === 'manual_milestone' ? 'badge-manual' : 'badge-session'">
                  {{ version.snapshotType === 'manual_milestone' ? '⭐ Milestone' : '⏱ Session' }}
                </div>
                <button
                  class="vh-delete-btn"
                  :class="{ 'is-deleting': deletingId === version.id }"
                  :disabled="deletingId === version.id"
                  @click.stop="deleteVersion(version)"
                  title="Delete this revision"
                  aria-label="Delete revision"
                >
                  <svg v-if="deletingId !== version.id" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                    <polyline points="3,6 5,6 21,6"/>
                    <path d="M19,6l-1,14a2,2,0,0,1-2,2H8a2,2,0,0,1-2-2L5,6"/>
                    <path d="M10,11v6"/>
                    <path d="M14,11v6"/>
                    <path d="M9,6V4a1,1,0,0,1,1-1h4a1,1,0,0,1,1,1v2"/>
                  </svg>
                  <span v-else class="vh-deleting-spinner">⏳</span>
                </button>
              </div>
              <div class="vh-version-time">{{ formatDate(version.createdAt) }}</div>
              <div class="vh-version-relative">{{ relativeTime(version.createdAt) }}</div>
            </div>
          </div>

          <!-- Version Preview Pane -->
          <div class="vh-preview-pane">
            <div v-if="!selectedVersion" class="vh-status vh-empty">
              <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                <circle cx="12" cy="12" r="3"/>
              </svg>
              <p>Select a version to preview its content.</p>
            </div>
            <div v-else-if="isLoadingContent" class="vh-status">Loading content...</div>
            <div v-else class="vh-preview-content">
              <div class="vh-preview-meta">
                <span class="vh-version-badge" :class="selectedVersion.snapshotType === 'manual_milestone' ? 'badge-manual' : 'badge-session'">
                  {{ selectedVersion.snapshotType === 'manual_milestone' ? '⭐ Milestone' : '⏱ Session End' }}
                </span>
                <span class="vh-preview-date">{{ formatDate(selectedVersion.createdAt) }}</span>
              </div>
              <div class="vh-tiptap-wrapper">
                <editor-content v-if="previewEditor" :editor="previewEditor" class="vh-tiptap-readonly" />
              </div>
              <div class="vh-preview-actions">
                <button class="btn-restore" @click="restoreVersion" :disabled="isRestoring">
                  {{ isRestoring ? 'Restoring...' : '↩ Restore this version' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Delete Confirmation Modal -->
    <DeleteModal
      :isOpen="isDeleteModalOpen"
      :title="deleteModalTitle"
      :message="deleteModalMessage"
      @confirm="executeDelete"
      @cancel="isDeleteModalOpen = false; versionToDelete = null"
    />
  </Teleport>
</template>

<script setup>
import { ref, watch, onBeforeUnmount } from 'vue'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import { get, put, del } from '../composables/useApi.js'
import DeleteModal from './DeleteModal.vue'

const props = defineProps({
  chapterId: { type: String, required: true }
})

const emit = defineEmits(['close', 'restored'])

// ── State ─────────────────────────────────────────────────────────
const versions = ref([])
const selectedVersion = ref(null)
const selectedContent = ref(null)
const isLoading = ref(false)
const isLoadingContent = ref(false)
const isRestoring = ref(false)
const deletingId = ref(null)

const isDeleteModalOpen = ref(false)
const versionToDelete = ref(null)
const deleteModalTitle = ref('')
const deleteModalMessage = ref('')

// ── Read-only preview editor ───────────────────────────────────────
const previewEditor = useEditor({
  extensions: [StarterKit],
  content: '',
  editable: false,
})

// ── Load version list ──────────────────────────────────────────────
async function loadVersions() {
  isLoading.value = true
  try {
    versions.value = await get(`/api/chapters/${props.chapterId}/versions`)
  } catch (err) {
    console.error('[VersionHistory] Failed to load versions:', err)
  } finally {
    isLoading.value = false
  }
}

// ── Select version and load its content ───────────────────────────
async function selectVersion(version) {
  selectedVersion.value = version
  selectedContent.value = null
  isLoadingContent.value = true

  try {
    const full = await get(`/api/versions/${version.id}`)
    selectedContent.value = full.content
    if (previewEditor.value) {
      previewEditor.value.commands.setContent(full.content || {})
    }
  } catch (err) {
    console.error('[VersionHistory] Failed to load version content:', err)
  } finally {
    isLoadingContent.value = false
  }
}

// ── Restore selected version to active draft ─────────────────────
async function restoreVersion() {
  if (!selectedContent.value || !selectedVersion.value) return
  isRestoring.value = true

  try {
    await put(`/api/chapters/${props.chapterId}`, {
      content: selectedContent.value
    })
    emit('restored', selectedContent.value)
    emit('close')
  } catch (err) {
    console.error('[VersionHistory] Failed to restore version:', err)
  } finally {
    isRestoring.value = false
  }
}

// ── Delete a specific version ─────────────────────────────────────
function deleteVersion(version) {
  versionToDelete.value = version
  const label = version.snapshotType === 'manual_milestone' ? 'Milestone' : 'Session'
  deleteModalTitle.value = `Delete ${label}`
  deleteModalMessage.value = `Are you sure you want to delete this ${label.toLowerCase()} snapshot from ${formatDate(version.createdAt)}?\n\nThis action cannot be undone.`
  isDeleteModalOpen.value = true
}

async function executeDelete() {
  const version = versionToDelete.value
  if (!version) return
  isDeleteModalOpen.value = false

  deletingId.value = version.id
  try {
    await del(`/api/versions/${version.id}`)
    // Optimistic remove from local state
    versions.value = versions.value.filter(v => v.id !== version.id)
    // Clear the preview pane if the deleted version was selected
    if (selectedVersion.value?.id === version.id) {
      selectedVersion.value = null
      selectedContent.value = null
      if (previewEditor.value) {
        previewEditor.value.commands.setContent('')
      }
    }
  } catch (err) {
    console.error('[VersionHistory] Failed to delete version:', err)
    alert('Failed to delete revision. Please try again.')
  } finally {
    deletingId.value = null
  }
}

// ── Date helpers ──────────────────────────────────────────────────
function formatDate(isoStr) {
  return new Date(isoStr).toLocaleString('id-ID', {
    day: '2-digit', month: 'short', year: 'numeric',
    hour: '2-digit', minute: '2-digit'
  })
}

function relativeTime(isoStr) {
  const diff = Date.now() - new Date(isoStr).getTime()
  const minutes = Math.floor(diff / 60000)
  if (minutes < 1) return 'just now'
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

// Reload when chapterId changes (e.g. user switches chapters while modal is open)
watch(() => props.chapterId, () => {
  selectedVersion.value = null
  selectedContent.value = null
  versions.value = []
  loadVersions()
}, { immediate: true })

onBeforeUnmount(() => {
  if (previewEditor.value) {
    previewEditor.value.destroy()
  }
})
</script>

<style scoped>
.vh-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: fadeIn 0.15s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to   { opacity: 1; }
}

.vh-panel {
  background: #18181b;
  border: 1px solid #27272a;
  border-radius: 12px;
  width: 90vw;
  max-width: 1100px;
  height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 25px 60px rgba(0, 0, 0, 0.5);
  animation: slideUp 0.2s ease;
  overflow: hidden;
}

@keyframes slideUp {
  from { transform: translateY(20px); opacity: 0; }
  to   { transform: translateY(0);    opacity: 1; }
}

/* ── Header ─────────────────────────────────────────────────────── */
.vh-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid #27272a;
  flex-shrink: 0;
}

.vh-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-family: 'Inter', sans-serif;
  font-size: 1rem;
  font-weight: 600;
  color: #f4f4f5;
}

.vh-close {
  background: none;
  border: none;
  color: #71717a;
  font-size: 1.1rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  transition: color 0.15s, background 0.15s;
}
.vh-close:hover { color: #f4f4f5; background: #27272a; }

/* ── Body layout ─────────────────────────────────────────────────── */
.vh-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* ── Versions list pane ─────────────────────────────────────────── */
.vh-list-pane {
  width: 260px;
  flex-shrink: 0;
  border-right: 1px solid #27272a;
  overflow-y: auto;
  padding: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.vh-status {
  color: #71717a;
  font-family: 'Inter', sans-serif;
  font-size: 0.875rem;
  padding: 1rem;
  text-align: center;
}

.vh-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 2rem 1rem;
  color: #52525b;
}

.vh-hint {
  font-size: 0.75rem;
  color: #3f3f46;
  line-height: 1.4;
}

.vh-version-item {
  padding: 0.6rem 0.75rem;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.15s;
  border: 1px solid transparent;
}

.vh-version-item:hover {
  background: #27272a;
}

.vh-version-item.is-selected {
  background: #1d1d27;
  border-color: #3b82f6;
}

.vh-version-badge {
  display: inline-block;
  font-size: 0.7rem;
  font-family: 'Inter', sans-serif;
  font-weight: 600;
  padding: 0.15rem 0.5rem;
  border-radius: 20px;
  margin-bottom: 0.3rem;
}

.badge-manual {
  background: rgba(251, 191, 36, 0.15);
  color: #fbbf24;
  border: 1px solid rgba(251, 191, 36, 0.3);
}

.badge-session {
  background: rgba(99, 102, 241, 0.15);
  color: #818cf8;
  border: 1px solid rgba(99, 102, 241, 0.3);
}

.vh-version-time {
  font-size: 0.8rem;
  color: #e4e4e7;
  font-family: 'Inter', sans-serif;
  margin-top: 0.1rem;
}

.vh-version-relative {
  font-size: 0.7rem;
  color: #52525b;
  font-family: 'Inter', sans-serif;
  margin-top: 0.1rem;
}

/* ── Delete button row ──────────────────────────────────────────── */
.vh-version-item-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.25rem;
  margin-bottom: 0.15rem;
}

.vh-delete-btn {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 1px solid transparent;
  color: #52525b;
  border-radius: 4px;
  padding: 0.2rem 0.25rem;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.15s, color 0.15s, background 0.15s, border-color 0.15s;
}

/* Show delete button only on row hover */
.vh-version-item:hover .vh-delete-btn {
  opacity: 1;
}

.vh-delete-btn:hover:not(:disabled) {
  color: #f87171;
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.3);
}

.vh-delete-btn.is-deleting {
  opacity: 1;
  color: #71717a;
  cursor: wait;
}

.vh-delete-btn:disabled {
  cursor: not-allowed;
}

.vh-deleting-spinner {
  font-size: 0.75rem;
  line-height: 1;
}

/* ── Preview pane ───────────────────────────────────────────────── */
.vh-preview-pane {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.vh-preview-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.vh-preview-meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1.25rem;
  border-bottom: 1px solid #27272a;
  flex-shrink: 0;
}

.vh-preview-date {
  font-size: 0.8rem;
  color: #71717a;
  font-family: 'Inter', sans-serif;
}

.vh-tiptap-wrapper {
  flex: 1;
  overflow-y: auto;
  padding: 2rem 3rem;
}

.vh-preview-actions {
  padding: 1rem 1.25rem;
  border-top: 1px solid #27272a;
  display: flex;
  justify-content: flex-end;
  flex-shrink: 0;
}

.btn-restore {
  background: linear-gradient(135deg, #3b82f6, #6366f1);
  color: #fff;
  border: none;
  padding: 0.5rem 1.25rem;
  border-radius: 6px;
  font-size: 0.875rem;
  font-family: 'Inter', sans-serif;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.2s, transform 0.15s;
}
.btn-restore:hover:not(:disabled) {
  opacity: 0.9;
  transform: translateY(-1px);
}
.btn-restore:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>

<!-- Unscoped: override TipTap styles inside the read-only preview -->
<style>
.vh-tiptap-readonly .tiptap {
  font-family: 'Georgia', 'Lora', serif;
  font-size: 1.05rem;
  line-height: 1.8;
  color: #d4d4d8;
  outline: none;
  min-height: 100px;
}

.vh-tiptap-readonly .tiptap p {
  margin-bottom: 1.25em;
  color: #d4d4d8 !important;
}

.vh-tiptap-readonly .tiptap h1,
.vh-tiptap-readonly .tiptap h2,
.vh-tiptap-readonly .tiptap h3 {
  color: #f4f4f5;
  margin-bottom: 0.75em;
  font-family: 'Inter', sans-serif;
}
</style>
