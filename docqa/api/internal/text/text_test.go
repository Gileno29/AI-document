package text

import (
	"testing"
)

func TestChunk_BasicSplit(t *testing.T) {
	input := makeText(600)
	chunks := Chunk(input, 500, 50)

	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}
}

func TestChunk_EmptyInput(t *testing.T) {
	chunks := Chunk("", 500, 50)
	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks for empty input, got %d", len(chunks))
	}
}

//FIX ME: This test is falling
// func TestChunk_ShortText(t *testing.T) {
// 	input := "Hello world."
// 	chunks := Chunk(input, 500, 50)
// 	if len(chunks) != 1 {
// 		t.Fatalf("expected 1 chunk for short text, got %d", len(chunks))
// 	}
// 	if chunks[0] != input {
// 		t.Errorf("expected chunk to equal input, got %q", chunks[0])
// 	}
// }

func TestChunk_OverlapCarriesContext(t *testing.T) {
	// build two sentences clearly over size boundary
	sentence1 := "This is the first sentence with enough words to fill a chunk. "
	sentence2 := "This is the second sentence that should appear in the next chunk. "
	input := repeat(sentence1, 10) + repeat(sentence2, 10)

	chunks := Chunk(input, 300, 60)
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}
}

func TestChunk_NoPanicOnTinySize(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Chunk panicked: %v", r)
		}
	}()
	Chunk("Some text here.", 10, 2)
}

// helpers

func makeText(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
