<template>
  <div class="app-shell">
    <BookDashboard 
      v-if="!currentBookId" 
      @open-book="currentBookId = $event" 
    />
    
    <BookEditor 
      v-else 
      :bookId="currentBookId" 
      @close="currentBookId = null" 
    />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import BookDashboard from './components/BookDashboard.vue'
import BookEditor from './components/BookEditor.vue'

const currentBookId = ref(null)
</script>

<style>
/* Global resets and font imports */
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=Lora:ital,wght@0,400;0,600;1,400&display=swap');

*, *::before, *::after {
  box-sizing: border-box;
}

body {
  margin: 0;
  padding: 0;
  font-family: 'Inter', system-ui, -apple-system, sans-serif;
  background: #0f0f12;
  color: #e4e4e7;
  overflow: hidden; /* Prevent body scroll, handle scrolling in child elements */
}

/* Base button styles */
.btn-primary {
  background-color: #f4f4f5;
  color: #18181b;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
}

.btn-primary:hover {
  background-color: #d4d4d8;
}

.btn-story-bible {
  background: rgba(139, 92, 246, 0.1);
  color: #c4b5fd;
  border: 1px solid rgba(139, 92, 246, 0.3);
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  font-family: 'Inter', sans-serif;
}

.btn-story-bible:hover {
  background: rgba(139, 92, 246, 0.2);
  border-color: rgba(139, 92, 246, 0.5);
}

.btn-story-bible.active {
  background: rgba(139, 92, 246, 0.25);
  border-color: #8b5cf6;
  color: #e4e4e7;
}

/* Custom Scrollbars */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: transparent;
  border-radius: 4px;
}

*:hover::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.15);
}

.app-shell {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #0f0f12;
}
</style>
