/**
 * Extracts raw text from a TipTap JSON document structure.
 * Recursively traverses the node tree to find text nodes.
 *
 * @param {Object} node The TipTap JSON node (e.g., the root document)
 * @returns {string} The combined raw text
 */
export function extractTextFromJson(node) {
  if (!node) return ''

  let text = ''
  if (node.type === 'text' && node.text) {
    text += node.text
  }

  if (node.content && Array.isArray(node.content)) {
    for (const child of node.content) {
      const childText = extractTextFromJson(child)
      if (childText) {
        text += childText + ' '
      }
    }
    // Add a space between block nodes like paragraphs
    if (node.type && node.type !== 'text') {
      text += ' '
    }
  }

  return text
}

/**
 * Calculates the word count and estimated reading time.
 * Assumes an average reading speed of 200 words per minute (WPM).
 *
 * @param {Object} json The TipTap JSON document
 * @returns {Object} { wordCount: number, readingTime: number (minutes), formattedTime: string }
 */
export function calculateReadingTime(json) {
  if (!json) return { wordCount: 0, readingTime: 1, formattedTime: '1 min' }

  const text = extractTextFromJson(json).trim()
  
  // Count words by splitting on whitespace
  // Use a regex to match words reliably, handling multiple spaces/newlines
  const words = text.match(/\S+/g) || []
  const wordCount = words.length

  // Formula: Minutes = Math.max(1, Math.round(WordCount / 200))
  const readingTime = Math.max(1, Math.round(wordCount / 200))
  
  // Format for display
  const hours = Math.floor(readingTime / 60)
  const mins = readingTime % 60
  
  let formattedTime = ''
  if (hours > 0) {
    formattedTime += `${hours} hour${hours > 1 ? 's' : ''} `
  }
  if (mins > 0 || hours === 0) {
    formattedTime += `${mins} min${mins > 1 ? 's' : ''}`
  }

  return {
    wordCount,
    readingTime,
    formattedTime: formattedTime.trim()
  }
}
