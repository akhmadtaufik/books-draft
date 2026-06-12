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
      <div class="editor-scroll-area" @click="onEditorClick" @contextmenu="handleContextMenu">
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



      <div v-if="commentTooltip.show" class="comment-tooltip" :style="{ top: commentTooltip.y + 'px', left: commentTooltip.x + 'px' }">
        <div class="comment-header">📝 Revision Note</div>
        <div class="comment-body">{{ commentTooltip.text }}</div>
      </div>

      <div v-if="contextMenu.show" class="custom-context-menu" :style="{ top: contextMenu.y + 'px', left: contextMenu.x + 'px' }" @click.stop>
        <template v-if="contextMenu.mode === 'menu'">
          <div class="context-menu-section">
            <span class="context-menu-label">Highlight:</span>
            <div class="preset-colors">
              <button v-for="color in presetColors" :key="color"
                      class="color-btn" :style="{ backgroundColor: color }"
                      @click="applyHighlight(color)"></button>
              <button class="clear-btn" @click="clearHighlight" title="Clear Highlight">🧹</button>
            </div>
          </div>
          
          <div class="context-menu-divider"></div>
          
          <button v-if="editor.isActive('comment')" class="context-menu-btn text-red" @click="removeComment">
            🗑️ Remove Comment
          </button>
          <button v-else class="context-menu-btn" @click="openCommentInput">
            💬 Add Comment
          </button>

          <div class="context-menu-divider"></div>
          
          <button class="context-menu-btn" @click="handleCopy">
            📋 Copy
          </button>
          <button class="context-menu-btn" @click="handlePaste">
            📄 Paste (with formatting)
          </button>
          <button class="context-menu-btn" @click="handlePastePlain">
            📝 Paste (without formatting)
          </button>
        </template>

        <template v-else-if="contextMenu.mode === 'comment-input'">
          <div class="context-input-wrapper">
            <input 
              v-model="contextMenu.commentText"
              type="text"
              placeholder="Type revision note..."
              class="context-input"
              @keyup.enter="saveContextComment"
              ref="commentInputRef"
            />
            <button class="btn-save-comment" @click="saveContextComment">Save</button>
          </div>
        </template>
        <template v-else-if="contextMenu.mode === 'spelling'">
          <div class="context-menu-section">
            <span class="context-menu-label">Suggestions:</span>
            <div v-if="contextMenu.suggestions.length === 0" class="context-menu-label" style="padding: 0.5rem; font-style: italic;">
              Loading... or no suggestions
            </div>
            <template v-else>
              <button 
                v-for="sugg in contextMenu.suggestions" 
                :key="sugg"
                class="context-menu-btn"
                @click="applySuggestion(sugg)"
              >
                ✨ {{ sugg }}
              </button>
            </template>
            
            <div class="context-menu-divider"></div>
            
            <button class="context-menu-btn text-blue" @click="handleAddWordContext">
              📚 Add "{{ contextMenu.word }}" to Dictionary
            </button>
          </div>
        </template>
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
import { ref, watch, onBeforeUnmount, toRef, nextTick, onMounted } from 'vue'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import CharacterCount from '@tiptap/extension-character-count'
import Highlight from '@tiptap/extension-highlight'
import { CommentMark } from '../extensions/commentExtension.js'

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

const commentTooltip = ref({ show: false, x: 0, y: 0, text: '' })

// --- Custom Context Menu State ---
const contextMenu = ref({
  show: false,
  x: 0,
  y: 0,
  mode: 'menu', // 'menu' | 'comment-input' | 'spelling'
  commentText: '',
  word: '',
  pos: null,
  suggestions: []
})

// Deep Muted colors for perfect Dark Mode contrast
const presetColors = ['#a16207', '#15803d', '#be185d', '#1d4ed8']
const commentInputRef = ref(null)

// --- Context Menu Handlers ---
function handleContextMenu(event) {
  if (!editor.value) return

  const { empty } = editor.value.state.selection
  const isCommentNode = event.target.closest('.editor-comment')
  const isSpellingError = event.target.classList.contains('spelling-error')

  if (!empty || isCommentNode || isSpellingError) {
    event.preventDefault() 

    // Calculate precise ProseMirror position for word replacement
    let pos = null
    if (isSpellingError && event.target.firstChild) {
      pos = editor.value.view.posAtDOM(event.target.firstChild, 0)
    }

    const word = isSpellingError ? event.target.dataset.word : ''

    contextMenu.value = {
      show: true,
      x: event.clientX,
      y: event.clientY,
      mode: isSpellingError ? 'spelling' : 'menu',
      commentText: '',
      word: word,
      pos: pos,
      suggestions: []
    }

    if (isSpellingError) {
      getSuggestions(word).then(suggs => {
        if (contextMenu.value.show && contextMenu.value.mode === 'spelling') {
          contextMenu.value.suggestions = suggs.slice(0, 5) // Top 5 suggestions
        }
      })
    }
  } else {
    contextMenu.value.show = false
  }
}

// Close context menu when clicking anywhere else
function closeContextMenu(e) {
  if (!e.target.closest('.custom-context-menu')) {
    contextMenu.value.show = false
  }
}

onMounted(() => document.addEventListener('click', closeContextMenu))
onBeforeUnmount(() => document.removeEventListener('click', closeContextMenu))

// --- Action Functions ---
async function handleCopy() {
  try {
    // Focus the editor to ensure the selection is active, then execute native copy
    editor.value.chain().focus().run()
    document.execCommand('copy')
  } catch (err) {
    console.error('Copy failed:', err)
  }
  contextMenu.value.show = false
}

async function handlePaste() {
  try {
    editor.value.chain().focus().run()
    
    // Attempt to read rich text via modern Clipboard API
    const clipboardItems = await navigator.clipboard.read()
    let htmlContent = null
    let textContent = null

    for (const item of clipboardItems) {
      if (item.types.includes('text/html')) {
        const blob = await item.getType('text/html')
        htmlContent = await blob.text()
      }
      if (item.types.includes('text/plain')) {
        const blob = await item.getType('text/plain')
        textContent = await blob.text()
      }
    }

    if (htmlContent) {
      editor.value.commands.insertContent(htmlContent)
    } else if (textContent) {
      editor.value.commands.insertContent(textContent)
    }
  } catch (err) {
    console.warn('Rich text paste failed or requires permission. Falling back to plain text...', err)
    try {
      const text = await navigator.clipboard.readText()
      editor.value.commands.insertContent(text)
    } catch (fallbackErr) {
      alert('Browser blocked clipboard access. Please use Ctrl+V to paste.')
    }
  }
  contextMenu.value.show = false
}

async function handlePastePlain() {
  try {
    editor.value.chain().focus().run()
    // Specifically request plain text
    const text = await navigator.clipboard.readText()
    editor.value.commands.insertContent(text)
  } catch (err) {
    console.error('Plain text paste failed:', err)
    alert('Browser blocked clipboard access. Please use Ctrl+Shift+V to paste plain text.')
  }
  contextMenu.value.show = false
}

function applyHighlight(color) {
  editor.value.chain().focus().setHighlight({ color }).run()
  contextMenu.value.show = false
}

function clearHighlight() {
  editor.value.chain().focus().unsetHighlight().run()
  contextMenu.value.show = false
}

function openCommentInput() {
  contextMenu.value.mode = 'comment-input'
  // Focus the input field after Vue renders it
  nextTick(() => {
    if (commentInputRef.value) commentInputRef.value.focus()
  })
}

function saveContextComment() {
  if (contextMenu.value.commentText.trim()) {
    editor.value.chain().focus().setMark('comment', { text: contextMenu.value.commentText }).run()
  }
  contextMenu.value.show = false
}

function removeComment() {
  editor.value.chain().focus().unsetMark('comment').run()
  contextMenu.value.show = false
}

function applySuggestion(suggestion) {
  if (contextMenu.value.pos !== null) {
    editor.value.chain()
      .focus()
      .deleteRange({ from: contextMenu.value.pos, to: contextMenu.value.pos + contextMenu.value.word.length })
      .insertContentAt(contextMenu.value.pos, suggestion)
      .run()
  }
  contextMenu.value.show = false
}

async function handleAddWordContext() {
  const word = contextMenu.value.word
  if (!word) return
  contextMenu.value.show = false

  try {
    await post('/api/dictionary', { word })
    addWord(word)
    if (editor.value) checkDocument(editor.value.state.doc)
  } catch (err) {
    console.error('Failed to add word:', err)
  }
}

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
    Highlight.configure({ multicolor: true }), // Enable custom colors
    CommentMark, // Add our custom extension
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
const { 
  isReady: spellCheckReady, 
  misspelledRanges, 
  checkDocument, 
  addWord, 
  terminate: terminateSpellCheck,
  getSuggestions // CRUCIAL FIX: Added missing getSuggestions
} = useSpellCheck()

// CRUCIAL FIX: Watch for the exact moment the dictionary worker is ready, then trigger the initial spellcheck
watch(spellCheckReady, (isReady) => {
  if (isReady && editor.value) {
    console.log('[TipTapEditor] Spellchecker is ready. Running initial check...')
    checkDocument(editor.value.state.doc)
  }
})

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
  // Handle Inline Comments ONLY
  const commentNode = event.target.closest('.editor-comment')
  if (commentNode) {
    const commentText = commentNode.getAttribute('data-comment')
    if (commentText) {
      commentTooltip.value = { show: true, x: event.clientX, y: event.clientY + 20, text: commentText }
      return
    }
  }
  commentTooltip.value.show = false
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

/* Global TipTap styles (unscoped) */
/* Refined Highlight Styling for Dark Mode */
.tiptap mark {
  color: #f4f4f5 !important; /* Strictly ensure the text remains bright white */
  border-radius: 4px;
  padding: 0.15em 0.25em; /* Add breathing room around the highlighted text */
  margin: 0 -0.1em;
  box-decoration-break: clone;
  -webkit-box-decoration-break: clone;
}

.tiptap .editor-comment {
  background-color: rgba(99, 102, 241, 0.2);
  border-bottom: 2px solid #6366f1;
  cursor: pointer;
  transition: background-color 0.2s;
}
.tiptap .editor-comment:hover {
  background-color: rgba(99, 102, 241, 0.4);
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


/* Scoped styles */
.comment-tooltip {
  position: fixed;
  background-color: #18181b;
  border: 1px solid #4f46e5;
  border-radius: 8px;
  padding: 0.75rem;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.5);
  max-width: 250px;
  z-index: 50;
  color: #e4e4e7;
  font-family: 'Inter', sans-serif;
}
.comment-header { font-size: 0.75rem; color: #a1a1aa; margin-bottom: 0.5rem; text-transform: uppercase; letter-spacing: 0.5px; font-weight: 600; }
.comment-body { font-size: 0.875rem; line-height: 1.5; }

/* Custom Context Menu */
.custom-context-menu {
  position: fixed;
  background-color: #18181b;
  border: 1px solid #3f3f46;
  border-radius: 8px;
  padding: 0.5rem;
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.5);
  z-index: 100;
  min-width: 180px;
  font-family: 'Inter', sans-serif;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.context-menu-section {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  padding: 0.25rem;
}

.context-menu-label {
  font-size: 0.75rem;
  color: #a1a1aa;
  font-weight: 500;
}

.preset-colors {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.color-btn {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  border: 1px solid rgba(255, 255, 255, 0.1);
  cursor: pointer;
  transition: transform 0.1s;
}

.color-btn:hover {
  transform: scale(1.15);
  border-color: rgba(255, 255, 255, 0.5);
}

.clear-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  font-size: 1rem;
  padding: 0;
  margin-left: auto;
}

.context-menu-divider {
  height: 1px;
  background-color: #27272a;
  margin: 0.25rem 0;
}

.context-menu-btn {
  background: transparent;
  border: none;
  color: #e4e4e7;
  text-align: left;
  padding: 0.5rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background-color 0.2s;
}

.context-menu-btn:hover {
  background-color: #27272a;
}

.text-red {
  color: #f87171 !important;
}

.text-blue {
  color: #60a5fa !important;
}

.context-input-wrapper {
  display: flex;
  gap: 0.5rem;
  padding: 0.25rem;
}

.context-input {
  background: #0f0f12;
  border: 1px solid #3f3f46;
  color: #e4e4e7;
  padding: 0.4rem;
  border-radius: 4px;
  font-size: 0.875rem;
  width: 150px;
}

.context-input:focus {
  outline: none;
  border-color: #8b5cf6;
}

.btn-save-comment {
  background: #f4f4f5;
  color: #18181b;
  border: none;
  border-radius: 4px;
  padding: 0 0.5rem;
  cursor: pointer;
  font-weight: 500;
  font-size: 0.8rem;
}
</style>
