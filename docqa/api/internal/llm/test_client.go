package llm

import (
	"os"
	"testing"
)

func TestNewClient_DefaultsToOllama(t *testing.T) {
	os.Unsetenv("LLM_PROVIDER")
	c := NewClient()
	if c.provider != "ollama" {
		t.Errorf("expected default provider 'ollama', got %q", c.provider)
	}
}

func TestNewClient_OllamaDefaults(t *testing.T) {
	os.Setenv("LLM_PROVIDER", "ollama")
	os.Unsetenv("OLLAMA_URL")
	os.Unsetenv("OLLAMA_MODEL")
	defer os.Unsetenv("LLM_PROVIDER")

	c := NewClient()
	if c.ollamaURL != "http://host.docker.internal:11434" {
		t.Errorf("unexpected ollamaURL: %q", c.ollamaURL)
	}
	if c.ollamaModel != "llama3.2" {
		t.Errorf("unexpected ollamaModel: %q", c.ollamaModel)
	}
}

func TestNewClient_ReadsOpenAIKey(t *testing.T) {
	os.Setenv("LLM_PROVIDER", "openai")
	os.Setenv("OPENAI_API_KEY", "sk-test-123")
	defer func() {
		os.Unsetenv("LLM_PROVIDER")
		os.Unsetenv("OPENAI_API_KEY")
	}()

	c := NewClient()
	if c.provider != "openai" {
		t.Errorf("expected provider 'openai', got %q", c.provider)
	}
	if c.apiKey != "sk-test-123" {
		t.Errorf("expected apiKey to be set, got %q", c.apiKey)
	}
}

func TestNewClient_ReadsAnthropicKey(t *testing.T) {
	os.Setenv("LLM_PROVIDER", "anthropic")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test")
	defer func() {
		os.Unsetenv("LLM_PROVIDER")
		os.Unsetenv("ANTHROPIC_API_KEY")
	}()

	c := NewClient()
	if c.provider != "anthropic" {
		t.Errorf("expected provider 'anthropic', got %q", c.provider)
	}
	if c.apiKey != "sk-ant-test" {
		t.Errorf("expected apiKey to be set, got %q", c.apiKey)
	}
}

func TestNewClient_CustomOllamaURL(t *testing.T) {
	os.Setenv("LLM_PROVIDER", "ollama")
	os.Setenv("OLLAMA_URL", "http://my-ollama:11434")
	os.Setenv("OLLAMA_MODEL", "mistral")
	defer func() {
		os.Unsetenv("LLM_PROVIDER")
		os.Unsetenv("OLLAMA_URL")
		os.Unsetenv("OLLAMA_MODEL")
	}()

	c := NewClient()
	if c.ollamaURL != "http://my-ollama:11434" {
		t.Errorf("unexpected ollamaURL: %q", c.ollamaURL)
	}
	if c.ollamaModel != "mistral" {
		t.Errorf("unexpected ollamaModel: %q", c.ollamaModel)
	}
}
