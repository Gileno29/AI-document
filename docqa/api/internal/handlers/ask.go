package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"docqa/internal/llm"
	"docqa/internal/python"
	"docqa/internal/text"
)

type Handler struct {
	python    *python.Client
	llm       *llm.Client
	uploadDir string
}

func New(p *python.Client, l *llm.Client, uploadDir string) *Handler {
	return &Handler{python: p, llm: l, uploadDir: uploadDir}
}

// GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /upload  multipart/form-data  field: "file"
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "file too large or invalid form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".txt" && ext != ".pdf" {
		writeError(w, http.StatusUnprocessableEntity, "only .txt and .pdf are supported")
		return
	}

	docID := generateID(header.Filename)
	dest := filepath.Join(h.uploadDir, docID+ext)

	out, err := os.Create(dest)
	if err != nil {
		slog.Error("create file", "err", err)
		writeError(w, http.StatusInternalServerError, "could not save file")
		return
	}
	defer out.Close()
	io.Copy(out, file)

	raw, err := text.Extract(dest, ext)
	if err != nil {
		slog.Error("extract text", "err", err)
		writeError(w, http.StatusInternalServerError, "could not extract text from file")
		return
	}

	chunks := text.Chunk(raw, 500, 50)

	if err := h.python.Ingest(r.Context(), docID, chunks); err != nil {
		slog.Error("ingest", "err", err)
		writeError(w, http.StatusBadGateway, "embed service error")
		return
	}

	slog.Info("uploaded", "doc_id", docID, "chunks", len(chunks), "file", header.Filename)
	writeJSON(w, http.StatusOK, map[string]any{
		"doc_id": docID,
		"chunks": len(chunks),
	})
}

// POST /ask  JSON {"doc_id":"…","question":"…"}
func (h *Handler) Ask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DocID    string `json:"doc_id"`
		Question string `json:"question"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DocID == "" || req.Question == "" {
		writeError(w, http.StatusBadRequest, "body must be {doc_id, question}")
		return
	}

	chunks, err := h.python.Retrieve(r.Context(), req.DocID, req.Question)
	if err != nil {
		slog.Error("retrieve", "err", err)
		writeError(w, http.StatusBadGateway, "embed service error")
		return
	}

	context := strings.Join(chunks, "\n\n---\n\n")
	prompt := fmt.Sprintf(
		"You are a helpful assistant. Answer the question using ONLY the context below.\n\n"+
			"Context:\n%s\n\nQuestion: %s\nAnswer:",
		context, req.Question,
	)

	answer, err := h.llm.Complete(r.Context(), prompt)
	if err != nil {
		slog.Error("llm complete", "err", err)
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"doc_id":   req.DocID,
		"question": req.Question,
		"answer":   answer,
		"sources":  len(chunks),
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func generateID(filename string) string {
	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		if r >= 'A' && r <= 'Z' {
			return r + 32
		}
		return '-'
	}, base)
	return fmt.Sprintf("%s-%d", safe, os.Getpid())
}
