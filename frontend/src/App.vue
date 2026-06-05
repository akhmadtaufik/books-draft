<template>
  <div class="app-shell">
    <template v-if="isPreviewMode">
      <BookPreview :bookId="currentBookId" @close="isPreviewMode = false" />
    </template>
    
    <template v-else>
      <header class="app-header">
        <div class="header-left">
          <button v-if="currentBookId" @click="currentBookId = null" class="btn-icon back-btn" title="Back to Library">⬅️</button>
          <h1>Novel Drafting App</h1>
          <span class="app-badge">v0.1 — MVP</span>
        </div>
        
        <div class="header-right" v-if="currentBookId">
          <button @click="isStoryBibleOpen = !isStoryBibleOpen" class="btn-story-bible" :class="{ active: isStoryBibleOpen }">
            📖 Story Bible
          </button>
          <button @click="exportToEpub" class="btn-primary" :disabled="isExporting">
            {{ isExporting ? 'Exporting...' : 'Export EPUB' }}
          </button>
          <button @click="exportToPdf" class="btn-primary" :disabled="isExportingPdf">
            {{ isExportingPdf ? 'Exporting...' : 'Export PDF' }}
          </button>
          <button @click="isPreviewMode = true" class="btn-primary">
            Preview Book
          </button>
        </div>
      </header>
      
      <main class="app-main">
        <template v-if="!currentBookId">
          <BookDashboard 
            :books="booksList" 
            :isLoading="isLoading"
            @open-book="openBook"
            @book-created="onBookCreated"
            @book-updated="onBookUpdated"
            @book-deleted="onBookDeleted"
          />
        </template>
        
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

          <StoryBible
            :bookId="currentBookId"
            :isOpen="isStoryBibleOpen"
            @toggle="isStoryBibleOpen = !isStoryBibleOpen"
          />
        </template>
        
        <div v-if="isLoading" class="loading-app">Loading...</div>
      </main>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { get, post } from './composables/useApi.js'
import { generateHTML } from '@tiptap/html'
import StarterKit from '@tiptap/starter-kit'
import ChapterSidebar from './components/ChapterSidebar.vue'
import TipTapEditor from './components/TipTapEditor.vue'
import BookPreview from './components/BookPreview.vue'
import StoryBible from './components/StoryBible.vue'
import BookDashboard from './components/BookDashboard.vue'

const booksList = ref([])

const currentBookId = ref(null)
const currentChapterId = ref(null)
const isPreviewMode = ref(false)
const isStoryBibleOpen = ref(false)
const sidebarRef = ref(null)

const newBookTitle = ref('')
const isLoading = ref(true)
const isExporting = ref(false)
const isExportingPdf = ref(false)

async function exportToPdf() {
  if (!currentBookId.value) return
  isExportingPdf.value = true
  
  try {
    const previewData = await get(`/api/books/${currentBookId.value}/preview`)
    
    let chaptersHtml = ''
    let tocHtml = '<div class="toc-page"><h2>Daftar Isi</h2><ul class="toc-list">'
    
    previewData.chapters.forEach((chapter, index) => {
      let bodyHtml = ''
      if (chapter.content && Object.keys(chapter.content).length > 0) {
        try {
          bodyHtml = generateHTML(chapter.content, [StarterKit])
        } catch (e) {
          console.error('Error generating HTML for chapter:', e)
        }
      }
      
      const chapterId = `chapter-${index + 1}`
      
      // TOC HTML using Flexbox for dots and an empty span for JS to inject the page number
      tocHtml += `
        <li>
          <a href="#${chapterId}">
            <span class="toc-text">${chapter.title || 'Untitled'}</span>
            <span class="toc-dots"></span>
            <span class="toc-page-num" data-target="${chapterId}"></span>
          </a>
        </li>
      `
      
      // Clean Chapter HTML (ID is purely on the H1 now)
      chaptersHtml += `
        <div class="chapter">
          <h1 class="chapter-title" id="${chapterId}">${chapter.title || 'Untitled'}</h1>
          <div class="chapter-content">
            ${bodyHtml}
          </div>
        </div>
      `
    })
    
    tocHtml += '</ul></div>'

    const fullHtml = `
      <!DOCTYPE html>
      <html lang="en">
      <head>
        <meta charset="UTF-8">
        <title>${previewData.title || 'Book'} - Print Ready</title>
        <script src="https://unpkg.com/pagedjs/dist/paged.polyfill.js"><\/script>
        <script>
          // Bulletproof JS Hook to bypass Chrome's broken CSS counters
          class PrintHandler extends Paged.Handler {
            afterRendered(pages) {
              let isMainMatter = false;
              let pageNum = 1;
              let chapterPages = {};

              // 1. Calculate Page Numbers & Find Chapters
              pages.forEach(page => {
                // Determine if this page is part of the main story
                if (page.element.querySelector('.main-matter')) {
                  isMainMatter = true;
                }

                if (isMainMatter) {
                  // Inject page number into the bottom center margin box
                  const bottomCenter = page.element.querySelector('.pagedjs_margin-bottom-center .pagedjs_margin-content');
                  if (bottomCenter) {
                    bottomCenter.innerText = pageNum;
                  }

                  // Record the start page of each chapter
                  const chapterTitles = page.element.querySelectorAll('h1.chapter-title');
                  chapterTitles.forEach(title => {
                    if (title.id && !chapterPages[title.id]) {
                      chapterPages[title.id] = pageNum;
                    }
                  });
                  pageNum++;
                }
              });

              // 2. Inject calculated numbers into the rendered Table of Contents
              const renderedTocNums = document.querySelectorAll('.pagedjs_pages .toc-page-num');
              renderedTocNums.forEach(span => {
                const targetId = span.getAttribute('data-target');
                if (chapterPages[targetId]) {
                  span.innerText = chapterPages[targetId];
                }
              });

              // Trigger print dialog
              setTimeout(() => { window.print(); }, 500);
            }
          }
          Paged.registerHandlers(PrintHandler);
        <\/script>
        <style>
          /* --------------------------------------------------------
             PAGE SETUP
             -------------------------------------------------------- */
          @page {
            size: 14cm 20cm;
            margin: 1.5cm 1.5cm 2cm 1.5cm;
            @bottom-center {
              content: ""; /* Strictly left empty for JS injection */
              font-family: "Georgia", serif;
              font-size: 10pt;
            }
          }

          /* --------------------------------------------------------
             GLOBAL TYPOGRAPHY
             -------------------------------------------------------- */
          body {
            font-family: "Georgia", "Times New Roman", serif;
            font-size: 12pt;
            line-height: 1.6;
            text-align: justify;
            color: #000;
            background: #fff;
          }

          /* --------------------------------------------------------
             FRONTMATTER (TITLE & TOC)
             -------------------------------------------------------- */
          .title-page {
            break-after: right; 
            text-align: center;
          }
          .title-page h1 {
            font-size: 24pt;
            text-transform: uppercase;
            letter-spacing: 2px;
            font-weight: normal;
            padding-top: 40%;
          }

          .toc-page h2 {
            text-align: center;
            font-size: 18pt;
            margin-top: 3cm;
            margin-bottom: 2cm;
            text-transform: uppercase;
            letter-spacing: 2px;
            font-weight: normal;
          }
          .toc-list {
            list-style: none;
            padding: 0;
            margin: 0;
          }
          .toc-list li {
            margin-bottom: 1.2em; 
            line-height: 1.5;
          }
          
          /* Flexbox TOC Layout */
          .toc-list a {
            display: flex;
            align-items: baseline;
            text-decoration: none;
            color: #000;
            width: 100%;
          }
          .toc-text {
            white-space: nowrap;
          }
          .toc-dots {
            flex-grow: 1;
            border-bottom: 2px dotted #000;
            margin: 0 8px;
            position: relative;
            top: -4px; 
          }
          .toc-page-num {
            white-space: nowrap;
          }

          /* --------------------------------------------------------
             CHAPTERS (MAIN CONTENT)
             -------------------------------------------------------- */
          .main-matter {
            break-before: right; 
          }

          .chapter {
            break-before: right; 
          }
          
          .chapter:first-of-type {
            break-before: avoid; 
          }

          .chapter-title {
            text-align: center;
            font-size: 18pt;
            font-weight: normal;
            margin-top: 3cm;
            margin-bottom: 2cm;
            text-transform: uppercase;
            letter-spacing: 2px;
          }
          
          .chapter-content p {
            margin: 0;
            margin-bottom: 1.2em; 
            text-indent: 0 !important; 
            widows: 2;
            orphans: 2;
          }
        </style>
      </head>
      <body>
        <div class="title-page">
          <div style="margin-top: 50%;">
            <h1>${previewData.title || 'Untitled Book'}</h1>
          </div>
        </div>
        ${tocHtml}
        <div class="main-matter">
          ${chaptersHtml}
        </div>
      </body>
      </html>
    `

    const printWindow = window.open('', '_blank')
    if (printWindow) {
      printWindow.document.open()
      printWindow.document.write(fullHtml)
      printWindow.document.close()
    } else {
      alert("Please allow popups to export to PDF.")
    }

  } catch (err) {
    console.error('Failed to generate PDF:', err)
  } finally {
    isExportingPdf.value = false
  }
}

async function exportToEpub() {
  if (!currentBookId.value || isExporting.value) return
  
  isExporting.value = true
  try {
    const preview = await get(`/api/books/${currentBookId.value}/preview`)
    
    const mappedChapters = preview.chapters.map(chapter => {
      let html = ''
      if (chapter.content && Object.keys(chapter.content).length > 0) {
        try {
          html = generateHTML(chapter.content, [StarterKit])
        } catch (e) {
          console.error('Error generating HTML for chapter:', e)
        }
      }
      return {
        title: chapter.title,
        html: html
      }
    })
    
    const payload = {
      title: preview.title || 'Exported Book',
      chapters: mappedChapters
    }
    
    const response = await fetch(`/api/books/${currentBookId.value}/export/epub`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(payload)
    })
    
    if (!response.ok) {
      throw new Error(`Export failed with status: ${response.status}`)
    }
    
    const blob = await response.blob()
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    
    let filename = 'book_export.epub'
    const contentDisposition = response.headers.get('Content-Disposition')
    if (contentDisposition && contentDisposition.includes('filename="')) {
      filename = contentDisposition.split('filename="')[1].split('"')[0]
    }
    
    a.download = filename
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    window.URL.revokeObjectURL(url)
    
  } catch (err) {
    console.error('Failed to export EPUB:', err)
    alert('Failed to export EPUB. Please try again.')
  } finally {
    isExporting.value = false
  }
}

onMounted(async () => {
  try {
    const books = await get('/api/books')
    if (books && books.length > 0) {
      booksList.value = books
      // Commented out to default to the dashboard view
      // currentBookId.value = books[0].id
      // const chapters = await get(`/api/books/${currentBookId.value}/chapters`)
      // if (chapters && chapters.length > 0) {
      //   currentChapterId.value = chapters[0].id
      // }
    }
  } catch (err) {
    console.error('Failed to load books:', err)
  } finally {
    isLoading.value = false
  }
})

async function openBook(id) {
  currentBookId.value = id
  try {
    const chapters = await get(`/api/books/${id}/chapters`)
    if (chapters && chapters.length > 0) {
      currentChapterId.value = chapters[0].id
    } else {
      currentChapterId.value = null
    }
  } catch (err) {
    console.error('Failed to load chapters for book:', err)
  }
}

function onBookCreated(newBook) {
  booksList.value.unshift(newBook)
}

function onBookUpdated(updatedBook) {
  const index = booksList.value.findIndex(b => b.id === updatedBook.id)
  if (index !== -1) {
    booksList.value[index] = updatedBook
  }
}

function onBookDeleted(bookId) {
  booksList.value = booksList.value.filter(b => b.id !== bookId)
  if (currentBookId.value === bookId) {
    currentBookId.value = null
    currentChapterId.value = null
  }
}

async function createInitialBook() {
  if (!newBookTitle.value) return
  isLoading.value = true
  try {
    const book = await post('/api/books', { title: newBookTitle.value })
    booksList.value.unshift(book)
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

.back-btn {
  background: transparent;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
  transition: transform 0.2s;
}

.back-btn:hover {
  transform: translateX(-3px);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 0.75rem;
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
