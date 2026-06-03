<template>
  <div class="app-shell">
    <template v-if="isPreviewMode">
      <BookPreview :bookId="currentBookId" @close="isPreviewMode = false" />
    </template>
    
    <template v-else>
      <header class="app-header">
        <div class="header-left">
          <h1>Novel Drafting App</h1>
          <span class="app-badge">v0.1 — MVP</span>
        </div>
        
        <div class="header-right" v-if="currentBookId">
          <button @click="isPreviewMode = true" class="btn-primary">
            Preview Book
          </button>
        </div>
      </header>
      
      <main class="app-main">
        <div v-if="!currentBookId && !isLoading" class="empty-book-state">
          <h2>Create Your First Book</h2>
          <form @click.prevent="createInitialBook" class="create-book-form">
            <input v-model="newBookTitle" placeholder="Enter book title..." required />
            <button type="submit" class="btn-primary">Create Book</button>
          </form>
        </div>
        
        <template v-else-if="currentBookId">
          <ChapterSidebar 
            ref="sidebarRef"
            :bookId="currentBookId" 
            :activeChapterId="currentChapterId"
            @select="currentChapterId = $event"
            @preview="isPreviewMode = true"
          />
          
          <section class="editor-area">
            <TipTapEditor 
              v-if="currentChapterId" 
              :chapterId="currentChapterId" 
              :key="currentChapterId"
              @title-updated="onChapterTitleUpdated"
            />
            <div v-else class="empty-editor-state">
              <p>Select a chapter from the sidebar or create a new one to start writing.</p>
            </div>
          </section>
        </template>
        
        <div v-if="isLoading" class="loading-app">Loading...</div>
      </main>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { get, post } from './composables/useApi.js'
import ChapterSidebar from './components/ChapterSidebar.vue'
import TipTapEditor from './components/TipTapEditor.vue'
import BookPreview from './components/BookPreview.vue'

const currentBookId = ref(null)
const currentChapterId = ref(null)
const isPreviewMode = ref(false)
const sidebarRef = ref(null)

const newBookTitle = ref('')
const isLoading = ref(true)

onMounted(async () => {
  try {
    const books = await get('/api/books')
    if (books && books.length > 0) {
      currentBookId.value = books[0].id
      
      // Load chapters for the first book
      const chapters = await get(`/api/books/${currentBookId.value}/chapters`)
      if (chapters && chapters.length > 0) {
        currentChapterId.value = chapters[0].id
      }
    }
  } catch (err) {
    console.error('Failed to load books:', err)
  } finally {
    isLoading.value = false
  }
})

async function createInitialBook() {
  if (!newBookTitle.value) return
  isLoading.value = true
  try {
    const book = await post('/api/books', { title: newBookTitle.value })
    currentBookId.value = book.id
    newBookTitle.value = ''
  } catch (err) {
    console.error('Failed to create book:', err)
  } finally {
    isLoading.value = false
  }
}

function onChapterTitleUpdated({ id, title }) {
  if (sidebarRef.value) {
    sidebarRef.value.updateChapterTitle(id, title)
  }
}
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
</style>

<style scoped>
.app-shell {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1.5rem;
  background: #18181b;
  border-bottom: 1px solid #27272a;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.app-header h1 {
  font-size: 1.125rem;
  font-weight: 600;
  margin: 0;
}

.app-badge {
  font-size: 0.7rem;
  padding: 0.2rem 0.6rem;
  border-radius: 999px;
  background: #3b0764;
  color: #c084fc;
}

.app-main {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.editor-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background-color: #0f0f12;
}

.empty-editor-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #71717a;
}

.empty-book-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1.5rem;
}

.empty-book-state h2 {
  font-size: 1.5rem;
  font-weight: 500;
  margin: 0;
}

.create-book-form {
  display: flex;
  gap: 0.5rem;
}

.create-book-form input {
  padding: 0.5rem 1rem;
  border-radius: 6px;
  border: 1px solid #3f3f46;
  background: #18181b;
  color: #fff;
  font-family: inherit;
  font-size: 0.875rem;
  outline: none;
  min-width: 250px;
}

.create-book-form input:focus {
  border-color: #71717a;
}

.loading-app {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #71717a;
}
</style>
