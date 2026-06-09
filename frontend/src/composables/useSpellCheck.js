import { ref, onUnmounted } from 'vue'
import { get } from './useApi.js'

/**
 * Composable for spell checking via a Web Worker.
 * Extracts text from ProseMirror docs and maps misspelled words
 * back to document positions.
 */
export function useSpellCheck() {
  const isReady = ref(false)
  const misspelledRanges = ref([]) // Array of { from, to, word }
  const pendingSuggestions = new Map()

  let worker = null
  let debounceTimer = null
  let checkId = 0
  const DEBOUNCE_MS = 500

  // ── Create worker ────────────────────────────────────────────────
  
  async function initWorker() {
    try {
      worker = new Worker(
        new URL('../workers/spellWorker.js', import.meta.url),
        { type: 'module' },
      )

      // Fetch custom dictionary words
      let customWords = []
      try {
        const dict = await get('/api/dictionary')
        if (dict && dict.words) {
          customWords = dict.words
        }
      } catch (err) {
        console.warn('[spellCheck] Failed to fetch custom dictionary:', err)
      }

      worker.postMessage({ type: 'init', customWords })

      worker.addEventListener('message', (e) => {
        // 1. Handle Suggestion Results
        if (e.data.action === 'suggest_result') {
          const resolve = pendingSuggestions.get(e.data.id)
          if (resolve) {
            resolve(e.data.suggestions)
            pendingSuggestions.delete(e.data.id)
          }
          return
        }

        // 2. Handle Initialization Ready State (CRUCIAL FOR RED LINES)
        if (e.data.type === 'ready') {
          isReady.value = true
          // Optional: re-check the document immediately once ready
          console.log('[spellCheck] Worker is ready!')
        } 
        
        // 3. Handle Initialization Errors
        else if (e.data.type === 'error') {
          console.error('[spellCheck worker error]:', e.data.message)
        }
        
        // 4. Handle Check Results
        else if (e.data.type === 'results') {
          // Only process the latest check
          if (e.data.id === checkId) {
            misspelledRanges.value = e.data.misspelled || []
          }
        }
      })
    } catch (err) {
      console.warn('[spellCheck] Worker creation failed:', err)
    }
  }
  
  initWorker()

  // ── Extract text from ProseMirror doc with position info ─────────
  function extractTextWithPositions(doc) {
    const segments = [] // { text, offset } — offset is the doc position
    let fullText = ''

    doc.descendants((node, pos) => {
      if (node.isText) {
        segments.push({ text: node.text, offset: pos, start: fullText.length })
        fullText += node.text
      } else if (node.isBlock && fullText.length > 0) {
        // Add a space between blocks so words don't merge
        fullText += ' '
      }
      return true // continue traversal
    })

    return { fullText, segments }
  }

  /**
   * Map character index in fullText back to a ProseMirror doc position.
   */
  function charIndexToDocPos(charIndex, segments) {
    for (const seg of segments) {
      const segEnd = seg.start + seg.text.length
      if (charIndex >= seg.start && charIndex < segEnd) {
        return seg.offset + (charIndex - seg.start)
      }
    }
    return -1
  }

  // ── Public API ───────────────────────────────────────────────────
  function checkDocument(doc) {
    if (!worker || !isReady.value) return

    clearTimeout(debounceTimer)
    debounceTimer = setTimeout(() => {
      checkId++
      const { fullText, segments } = extractTextWithPositions(doc)

      // Store segments so we can map results back
      worker._segments = segments

      worker.addEventListener(
        'message',
        function handler(e) {
          if (e.data.type === 'results' && e.data.id === checkId) {
            worker.removeEventListener('message', handler)

            const ranges = []
            for (const entry of e.data.misspelled) {
              for (const charIdx of entry.indices) {
                const from = charIndexToDocPos(charIdx, segments)
                if (from === -1) continue
                const to = from + entry.word.length
                ranges.push({ from, to, word: entry.word })
              }
            }
            misspelledRanges.value = ranges
          }
        },
      )

      worker.postMessage({ type: 'check', text: fullText, id: checkId })
    }, DEBOUNCE_MS)
  }

  function addWord(word) {
    if (worker) {
      worker.postMessage({ type: 'add-word', word })
    }
  }

  function terminate() {
    clearTimeout(debounceTimer)
    if (worker) {
      worker.terminate()
      worker = null
    }
  }

  onUnmounted(terminate)

  function getSuggestions(word) {
    return new Promise(resolve => {
      const id = Date.now().toString() + Math.random()
      pendingSuggestions.set(id, resolve)
      worker.postMessage({ action: 'suggest', id, word })
    })
  }

  return {
    isReady,
    misspelledRanges,
    checkDocument,
    addWord,
    terminate,
    getSuggestions,
  }
}
