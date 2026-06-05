<template>
  <div v-if="isOpen" class="modal-overlay" @mousedown.self="$emit('close')">
    <div class="modal-content">
      <div class="modal-header">
        <h3>{{ initialData?.id ? 'Edit Note' : 'Create Note' }}</h3>
        <button @click="$emit('close')" class="btn-close">&times;</button>
      </div>
      
      <form @submit.prevent="submit" class="metadata-form">
        <div class="form-group full-width">
          <label>Title</label>
          <input v-model="localForm.title" placeholder="Note title..." required />
        </div>
        <div class="form-group full-width">
          <label>Content</label>
          <textarea v-model="localForm.content" rows="10" placeholder="Write your notes here..."></textarea>
        </div>
        <div class="modal-actions full-width">
          <button type="button" @click="$emit('close')" class="btn-secondary">Cancel</button>
          <button type="submit" class="btn-primary">Save Note</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  isOpen: Boolean,
  initialData: Object
})

const emit = defineEmits(['close', 'save'])
const localForm = ref({ title: '', content: '' })

watch(() => props.isOpen, (newVal) => {
  if (newVal && props.initialData) {
    localForm.value = {
      title: props.initialData.title || '',
      content: props.initialData.content?.text || ''
    }
  } else {
    localForm.value = { title: '', content: '' }
  }
})

function submit() {
  emit('save', { title: localForm.value.title, content: { text: localForm.value.content } })
}
</script>

<style scoped>
.modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.8); backdrop-filter: blur(4px); display: flex; align-items: center; justify-content: center; z-index: 100; }
.modal-content { background: #18181b; border-radius: 12px; width: 100%; max-width: 600px; border: 1px solid #27272a; box-shadow: 0 25px 50px -12px rgba(0,0,0,0.7); max-height: 90vh; overflow-y: auto; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 1.5rem 2rem; border-bottom: 1px solid #27272a; position: sticky; top: 0; background: #18181b; z-index: 10; }
.modal-header h3 { margin: 0; font-size: 1.25rem; font-weight: 600; color: #e4e4e7; }
.btn-close { background: transparent; border: none; color: #a1a1aa; font-size: 1.5rem; cursor: pointer; padding: 0; line-height: 1; }
.btn-close:hover { color: #fff; }

.metadata-form { display: grid; grid-template-columns: 1fr; gap: 1.25rem; padding: 2rem; }
.full-width { grid-column: span 2; }
.form-group { display: flex; flex-direction: column; gap: 0.5rem; }
.form-group label { font-size: 0.85rem; color: #a1a1aa; font-weight: 500; }
.form-group input, .form-group textarea { width: 100%; padding: 0.75rem 1rem; background: #0f0f12; border: 1px solid #3f3f46; color: #e4e4e7; border-radius: 6px; font-family: inherit; font-size: 0.95rem; transition: border-color 0.2s; box-sizing: border-box; }
.form-group input:focus, .form-group textarea:focus { outline: none; border-color: #8b5cf6; }
.form-group textarea { resize: vertical; min-height: 150px; font-family: 'Lora', serif; }

.modal-actions { display: flex; justify-content: flex-end; gap: 1rem; margin-top: 1rem; padding-top: 1.5rem; border-top: 1px solid #27272a; }
.btn-secondary { background: transparent; color: #e4e4e7; border: 1px solid #3f3f46; padding: 0.6rem 1.5rem; border-radius: 6px; cursor: pointer; font-weight: 500; transition: 0.2s; }
.btn-secondary:hover { background: #27272a; }
.btn-primary { background: #fafafa; color: #000; border: none; padding: 0.6rem 1.5rem; border-radius: 6px; cursor: pointer; font-weight: 600; transition: 0.2s; }
.btn-primary:hover { background: #e4e4e7; }
</style>
