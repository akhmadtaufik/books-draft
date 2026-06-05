import { bookApi } from '../api/bookApi.js'
import { generateHTML } from '@tiptap/html'
import StarterKit from '@tiptap/starter-kit'

export async function generatePdfExport(bookId) {
  if (!bookId) return
  
  try {
    const previewData = await bookApi.getPreview(bookId)
    
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
      
      tocHtml += `
        <li>
          <a href="#${chapterId}">
            <span class="toc-text">${chapter.title || 'Untitled'}</span>
            <span class="toc-dots"></span>
            <span class="toc-page-num" data-target="${chapterId}"></span>
          </a>
        </li>
      `
      
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
          class PrintHandler extends Paged.Handler {
            afterRendered(pages) {
              let isMainMatter = false;
              let pageNum = 1;
              let chapterPages = {};

              pages.forEach(page => {
                if (page.element.querySelector('.main-matter')) {
                  isMainMatter = true;
                }

                if (isMainMatter) {
                  const bottomCenter = page.element.querySelector('.pagedjs_margin-bottom-center .pagedjs_margin-content');
                  if (bottomCenter) {
                    bottomCenter.innerText = pageNum;
                  }

                  const chapterTitles = page.element.querySelectorAll('h1.chapter-title');
                  chapterTitles.forEach(title => {
                    if (title.id && !chapterPages[title.id]) {
                      chapterPages[title.id] = pageNum;
                    }
                  });
                  pageNum++;
                }
              });

              const renderedTocNums = document.querySelectorAll('.pagedjs_pages .toc-page-num');
              renderedTocNums.forEach(span => {
                const targetId = span.getAttribute('data-target');
                if (chapterPages[targetId]) {
                  span.innerText = chapterPages[targetId];
                }
              });

              setTimeout(() => { window.print(); }, 500);
            }
          }
          Paged.registerHandlers(PrintHandler);
        <\/script>
        <style>
          @page {
            size: 14cm 20cm;
            margin: 1.5cm 1.5cm 2cm 1.5cm;
            @bottom-center {
              content: ""; 
              font-family: "Georgia", serif;
              font-size: 10pt;
            }
          }
          body {
            font-family: "Georgia", "Times New Roman", serif;
            font-size: 12pt;
            line-height: 1.6;
            text-align: justify;
            color: #000;
            background: #fff;
          }
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
  }
}

export async function generateEpubExport(bookId) {
  if (!bookId) return
  
  try {
    const preview = await bookApi.getPreview(bookId)
    
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
    
    const response = await fetch(`/api/books/${bookId}/export/epub`, {
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
  }
}
