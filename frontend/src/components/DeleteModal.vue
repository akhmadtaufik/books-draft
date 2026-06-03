<template>
  <Teleport to="body">
    <div v-if="isOpen" class="modal-overlay" @click.self="cancel">
      <div class="modal-content">
        <div class="modal-icon">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-red-500"><path d="M3 6h18"/><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"/><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"/><line x1="10" x2="10" y1="11" y2="17"/><line x1="14" x2="14" y1="11" y2="17"/></svg>
        </div>
        <div class="modal-text">
          <h3 class="modal-title">{{ title }}</h3>
          <p class="modal-message">{{ message }}</p>
        </div>
        <div class="modal-actions">
          <button @click="cancel" class="btn-cancel">Cancel</button>
          <button @click="confirm" class="btn-danger">Delete</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
const props = defineProps({
  isOpen: Boolean,
  title: { type: String, default: 'Confirm Deletion' },
  message: { type: String, default: 'Are you sure you want to delete this item? This action cannot be undone.' }
})
const emit = defineEmits(['confirm', 'cancel'])

function confirm() { emit('confirm') }
function cancel() { emit('cancel') }
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: fadeIn 0.15s ease-out;
}

.modal-content {
  background: #18181b;
  border: 1px solid #27272a;
  border-radius: 8px;
  width: 90vw;
  max-width: 400px;
  padding: 1.5rem;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.5), 0 10px 10px -5px rgba(0, 0, 0, 0.2);
  display: flex;
  flex-direction: column;
  gap: 1rem;
  animation: slideUp 0.2s ease-out;
}

.modal-icon {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 0.5rem;
}

.modal-text {
  text-align: center;
}

.modal-title {
  margin: 0 0 0.5rem 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: #f4f4f5;
  font-family: 'Inter', sans-serif;
}

.modal-message {
  margin: 0;
  font-size: 0.875rem;
  color: #a1a1aa;
  font-family: 'Inter', sans-serif;
  line-height: 1.5;
}

.modal-actions {
  display: flex;
  gap: 0.75rem;
  margin-top: 1rem;
}

.modal-actions button {
  flex: 1;
  padding: 0.625rem;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  font-family: 'Inter', sans-serif;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel {
  background: transparent;
  border: 1px solid #3f3f46;
  color: #e4e4e7;
}

.btn-cancel:hover {
  background: #27272a;
}

.btn-danger {
  background: #ef4444;
  border: 1px solid #ef4444;
  color: #ffffff;
}

.btn-danger:hover {
  background: #dc2626;
  border-color: #dc2626;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes slideUp {
  from { transform: translateY(10px) scale(0.95); opacity: 0; }
  to { transform: translateY(0) scale(1); opacity: 1; }
}
</style>
