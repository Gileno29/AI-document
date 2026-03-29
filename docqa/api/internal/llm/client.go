package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	provider    string
	apiKey      string
	ollamaURL   string
	ollamaModel string
	http        *http.Client
}

func NewClient() *Client {
	provider := os.Getenv("LLM_PROVIDER")
	if provider == "" {
		provider = "ollama"
	}

	key := os.Getenv("OPENAI_API_KEY")
	if provider == "anthropic" {
		key = os.Getenv("ANTHROPIC_API_KEY")
	}

	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://host.docker.internal:11434"
	}

	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "llama3.2"
	}

	return &Client{
		provider:    provider,
		apiKey:      key,
		ollamaURL:   ollamaURL,
		ollamaModel: ollamaModel,
		http:        &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *Client) Complete(ctx context.Context, prompt string) (string, error) {
	switch c.provider {
	case "anthropic":
		return c.anthropic(ctx, prompt)
	case "ollama":
		return c.ollama(ctx, prompt)
	default:
		return c.openai(ctx, prompt)
	}
}

// ── Ollama (local, free) ─────────────────────────────────────────────────────

func (c *Client) ollama(ctx context.Context, prompt string) (string, error) {
	payload := map[string]any{
		"model":  c.ollamaModel,
		"prompt": prompt,
		"stream": false,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		c.ollamaURL+"/api/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request: %w — is Ollama running? (ollama serve)", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama %d: %s", resp.StatusCode, raw)
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("parse ollama response: %w", err)
	}
	return result.Response, nil
}

// ── OpenAI ──────────────────────────────────────────────────────────────────

func (c *Client) openai(ctx context.Context, prompt string) (string, error) {
	payload := map[string]any{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": 1024,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai %d: %s", resp.StatusCode, raw)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &result); err != nil || len(result.Choices) == 0 {
		return "", fmt.Errorf("parse openai response: %w", err)
	}
	return result.Choices[0].Message.Content, nil
}

// ── Anthropic ────────────────────────────────────────────────────────────────

func (c *Client) anthropic(ctx context.Context, prompt string) (string, error) {
	payload := map[string]any{
		"model": "claude-haiku-4-5-20251001",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": 1024,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic request: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic %d: %s", resp.StatusCode, raw)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &result); err != nil || len(result.Content) == 0 {
		return "", fmt.Errorf("parse anthropic response: %w", err)
	}
	return result.Content[0].Text, nil
}
