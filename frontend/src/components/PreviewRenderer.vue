<template>
  <editor-content :editor="editor" class="preview-tiptap" />
</template>

<script setup>
import { onBeforeUnmount } from 'vue'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'

const props = defineProps({
  content: { type: Object, required: true }
})

const editor = useEditor({
  editable: false,
  extensions: [StarterKit],
  content: props.content,
})

onBeforeUnmount(() => {
  if (editor.value) {
    editor.value.destroy()
  }
})
</script>

<style>
/* Scoped to the preview renderer to avoid conflicting with the main editor */
.preview-tiptap {
  font-size: 0.95rem;
  line-height: 1.6;
  color: #27272a;
}

.preview-tiptap p {
  margin-bottom: 1.2em;
  color: #111827 !important;
  text-align: justify;
  break-inside: auto;
  orphans: 2;
  widows: 2;
}

.preview-tiptap h1, .preview-tiptap h2, .preview-tiptap h3 {
  color: #18181b;
  margin-top: 2em;
  margin-bottom: 1em;
  font-family: 'Inter', sans-serif;
  break-after: avoid;
  break-inside: avoid;
}

.preview-tiptap blockquote {
  border-left: 3px solid #d4d4d8;
  padding-left: 1rem;
  color: #52525b;
  font-style: italic;
  break-inside: avoid;
}

.preview-tiptap img, 
.preview-tiptap pre,
.preview-tiptap ul,
.preview-tiptap ol {
  break-inside: avoid;
}
</style>
