package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

// ── fakes ─────────────────────────────────────────────────────────────────────

type fakePython struct {
	ingestErr   error
	retrieveErr error
	chunks      []string
}

func (f *fakePython) Ingest(_ context.Context, _ string, _ []string) error {
	return f.ingestErr
}
func (f *fakePython) Retrieve(_ context.Context, _, _ string) ([]string, error) {
	return f.chunks, f.retrieveErr
}

type fakeLLM struct {
	answer string
	err    error
}

func (f *fakeLLM) Complete(_ context.Context, _ string) (string, error) {
	return f.answer, f.err
}

// ── health ────────────────────────────────────────────────────────────────────

func TestHealth_Returns200(t *testing.T) {
	h := New(&fakePython{}, &fakeLLM{answer: "ok"}, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// ── upload ────────────────────────────────────────────────────────────────────

func TestUpload_ValidTxt(t *testing.T) {
	h := New(&fakePython{}, &fakeLLM{}, t.TempDir())

	body, ct := multipartFile(t, "test.txt", []byte("Hello world. This is a test document for chunking purposes."))
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()

	h.Upload(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["doc_id"] == "" {
		t.Error("expected doc_id in response")
	}
}

func TestUpload_InvalidExtension(t *testing.T) {
	h := New(&fakePython{}, &fakeLLM{}, t.TempDir())

	body, ct := multipartFile(t, "test.csv", []byte("a,b,c"))
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()

	h.Upload(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rec.Code)
	}
}

func TestUpload_MissingFile(t *testing.T) {
	h := New(&fakePython{}, &fakeLLM{}, t.TempDir())

	req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")
	rec := httptest.NewRecorder()

	h.Upload(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// ── ask ───────────────────────────────────────────────────────────────────────

func TestAsk_ValidRequest(t *testing.T) {
	python := &fakePython{chunks: []string{"chunk one", "chunk two"}}
	llm := &fakeLLM{answer: "This is the answer."}
	h := New(python, llm, t.TempDir())

	body, _ := json.Marshal(map[string]string{
		"doc_id":   "test-doc",
		"question": "What is this about?",
	})
	req := httptest.NewRequest(http.MethodPost, "/ask", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Ask(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["answer"] == "" {
		t.Error("expected answer in response")
	}
}

func TestAsk_MissingQuestion(t *testing.T) {
	h := New(&fakePython{}, &fakeLLM{}, t.TempDir())

	body, _ := json.Marshal(map[string]string{"doc_id": "test-doc"})
	req := httptest.NewRequest(http.MethodPost, "/ask", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Ask(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAsk_EmptyBody(t *testing.T) {
	h := New(&fakePython{}, &fakeLLM{}, t.TempDir())

	req := httptest.NewRequest(http.MethodPost, "/ask", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Ask(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func multipartFile(t *testing.T, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		t.Fatal(err)
	}
	fw.Write(content)
	w.Close()
	return &buf, w.FormDataContentType()
}
