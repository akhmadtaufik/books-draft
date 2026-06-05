<script setup>
import { ref, onMounted } from 'vue'
import { get, post, put, del } from '../composables/useApi.js'

const props = defineProps({
  books: {
    type: Array,
    required: true
  },
  isLoading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['open-book', 'book-created', 'book-updated', 'book-deleted'])

const showModal = ref(false)
const isEditing = ref(false)
const isSaving = ref(false)
const titleInput = ref(null)

const defaultFormState = {
  id: '',
  title: '',
  author: '',
  genre: '',
  synopsis: '',
  language: 'Indonesian',
  status: 'Draft',
  isbn: '',
  publisher: ''
}

const bookForm = ref({ ...defaultFormState })

function openCreateModal() {
  isEditing.value = false
  bookForm.value = { ...defaultFormState }
  showModal.value = true
  setTimeout(() => {
    if (titleInput.value) titleInput.value.focus()
  }, 100)
}

function openEditModal(book) {
  isEditing.value = true
  bookForm.value = { 
    id: book.id,
    title: book.title || '',
    author: book.author || '',
    genre: book.genre || '',
    synopsis: book.synopsis || '',
    language: book.language || 'Indonesian',
    status: book.status || 'Draft',
    isbn: book.isbn || '',
    publisher: book.publisher || ''
  }
  showModal.value = true
  setTimeout(() => {
    if (titleInput.value) titleInput.value.focus()
  }, 100)
}

function closeModal() {
  showModal.value = false
  bookForm.value = { ...defaultFormState }
}

async function submitBook() {
  if (!bookForm.value.title) return
  isSaving.value = true
  
  try {
    const payload = { ...bookForm.value }
    delete payload.id // Don't send id in body typically
    
    if (isEditing.value) {
      const updatedBook = await put(`/api/books/${bookForm.value.id}`, payload)
      emit('book-updated', updatedBook)
    } else {
      const newBook = await post('/api/books', payload)
      emit('book-created', newBook)
    }
    closeModal()
  } catch (err) {
    console.error('Failed to save book:', err)
    alert('Failed to save book. Please try again.')
  } finally {
    isSaving.value = false
  }
}

async function confirmDelete(bookId) {
  if (confirm('Are you sure you want to delete this book? This action cannot be undone.')) {
    try {
      await del(`/api/books/${bookId}`)
      emit('book-deleted', bookId)
    } catch (err) {
      console.error('Failed to delete book:', err)
      alert('Failed to delete book. Please try again.')
    }
  }
}
</script>

<template>
  <div class="dashboard-container">
    <div class="dashboard-header">
      <div class="header-titles">
        <h2>My Library</h2>
        <p class="subtitle">Manage your book metadata and manuscripts</p>
      </div>
      <button @click="openCreateModal" class="btn-primary">+ Create New Book</button>
    </div>

    <div v-if="isLoading" class="loading-state">Loading books...</div>
    
    <div v-else-if="books.length === 0" class="empty-state">
      <div class="empty-icon">📚</div>
      <h3>Your library is empty</h3>
      <p>Start your writing journey by creating your first book.</p>
    </div>

    <div v-else class="books-grid">
      <div v-for="book in books" :key="book.id" class="book-card">
        <div class="book-status-badge" :class="book.status?.toLowerCase() || 'draft'">
          {{ book.status || 'Draft' }}
        </div>
        <div class="book-info">
          <h3>{{ book.title }}</h3>
          <p class="book-author">by {{ book.author || 'Unknown Author' }}</p>
          <p class="book-synopsis">{{ book.synopsis || 'No synopsis provided.' }}</p>
          <div class="book-meta-tags">
            <span v-if="book.genre" class="meta-tag">{{ book.genre }}</span>
            <span class="meta-tag">{{ new Date(book.updated_at || book.created_at).toLocaleDateString() }}</span>
          </div>
        </div>
        <div class="book-actions">
          <button @click="$emit('open-book', book.id)" class="btn-open">Open Editor</button>
          <button @click="openEditModal(book)" class="btn-icon" title="Edit Metadata">⚙️</button>
          <button @click="confirmDelete(book.id)" class="btn-icon danger" title="Delete Book">🗑️</button>
        </div>
      </div>
    </div>

    <div v-if="showModal" class="modal-overlay" @mousedown.self="closeModal">
      <div class="modal-content">
        <div class="modal-header">
          <h3>{{ isEditing ? 'Edit Book Metadata' : 'Create New Book' }}</h3>
          <button @click="closeModal" class="btn-close">&times;</button>
        </div>
        
        <form @submit.prevent="submitBook" class="metadata-form">
          <div class="form-group full-width">
            <label>Book Title <span class="required">*</span></label>
            <input v-model="bookForm.title" placeholder="e.g., The Great Gatsby" required ref="titleInput" />
          </div>
          
          <div class="form-group">
            <label>Author</label>
            <input v-model="bookForm.author" placeholder="Pen name or real name" />
          </div>
          
          <div class="form-group">
            <label>Genre / Category</label>
            <input v-model="bookForm.genre" placeholder="e.g., Sci-Fi, Romance, Non-Fiction" />
          </div>
          
          <div class="form-group full-width">
            <label>Synopsis</label>
            <textarea v-model="bookForm.synopsis" rows="4" placeholder="Brief description or blurb of the book..."></textarea>
          </div>
          
          <div class="form-group">
            <label>Language</label>
            <select v-model="bookForm.language">
              <option value="Indonesian">Indonesian</option>
              <option value="English">English</option>
              <option value="Other">Other</option>
            </select>
          </div>
          
          <div class="form-group">
            <label>Status</label>
            <select v-model="bookForm.status">
              <option value="Draft">Draft</option>
              <option value="Editing">Editing</option>
              <option value="Completed">Completed</option>
              <option value="Published">Published</option>
            </select>
          </div>
          
          <div class="form-group">
            <label>ISBN (Optional)</label>
            <input v-model="bookForm.isbn" placeholder="978-..." />
          </div>
          
          <div class="form-group">
            <label>Publisher</label>
            <input v-model="bookForm.publisher" placeholder="Publisher name or Self-Published" />
          </div>

          <div class="modal-actions full-width">
            <button type="button" @click="closeModal" class="btn-secondary">Cancel</button>
            <button type="submit" class="btn-primary" :disabled="isSaving">
              {{ isSaving ? 'Saving Metadata...' : 'Save Book' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Dashboard Layout */
.dashboard-container { padding: 2.5rem; max-width: 1200px; margin: 0 auto; width: 100%; color: #e4e4e7; overflow-y: auto; height: 100%; }
.dashboard-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 3rem; }
.header-titles h2 { font-size: 2rem; margin: 0 0 0.5rem 0; font-weight: 700; letter-spacing: -0.5px; }
.subtitle { color: #a1a1aa; margin: 0; font-size: 1rem; }

/* Empty State */
.empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 5rem 0; color: #71717a; text-align: center; }
.empty-icon { font-size: 4rem; margin-bottom: 1rem; opacity: 0.5; }
.empty-state h3 { font-size: 1.5rem; color: #e4e4e7; margin: 0 0 0.5rem 0; }

/* Grid & Cards */
.books-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 2rem; }
.book-card { background: #18181b; border: 1px solid #27272a; border-radius: 12px; padding: 1.5rem; display: flex; flex-direction: column; position: relative; transition: all 0.2s ease; }
.book-card:hover { transform: translateY(-4px); border-color: #52525b; box-shadow: 0 10px 30px -10px rgba(0,0,0,0.5); }
.book-status-badge { position: absolute; top: 1.5rem; right: 1.5rem; font-size: 0.7rem; font-weight: 600; text-transform: uppercase; padding: 0.25rem 0.75rem; border-radius: 99px; letter-spacing: 0.5px; }
.book-status-badge.draft { background: #3f3f46; color: #e4e4e7; }
.book-status-badge.editing { background: #7c2d12; color: #fdba74; }
.book-status-badge.completed { background: #14532d; color: #86efac; }
.book-status-badge.published { background: #1e3a8a; color: #93c5fd; }

.book-info { margin-top: 1.5rem; flex-grow: 1; }
.book-info h3 { margin: 0 0 0.25rem 0; font-size: 1.4rem; font-weight: 600; line-height: 1.3; padding-right: 4rem; }
.book-author { color: #a1a1aa; font-size: 0.9rem; margin: 0 0 1rem 0; font-style: italic; }
.book-synopsis { color: #71717a; font-size: 0.9rem; line-height: 1.5; margin: 0 0 1.5rem 0; display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; overflow: hidden; }

.book-meta-tags { display: flex; gap: 0.5rem; flex-wrap: wrap; margin-bottom: 1.5rem; }
.meta-tag { background: #27272a; color: #a1a1aa; font-size: 0.75rem; padding: 0.2rem 0.6rem; border-radius: 4px; }

.book-actions { display: flex; gap: 0.5rem; border-top: 1px solid #27272a; padding-top: 1.5rem; }
.btn-open { flex: 1; background: #fafafa; color: #000; font-weight: 600; border: none; padding: 0.6rem; border-radius: 6px; cursor: pointer; transition: 0.2s; }
.btn-open:hover { background: #e4e4e7; }
.btn-icon { background: #27272a; border: none; color: white; border-radius: 6px; padding: 0.6rem; cursor: pointer; transition: 0.2s; font-size: 1rem; display: flex; align-items: center; justify-content: center; width: 40px; }
.btn-icon:hover { background: #3f3f46; }
.btn-icon.danger:hover { background: #991b1b; }

/* Modal Design */
.modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.8); backdrop-filter: blur(4px); display: flex; align-items: center; justify-content: center; z-index: 100; }
.modal-content { background: #18181b; border-radius: 12px; width: 100%; max-width: 650px; border: 1px solid #27272a; box-shadow: 0 25px 50px -12px rgba(0,0,0,0.7); max-height: 90vh; overflow-y: auto; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 1.5rem 2rem; border-bottom: 1px solid #27272a; position: sticky; top: 0; background: #18181b; z-index: 10; }
.modal-header h3 { margin: 0; font-size: 1.25rem; font-weight: 600; }
.btn-close { background: transparent; border: none; color: #a1a1aa; font-size: 1.5rem; cursor: pointer; padding: 0; line-height: 1; }
.btn-close:hover { color: #fff; }

.metadata-form { display: grid; grid-template-columns: 1fr 1fr; gap: 1.25rem; padding: 2rem; }
.full-width { grid-column: span 2; }
.form-group { display: flex; flex-direction: column; gap: 0.5rem; }
.form-group label { font-size: 0.85rem; color: #a1a1aa; font-weight: 500; }
.required { color: #ef4444; }
.form-group input, .form-group textarea, .form-group select { width: 100%; padding: 0.75rem 1rem; background: #0f0f12; border: 1px solid #3f3f46; color: #e4e4e7; border-radius: 6px; font-family: inherit; font-size: 0.95rem; transition: border-color 0.2s; }
.form-group input:focus, .form-group textarea:focus, .form-group select:focus { outline: none; border-color: #8b5cf6; }
.form-group textarea { resize: vertical; min-height: 100px; }

.modal-actions { display: flex; justify-content: flex-end; gap: 1rem; margin-top: 1rem; padding-top: 1.5rem; border-top: 1px solid #27272a; }
.btn-secondary { background: transparent; color: #e4e4e7; border: 1px solid #3f3f46; padding: 0.6rem 1.5rem; border-radius: 6px; cursor: pointer; font-weight: 500; transition: 0.2s; }
.btn-secondary:hover { background: #27272a; }
.btn-primary { background: #fafafa; color: #000; border: none; padding: 0.6rem 1.5rem; border-radius: 6px; cursor: pointer; font-weight: 600; transition: 0.2s; }
.btn-primary:hover:not(:disabled) { background: #e4e4e7; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
