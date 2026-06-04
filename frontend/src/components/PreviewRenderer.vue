<template>
  <div class="preview-tiptap" v-html="htmlContent"></div>
</template>

<script setup>
import { computed } from 'vue'
import { generateHTML } from '@tiptap/html'
import StarterKit from '@tiptap/starter-kit'

const props = defineProps({
  content: { type: Object, required: true }
})

// Convert TipTap JSON directly to a lightweight static HTML string
const htmlContent = computed(() => {
  if (!props.content || Object.keys(props.content).length === 0) return ''
  try {
    return generateHTML(props.content, [StarterKit])
  } catch (e) {
    console.error("Failed to parse chapter content", e)
    return ''
  }
})
</script>

<style scoped>
/* Scoped to the preview renderer to avoid conflicting with the main editor */
.preview-tiptap {
  font-size: 13px !important;
  line-height: 1.6;
  color: #27272a;
}

.preview-tiptap :deep(p) {
  margin-bottom: 0.8em;
  color: #111827 !important;
  text-align: justify;
  break-inside: auto;
  orphans: 2;
  widows: 2;
}

.preview-tiptap :deep(h1), 
.preview-tiptap :deep(h2), 
.preview-tiptap :deep(h3) {
  color: #18181b;
  margin-top: 1.2em;
  margin-bottom: 0.5em;
  font-family: 'Inter', sans-serif;
  break-after: avoid;
  break-inside: avoid;
}

.preview-tiptap :deep(blockquote) {
  border-left: 3px solid #d4d4d8;
  padding-left: 1rem;
  color: #52525b;
  font-style: italic;
  break-inside: avoid;
}

.preview-tiptap :deep(img), 
.preview-tiptap :deep(pre),
.preview-tiptap :deep(ul),
.preview-tiptap :deep(ol) {
  break-inside: avoid;
}
</style>
