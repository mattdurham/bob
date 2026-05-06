package main

import (
	"os"
	"testing"
)

func TestLoadConfig_APIKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "test-key-123")
	t.Setenv("BOB_EXTENSIONS_DIR", "")
	t.Setenv("BOB_MODEL", "")
	t.Setenv("BOB_PROVIDER", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.AnthropicAPIKey != "test-key-123" {
		t.Errorf("AnthropicAPIKey = %q, want %q", cfg.AnthropicAPIKey, "test-key-123")
	}
}

func TestLoadConfig_ExtensionsDir(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "key")
	t.Setenv("BOB_EXTENSIONS_DIR", "/tmp/extensions")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.ExtensionsDir != "/tmp/extensions" {
		t.Errorf("ExtensionsDir = %q, want %q", cfg.ExtensionsDir, "/tmp/extensions")
	}
}

func TestLoadConfig_ModelDefault(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "key")
	t.Setenv("BOB_MODEL", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.Model != "claude-sonnet-4-5" {
		t.Errorf("Model = %q, want %q", cfg.Model, "claude-sonnet-4-5")
	}
}

func TestLoadConfig_ModelOverride(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "key")
	t.Setenv("BOB_MODEL", "claude-haiku-3-5")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.Model != "claude-haiku-3-5" {
		t.Errorf("Model = %q, want %q", cfg.Model, "claude-haiku-3-5")
	}
}

func TestLoadConfig_ProviderDefault(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "key")
	t.Setenv("BOB_PROVIDER", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.Provider != "anthropic" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "anthropic")
	}
}

func TestLoadConfig_ProviderOverride(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "key")
	t.Setenv("BOB_PROVIDER", "custom")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.Provider != "custom" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "custom")
	}
}

func TestLoadConfig_MissingAPIKeyError(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("BOB_PROVIDER", "anthropic")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for missing ANTHROPIC_API_KEY, got nil")
	}
}

func TestLoadConfig_MissingAPIKeyNonAnthropicOK(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("BOB_PROVIDER", "custom")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig with non-anthropic provider should not error on missing key: %v", err)
	}
	if cfg.AnthropicAPIKey != "" {
		t.Errorf("AnthropicAPIKey should be empty for custom provider: got %q", cfg.AnthropicAPIKey)
	}
}

func TestLoadConfig_AllEnvVars(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "mykey")
	t.Setenv("OPENAI_API_KEY", "oai-key")
	t.Setenv("GEMINI_API_KEY", "gem-key")
	t.Setenv("BOB_EXTENSIONS_DIR", "/ext")
	t.Setenv("BOB_MODEL", "claude-opus-4-5")
	t.Setenv("BOB_PROVIDER", "anthropic")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.AnthropicAPIKey != "mykey" {
		t.Errorf("AnthropicAPIKey = %q", cfg.AnthropicAPIKey)
	}
	if cfg.OpenAIAPIKey != "oai-key" {
		t.Errorf("OpenAIAPIKey = %q", cfg.OpenAIAPIKey)
	}
	if cfg.GeminiAPIKey != "gem-key" {
		t.Errorf("GeminiAPIKey = %q", cfg.GeminiAPIKey)
	}
	if cfg.ExtensionsDir != "/ext" {
		t.Errorf("ExtensionsDir = %q", cfg.ExtensionsDir)
	}
	if cfg.Model != "claude-opus-4-5" {
		t.Errorf("Model = %q", cfg.Model)
	}
	if cfg.Provider != "anthropic" {
		t.Errorf("Provider = %q", cfg.Provider)
	}
}

func TestLoadConfig_ExpandsHomeTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}
	t.Setenv("ANTHROPIC_API_KEY", "key")
	t.Setenv("BOB_EXTENSIONS_DIR", "~/myext")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	expected := home + "/myext"
	if cfg.ExtensionsDir != expected {
		t.Errorf("ExtensionsDir = %q, want %q", cfg.ExtensionsDir, expected)
	}
}

func TestLoadConfig_OpenAIAPIKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "oai-key-456")
	t.Setenv("BOB_PROVIDER", "openai")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.OpenAIAPIKey != "oai-key-456" {
		t.Errorf("OpenAIAPIKey = %q, want %q", cfg.OpenAIAPIKey, "oai-key-456")
	}
}

func TestLoadConfig_GeminiAPIKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "gem-key-789")
	t.Setenv("BOB_PROVIDER", "gemini")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.GeminiAPIKey != "gem-key-789" {
		t.Errorf("GeminiAPIKey = %q, want %q", cfg.GeminiAPIKey, "gem-key-789")
	}
}

func TestLoadConfig_MissingOpenAIKeyError(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("BOB_PROVIDER", "openai")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for missing OPENAI_API_KEY, got nil")
	}
}

func TestLoadConfig_MissingGeminiKeyError(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("BOB_PROVIDER", "gemini")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for missing GEMINI_API_KEY, got nil")
	}
}
