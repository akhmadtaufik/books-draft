import nspell from 'nspell'

let spell = null

/**
 * Initialise the nspell instance by fetching the Indonesian dictionary files
 * from the public directory.
 */
async function init(customWords = []) {
  try {
    const [affRes, dicRes] = await Promise.all([
      fetch('/dictionaries/id_ID.aff'),
      fetch('/dictionaries/id_ID.dic'),
    ])

    if (!affRes.ok || !dicRes.ok) {
      throw new Error('Failed to fetch dictionary files')
    }

    const [aff, dic] = await Promise.all([affRes.text(), dicRes.text()])

    spell = nspell(aff, dic)
    
    // Add any custom words
    for (const word of customWords) {
      if (word) spell.add(word)
    }

    self.postMessage({ type: 'ready' })
  } catch (err) {
    self.postMessage({ type: 'error', message: err.message })
  }
}

/**
 * Tokenise text and check each unique word.
 * Returns misspelled words with all their character-offset indices.
 */
function checkText(text, id) {
  if (!spell) {
    self.postMessage({ type: 'error', message: 'Spell checker not initialised' })
    return
  }

  const wordRegex = /[\p{L}\p{M}]+/gu
  const seen = new Map() // word (lowercase) → { word, indices }
  let match

  while ((match = wordRegex.exec(text)) !== null) {
    const word = match[0]
    const key = word.toLowerCase()

    if (!seen.has(key)) {
      seen.set(key, { word, indices: [match.index] })
    } else {
      seen.get(key).indices.push(match.index)
    }
  }

  const misspelled = []

  for (const [, entry] of seen) {
    if (!spell.correct(entry.word)) {
      misspelled.push({ word: entry.word, indices: entry.indices })
    }
  }

  self.postMessage({ type: 'results', misspelled, id })
}

/**
 * Handle incoming messages from the main thread.
 */
self.addEventListener('message', (e) => {
  const { type, text, id, word, customWords } = e.data

  switch (type) {
    case 'init':
      init(customWords)
      break
    case 'check':
      checkText(text, id)
      break
    case 'add-word':
      if (spell && word) spell.add(word)
      break
    default:
      self.postMessage({ type: 'error', message: `Unknown message type: ${type}` })
  }
})
