<template>
  <div class="editor-container">
    <div v-if="isLoading" class="loading-state">Loading chapter...</div>
    <div v-else-if="error" class="error-state">{{ error }}</div>
    <div v-else class="editor-wrapper">
      
      <!-- Toolbar -->
      <EditorToolbar 
        v-if="editor" 
        :editor="editor" 
        :is-saving-milestone="isSavingMilestone" 
        @save-milestone="saveManualMilestone" 
        @show-history="showVersionHistory = true" 
      />

      <!-- Editor Content -->
      <div class="editor-scroll-area" @click="onEditorClick">
        <textarea 
          v-model="chapterTitle" 
          @input="handleTitleInput"
          class="chapter-title-input" 
          placeholder="Chapter Title"
          rows="1"
          ref="titleTextarea"
        ></textarea>
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
import { ref, watch, onBeforeUnmount, toRef, nextTick } from 'vue'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import CharacterCount from '@tiptap/extension-character-count'

import { get, put, post } from '../composables/useApi.js'
import { chapterApi } from '../api/chapterApi.js'
import { useAutosave } from '../composables/useAutosave.js'
import { useSpellCheck } from '../composables/useSpellCheck.js'
import { calculateReadingTime } from '../composables/useReadingTime.js'
import { SpellCheckExtension, applySpellDecorations } from '../extensions/spellCheckExtension.js'
import VersionHistory from './VersionHistory.vue'
import EditorToolbar from './EditorToolbar.vue'

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

const titleTextarea = ref(null)

function handleTitleInput(event) {
  const el = event.target
  el.style.height = 'auto'
  el.style.height = el.scrollHeight + 'px'
  onTitleUpdate()
}

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
  chapterApi.update(props.chapterId, { title: chapterTitle.value })
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
    const chapter = await chapterApi.getById(props.chapterId)
    chapterTitle.value = chapter.title || ''
    
    nextTick(() => {
      if (titleTextarea.value) {
        titleTextarea.value.style.height = 'auto'
        titleTextarea.value.style.height = titleTextarea.value.scrollHeight + 'px'
      }
    })
    
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
    await chapterApi.saveSnapshot(chapterId, 'session_end')
    isDirty.value = false // Reset after successful save
  } catch (err) {
    console.warn('[VersionHistory] Session snapshot failed:', err)
  }
}

async function saveManualMilestone() {
  if (!props.chapterId || isSavingMilestone.value) return
  isSavingMilestone.value = true
  try {
    await chapterApi.saveSnapshot(props.chapterId, 'manual_milestone')
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
  white-space: pre-wrap;
  word-break: break-word;
  overflow: hidden;
  resize: none;
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
