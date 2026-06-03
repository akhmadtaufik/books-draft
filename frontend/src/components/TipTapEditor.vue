<template>
  <div class="editor-container">
    <div v-if="isLoading" class="loading-state">Loading chapter...</div>
    <div v-else-if="error" class="error-state">{{ error }}</div>
    <div v-else class="editor-wrapper">
      
      <!-- Toolbar -->
      <div class="toolbar" v-if="editor">
        <button @click="editor.chain().focus().toggleBold().run()" :class="{ 'is-active': editor.isActive('bold') }">
          <b>B</b>
        </button>
        <button @click="editor.chain().focus().toggleItalic().run()" :class="{ 'is-active': editor.isActive('italic') }">
          <i>I</i>
        </button>
        <button @click="editor.chain().focus().toggleStrike().run()" :class="{ 'is-active': editor.isActive('strike') }">
          <s>S</s>
        </button>
        <div class="divider"></div>
        <button @click="editor.chain().focus().toggleHeading({ level: 1 }).run()" :class="{ 'is-active': editor.isActive('heading', { level: 1 }) }">
          H1
        </button>
        <button @click="editor.chain().focus().toggleHeading({ level: 2 }).run()" :class="{ 'is-active': editor.isActive('heading', { level: 2 }) }">
          H2
        </button>
        <button @click="editor.chain().focus().toggleHeading({ level: 3 }).run()" :class="{ 'is-active': editor.isActive('heading', { level: 3 }) }">
          H3
        </button>
        <div class="divider"></div>
        <button @click="editor.chain().focus().toggleBulletList().run()" :class="{ 'is-active': editor.isActive('bulletList') }">
          • List
        </button>
        <button @click="editor.chain().focus().toggleOrderedList().run()" :class="{ 'is-active': editor.isActive('orderedList') }">
          1. List
        </button>
        <button @click="editor.chain().focus().toggleBlockquote().run()" :class="{ 'is-active': editor.isActive('blockquote') }">
          " Quote
        </button>
        <div class="divider"></div>
        <button @click="editor.chain().focus().undo().run()" :disabled="!editor.can().undo()">
          Undo
        </button>
        <button @click="editor.chain().focus().redo().run()" :disabled="!editor.can().redo()">
          Redo
        </button>
        <div class="divider"></div>
        <button @click="saveManualMilestone" :disabled="isSavingMilestone" class="btn-milestone">
          {{ isSavingMilestone ? '...' : '⭐ Save Milestone' }}
        </button>
        <button @click="showVersionHistory = true" class="btn-history">
          ⏱ History
        </button>
      </div>

      <!-- Editor Content -->
      <div class="editor-scroll-area" @click="onEditorClick">
        <input 
          v-model="chapterTitle" 
          @input="onTitleUpdate"
          class="chapter-title-input" 
          placeholder="Chapter Title"
        />
        <editor-content :editor="editor" class="tiptap-content" spellcheck="false" />
      </div>

      <!-- Spell Check Tooltip -->
      <div v-if="tooltip.show" class="spell-tooltip" :style="{ top: tooltip.y + 'px', left: tooltip.x + 'px' }">
        <span class="misspelled-word">"{{ tooltip.word }}"</span>
        <button @click="handleAddWord" class="btn-add-word">Add to Dictionary</button>
      </div>

      <!-- Status Bar -->
      <div class="status-bar">
        <div class="status-left">
          <span v-if="isSaving" class="status-saving">Saving...</span>
          <span v-else-if="hasUnsavedChanges" class="status-unsaved">Unsaved changes</span>
          <span v-else-if="lastSavedAt" class="status-saved">Saved ✓</span>
        </div>
        <div class="status-right" v-if="editor">
          <span>Words: {{ editor.storage.characterCount?.words() || 0 }} | ⏱️ Est. reading time: {{ estimatedReadingTime }}</span>
          <span>{{ editor.storage.characterCount?.characters() || 0 }} chars</span>
        </div>
      </div>
    </div>
  </div>

  <!-- Revision History Modal -->
  <VersionHistory
    v-if="showVersionHistory"
    :chapterId="props.chapterId"
    @close="showVersionHistory = false"
    @restored="onVersionRestored"
  />
</template>

<script setup>
import { ref, watch, onBeforeUnmount, toRef } from 'vue'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import CharacterCount from '@tiptap/extension-character-count'

import { get, put, post } from '../composables/useApi.js'
import { useAutosave } from '../composables/useAutosave.js'
import { useSpellCheck } from '../composables/useSpellCheck.js'
import { calculateReadingTime } from '../composables/useReadingTime.js'
import { SpellCheckExtension, applySpellDecorations } from '../extensions/spellCheckExtension.js'
import VersionHistory from './VersionHistory.vue'

const props = defineProps({
  chapterId: { type: String, required: true }
})
const emit = defineEmits(['title-updated'])

const isLoading = ref(false)
const error = ref(null)
const chapterTitle = ref('')
const isDirty = ref(false)

const tooltip = ref({ show: false, x: 0, y: 0, word: '' })
const showVersionHistory = ref(false)
const isSavingMilestone = ref(false)
const estimatedReadingTime = ref('1 min')

const editor = useEditor({
  extensions: [
    StarterKit,
    CharacterCount,
    SpellCheckExtension,
  ],
  content: '',
  onUpdate: ({ editor }) => {
    onEditorUpdate()
  }
})

const chapterIdRef = toRef(props, 'chapterId')

// Autosave integration
const { isSaving, lastSavedAt, hasUnsavedChanges, triggerSave, recoverDraft } = useAutosave(
  chapterIdRef,
  () => editor.value?.getJSON()
)

// Spell check integration
const { isReady: spellCheckReady, misspelledRanges, checkDocument, addWord, terminate: terminateSpellCheck } = useSpellCheck()

// Watch for spell check results and apply decorations
watch(misspelledRanges, (ranges) => {
  if (editor.value) {
    applySpellDecorations(editor.value, ranges)
  }
})

// Combined update handler
function onEditorUpdate() {
  isDirty.value = true
  triggerSave()
  if (editor.value) {
    checkDocument(editor.value.state.doc)
    // Update reading time efficiently without blocking the event loop
    requestAnimationFrame(() => {
      const stats = calculateReadingTime(editor.value.getJSON())
      estimatedReadingTime.value = stats.formattedTime
    })
  }
}

function onTitleUpdate() {
  isDirty.value = true
  hasUnsavedChanges.value = true
  // Emit to parent immediately for sidebar reactivity
  emit('title-updated', { id: props.chapterId, title: chapterTitle.value })
  
  // Trigger title save immediately or debounce
  put(`/api/chapters/${props.chapterId}`, { title: chapterTitle.value })
    .then(() => {
      lastSavedAt.value = new Date()
      hasUnsavedChanges.value = false
    })
    .catch(err => console.error('Failed to save title:', err))
}

function onEditorClick(event) {
  if (event.target && event.target.classList.contains('spelling-error')) {
    const word = event.target.dataset.word
    if (word) {
      tooltip.value = {
        show: true,
        x: event.clientX,
        y: event.clientY + 20,
        word
      }
      return
    }
  }
  tooltip.value.show = false
}

async function handleAddWord() {
  const word = tooltip.value.word
  if (!word) return

  tooltip.value.show = false

  try {
    await post('/api/dictionary', { word })
    addWord(word)
    // Re-check document to clear underline
    if (editor.value) {
      checkDocument(editor.value.state.doc)
    }
  } catch (err) {
    console.error('Failed to add word to dictionary:', err)
  }
}

async function loadChapter() {
  if (!props.chapterId) return
  
  isLoading.value = true
  error.value = null
  
  try {
    const chapter = await get(`/api/chapters/${props.chapterId}`)
    chapterTitle.value = chapter.title || ''
    
    // Check for draft recovery
    const draft = recoverDraft(new Date(chapter.updatedAt).getTime())
    const initialContent = draft || chapter.content || {}
    
    if (editor.value) {
      editor.value.commands.setContent(initialContent)
      // Reset isDirty since we just loaded it
      isDirty.value = false
      // Initial spell check and reading time
      checkDocument(editor.value.state.doc)
      const stats = calculateReadingTime(editor.value.getJSON())
      estimatedReadingTime.value = stats.formattedTime
    }
  } catch (err) {
    error.value = err.message
  } finally {
    isLoading.value = false
  }
}

watch(() => props.chapterId, (newId, oldId) => {
  // Save session snapshot for the chapter we're leaving
  if (oldId) {
    saveSessionSnapshot(oldId)
  }
  loadChapter()
})

// Initial load if component mounted with a chapterId
if (props.chapterId) {
  loadChapter()
}

async function saveSessionSnapshot(chapterId) {
  if (!isDirty.value) {
    console.log(`[VersionHistory] No changes made in this session (Chapter ${chapterId}). Skipping snapshot generation.`)
    return
  }

  try {
    await post(`/api/chapters/${chapterId}/versions`, { snapshot_type: 'session_end' })
    isDirty.value = false // Reset after successful save
  } catch (err) {
    console.warn('[VersionHistory] Session snapshot failed:', err)
  }
}

async function saveManualMilestone() {
  if (!props.chapterId || isSavingMilestone.value) return
  isSavingMilestone.value = true
  try {
    await post(`/api/chapters/${props.chapterId}/versions`, { snapshot_type: 'manual_milestone' })
    isDirty.value = false // Reset since current state is safely backed up
  } catch (err) {
    console.error('[VersionHistory] Milestone save failed:', err)
  } finally {
    isSavingMilestone.value = false
  }
}

async function onVersionRestored(restoredContent) {
  if (editor.value) {
    editor.value.commands.setContent(restoredContent)
    triggerSave()
  }
}

onBeforeUnmount(() => {
  // Fire session snapshot on unmount (user navigated away)
  if (props.chapterId) {
    saveSessionSnapshot(props.chapterId)
  }
  if (editor.value) {
    editor.value.destroy()
  }
})
</script>

<style>
/* Unscoped styles for TipTap content */
.tiptap {
  min-height: 50vh;
  outline: none;
  font-family: 'Georgia', 'Lora', serif;
  font-size: 1.125rem;
  line-height: 1.8;
  color: #d4d4d8;
}

.tiptap p {
  margin-bottom: 1.5em;
}

.tiptap h1, .tiptap h2, .tiptap h3 {
  color: #f4f4f5;
  margin-top: 2em;
  margin-bottom: 1em;
  font-family: 'Inter', sans-serif;
}

.tiptap blockquote {
  border-left: 3px solid #52525b;
  padding-left: 1rem;
  margin-left: 0;
  color: #a1a1aa;
  font-style: italic;
}

.spelling-error {
  text-decoration: wavy underline #ef4444;
  text-underline-offset: 3px;
}
</style>

<style scoped>
.editor-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #0f0f12;
}

.editor-wrapper {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.5rem 1rem;
  background-color: #18181b;
  border-bottom: 1px solid #27272a;
  flex-wrap: wrap;
}

.toolbar button {
  background: transparent;
  border: 1px solid transparent;
  color: #a1a1aa;
  padding: 0.4rem 0.6rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  font-family: 'Inter', sans-serif;
  transition: all 0.2s;
}

.toolbar button:hover:not(:disabled) {
  background-color: #27272a;
  color: #e4e4e7;
}

.toolbar button.is-active {
  background-color: #27272a;
  color: #fff;
  border-color: #3f3f46;
}

.toolbar button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.divider {
  width: 1px;
  height: 20px;
  background-color: #3f3f46;
  margin: 0 0.5rem;
}

.btn-milestone {
  background: rgba(251, 191, 36, 0.1) !important;
  color: #fbbf24 !important;
  border-color: rgba(251, 191, 36, 0.3) !important;
}
.btn-milestone:hover:not(:disabled) {
  background: rgba(251, 191, 36, 0.2) !important;
}

.btn-history {
  background: rgba(99, 102, 241, 0.1) !important;
  color: #818cf8 !important;
  border-color: rgba(99, 102, 241, 0.3) !important;
}
.btn-history:hover {
  background: rgba(99, 102, 241, 0.2) !important;
}

.editor-scroll-area {
  flex: 1;
  overflow-y: auto;
  padding: 3rem 15%;
}

.chapter-title-input {
  width: 100%;
  background: transparent;
  border: none;
  border-bottom: 1px solid transparent;
  color: #fff;
  font-size: 2.5rem;
  font-weight: 700;
  font-family: 'Inter', sans-serif;
  padding: 0.5rem 0;
  margin-bottom: 2rem;
  outline: none;
  transition: border-color 0.2s;
}

.chapter-title-input:focus {
  border-bottom-color: #3f3f46;
}

.status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 1rem;
  background-color: #18181b;
  border-top: 1px solid #27272a;
  font-size: 0.75rem;
  color: #a1a1aa;
  font-family: 'Inter', sans-serif;
}

.status-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.status-saving {
  color: #fbbf24;
}

.status-unsaved {
  color: #f87171;
}

.status-saved {
  color: #34d399;
}

.status-right {
  display: flex;
  gap: 1rem;
}

.loading-state, .error-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #a1a1aa;
}

.error-state {
  color: #ef4444;
}

.spell-tooltip {
  position: fixed;
  background-color: #27272a;
  border: 1px solid #3f3f46;
  border-radius: 6px;
  padding: 0.5rem;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  z-index: 50;
}

.misspelled-word {
  color: #f87171;
  font-family: 'Inter', sans-serif;
  font-size: 0.875rem;
  font-weight: 500;
  text-align: center;
}

.btn-add-word {
  background-color: #3f3f46;
  color: #e4e4e7;
  border: none;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.75rem;
  font-family: 'Inter', sans-serif;
  transition: background-color 0.2s;
}

.btn-add-word:hover {
  background-color: #52525b;
}
</style>
