<template>
  <div class="toolbar">
    <button @click="editor.chain().focus().toggleBold().run()" :class="{ 'is-active': editor.isActive('bold') }"><b>B</b></button>
    <button @click="editor.chain().focus().toggleItalic().run()" :class="{ 'is-active': editor.isActive('italic') }"><i>I</i></button>
    <button @click="editor.chain().focus().toggleStrike().run()" :class="{ 'is-active': editor.isActive('strike') }"><s>S</s></button>
    <div class="divider"></div>
    <button @click="editor.chain().focus().toggleHeading({ level: 1 }).run()" :class="{ 'is-active': editor.isActive('heading', { level: 1 }) }">H1</button>
    <button @click="editor.chain().focus().toggleHeading({ level: 2 }).run()" :class="{ 'is-active': editor.isActive('heading', { level: 2 }) }">H2</button>
    <button @click="editor.chain().focus().toggleHeading({ level: 3 }).run()" :class="{ 'is-active': editor.isActive('heading', { level: 3 }) }">H3</button>
    <div class="divider"></div>
    <button @click="editor.chain().focus().toggleBulletList().run()" :class="{ 'is-active': editor.isActive('bulletList') }">• List</button>
    <button @click="editor.chain().focus().toggleOrderedList().run()" :class="{ 'is-active': editor.isActive('orderedList') }">1. List</button>
    <button @click="editor.chain().focus().toggleBlockquote().run()" :class="{ 'is-active': editor.isActive('blockquote') }">" Quote</button>
    <div class="divider"></div>
    <button @click="editor.chain().focus().undo().run()" :disabled="!editor.can().undo()">Undo</button>
    <button @click="editor.chain().focus().redo().run()" :disabled="!editor.can().redo()">Redo</button>
    <div class="divider"></div>
    <button @click="$emit('save-milestone')" :disabled="isSavingMilestone" class="btn-milestone">
      {{ isSavingMilestone ? '...' : '⭐ Save Milestone' }}
    </button>
    <button @click="$emit('show-history')" class="btn-history">⏱ History</button>
  </div>
</template>

<script setup>
defineProps({
  editor: { type: Object, required: true },
  isSavingMilestone: Boolean
})
defineEmits(['save-milestone', 'show-history'])
</script>

<style scoped>
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
</style>
