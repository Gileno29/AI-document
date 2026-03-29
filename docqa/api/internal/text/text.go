package text

import (
	"os/exec"
	"strings"
	"unicode/utf8"
)

// Extract returns plain text from a .txt or .pdf file.
// For PDFs it shells out to pdftotext (poppler-utils, included in the Dockerfile).
func Extract(path, ext string) (string, error) {
	if ext == ".txt" {
		b, err := readFile(path)
		return string(b), err
	}
	// PDF → pdftotext
	out, err := exec.Command("pdftotext", "-enc", "UTF-8", path, "-").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// Chunk splits text into overlapping windows of ~size runes with overlap runes
// of context carried over to the next chunk.
func Chunk(text string, size, overlap int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	runes := []rune(text)
	total := len(runes)
	var chunks []string

	for start := 0; start < total; {
		end := start + size
		if end > total {
			end = total
		}
		// Extend to the next sentence boundary (. ! ?) if within 60 runes
		if end < total {
			for i := end; i < end+60 && i < total; i++ {
				if runes[i] == '.' || runes[i] == '!' || runes[i] == '?' {
					end = i + 1
					break
				}
			}
		}
		chunk := strings.TrimSpace(string(runes[start:end]))
		if utf8.RuneCountInString(chunk) > 20 {
			chunks = append(chunks, chunk)
		}
		next := end - overlap
		if next <= start {
			next = start + 1
		}
		start = next
	}
	return chunks
}

func readFile(path string) ([]byte, error) {
	return exec.Command("cat", path).Output()
}
