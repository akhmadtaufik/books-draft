import { ref, watch, onUnmounted } from 'vue'
import { put } from './useApi.js'

/**
 * Autosave composable for TipTap editor content.
 *
 * @param {import('vue').Ref<string|number>} chapterId  – reactive chapter ID
 * @param {() => object} getContent – returns editor JSON content
 */
export function useAutosave(chapterId, getContent) {
  const isSaving = ref(false)
  const lastSavedAt = ref(null)
  const hasUnsavedChanges = ref(false)

  let debounceTimer = null
  const DEBOUNCE_MS = 2000

  // ── localStorage helpers ─────────────────────────────────────────
  function draftKey(id) {
    return `draft_chapter_${id}`
  }

  function saveDraftLocal(id, content) {
    try {
      localStorage.setItem(
        draftKey(id),
        JSON.stringify({ content, savedAt: Date.now() }),
      )
    } catch {
      // quota exceeded — best-effort
    }
  }

  function loadDraftLocal(id) {
    try {
      const raw = localStorage.getItem(draftKey(id))
      return raw ? JSON.parse(raw) : null
    } catch {
      return null
    }
  }

  function clearDraftLocal(id) {
    localStorage.removeItem(draftKey(id))
  }

  // ── Core save logic ──────────────────────────────────────────────
  async function performSave() {
    const id = chapterId.value
    if (!id) return

    const content = getContent()
    if (!content) return

    // 1. Persist to localStorage immediately
    saveDraftLocal(id, content)
    hasUnsavedChanges.value = true
    isSaving.value = true

    try {
      await put(`/api/chapters/${id}`, { content })
      clearDraftLocal(id)
      lastSavedAt.value = new Date()
      hasUnsavedChanges.value = false
    } catch (err) {
      console.warn('[autosave] API save failed, draft kept in localStorage', err)
      // Draft stays in localStorage for recovery
    } finally {
      isSaving.value = false
    }
  }

  // ── Public trigger (call from editor 'update' handler) ───────────
  function triggerSave() {
    hasUnsavedChanges.value = true
    clearTimeout(debounceTimer)
    debounceTimer = setTimeout(performSave, DEBOUNCE_MS)
  }

  // ── Draft recovery ───────────────────────────────────────────────
  /**
   * Returns locally-saved content if it's newer than `serverSavedAt` (ms epoch).
   * Otherwise returns null.
   */
  function recoverDraft(serverSavedAt = 0) {
    const id = chapterId.value
    if (!id) return null

    const draft = loadDraftLocal(id)
    if (!draft) return null

    if (draft.savedAt > serverSavedAt) {
      return draft.content
    }
    // Server has newer data — discard stale draft
    clearDraftLocal(id)
    return null
  }

  // Cleanup
  onUnmounted(() => {
    clearTimeout(debounceTimer)
  })

  return {
    isSaving,
    lastSavedAt,
    hasUnsavedChanges,
    triggerSave,
    recoverDraft,
  }
}
