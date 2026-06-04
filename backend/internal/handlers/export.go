package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bmaupin/go-epub"
)

// ExportEPUBRequest is the payload expected by the ExportEPUB handler.
type ExportEPUBRequest struct {
	Title    string             `json:"title"`
	Chapters []ExportChapterReq `json:"chapters"`
}

// ExportChapterReq represents a single chapter in the export payload.
type ExportChapterReq struct {
	Title string `json:"title"`
	HTML  string `json:"html"`
}

// tradePaperbackCSS enforces premium typography layout,
// designed for comfortable reading on e-readers and print-on-demand.
const tradePaperbackCSS = `@page {
  margin: 5%;
}
body {
  font-family: "Palatino Linotype", "Book Antiqua", Palatino, serif;
  font-size: 12pt;
  line-height: 1.65;
  text-align: justify;
  color: #000000;
}
.chapter-heading {
  text-align: center;
  font-family: "Helvetica Neue", Helvetica, Arial, sans-serif;
  font-size: 1.75em;
  font-weight: normal;
  margin-top: 3em;      /* Guaranteed breathing room at the top */
  margin-bottom: 3em;   /* Wide, traditional gap before the text starts */
  page-break-before: always;
  text-transform: uppercase;
  letter-spacing: 2px;
}
.chapter-body p {
  margin-top: 0;
  margin-bottom: 0.6em; /* Adds a comfortable vertical gap between paragraphs */
  text-indent: 1.5em;   /* Keeps the traditional novel indentation */
  widows: 2;
  orphans: 2;
}
/* Ensure the very first paragraph after the title is pristine */
.chapter-body p:first-of-type {
  text-indent: 0;       /* Standard publishing rule: no indent on first paragraph */
  margin-top: 1em;      /* Extra fallback gap just in case */
}
`

// ExportEPUB handles POST /api/books/{id}/export/epub.
// It receives pre-rendered HTML chapters from the frontend and packages
// them into a standards-compliant EPUB file with an auto-generated TOC.
func (h *BookHandler) ExportEPUB(w http.ResponseWriter, r *http.Request) {
	var payload ExportEPUBRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if payload.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if len(payload.Chapters) == 0 {
		writeError(w, http.StatusBadRequest, "at least one chapter is required")
		return
	}

	// Initialize the EPUB.
	e := epub.NewEpub(payload.Title)
	e.SetAuthor("Novel Drafting App")

	// Write the CSS to a temp file — go-epub's AddCSS expects a file path, not a raw string.
	cssTmpFile, err := os.CreateTemp("", "epub-style-*.css")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create CSS temp file")
		return
	}
	if _, err := cssTmpFile.WriteString(tradePaperbackCSS); err != nil {
		cssTmpFile.Close()
		os.Remove(cssTmpFile.Name())
		writeError(w, http.StatusInternalServerError, "failed to write CSS temp file")
		return
	}
	cssTmpFile.Close()
	defer os.Remove(cssTmpFile.Name())

	cssPath, err := e.AddCSS(cssTmpFile.Name(), "trade-paperback.css")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add CSS to EPUB")
		return
	}

	// Add each chapter as a section — this auto-generates the TOC.
	for _, ch := range payload.Chapters {
		chapterTitle := ch.Title
		if chapterTitle == "" {
			chapterTitle = "Untitled Chapter"
		}

		// Construct premium chapter HTML
		chapterContent := fmt.Sprintf("<h1 class=\"chapter-heading\">%s</h1>\n<div class=\"chapter-body\">\n%s\n</div>", chapterTitle, ch.HTML)

		_, err := e.AddSection(chapterContent, chapterTitle, "", cssPath)
		if err != nil {
			writeError(w, http.StatusInternalServerError,
				fmt.Sprintf("failed to add chapter %q: %v", chapterTitle, err))
			return
		}
	}

	// Write the EPUB to a temp file, then stream it to the response.
	// go-epub requires writing to a filesystem path.
	tmpFile, err := os.CreateTemp("", "novel-export-*.epub")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create temp file")
		return
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := e.Write(tmpPath); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate EPUB")
		return
	}

	// Build a safe filename from the book title.
	safeTitle := sanitizeFilename(payload.Title)

	// Override the Content-Type set by the JSONContent middleware.
	w.Header().Set("Content-Type", "application/epub+zip")
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s.epub"`, safeTitle))

	http.ServeFile(w, r, tmpPath)
}

// sanitizeFilename strips characters that are unsafe for filenames.
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "-", "\\", "-", ":", "-", "*", "", "?", "",
		"\"", "", "<", "", ">", "", "|", "",
	)
	safe := replacer.Replace(name)
	safe = strings.TrimSpace(safe)
	if safe == "" {
		safe = "book"
	}
	return safe
}
