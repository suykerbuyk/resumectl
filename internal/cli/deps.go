package cli

import (
	"fmt"

	"github.com/jsuykerbuyk/resumectl/internal/llm"
	"github.com/jsuykerbuyk/resumectl/internal/prompts"
	"github.com/jsuykerbuyk/resumectl/internal/vault"
)

// Dependencies holds injectable dependencies for CLI commands.
// In production, these are built from config; in tests, they are mocked.
type Dependencies struct {
	Vault   *vault.Vault
	Router  *llm.Router
	Prompts *prompts.Library
}

// deps is the global dependency holder, set by initDeps or tests.
var deps *Dependencies

// initDeps creates the real dependencies from the loaded config.
func initDeps() error {
	if deps != nil {
		return nil
	}

	v, err := vault.Open(cfg.Vault.Path, cfg)
	if err != nil {
		return fmt.Errorf("open vault: %w", err)
	}

	lib, err := prompts.NewLibrary()
	if err != nil {
		return fmt.Errorf("load prompts: %w", err)
	}

	providers := make(map[string]llm.Provider)

	if cfg.Providers.Claude.APIKey != "" {
		providers["claude"] = llm.NewClaudeProvider(
			cfg.Providers.Claude.APIKey,
			cfg.Providers.Claude.DefaultModel,
		)
	}
	if cfg.Providers.OpenAI.APIKey != "" {
		providers["openai"] = llm.NewOpenAIProvider(
			cfg.Providers.OpenAI.APIKey,
			cfg.Providers.OpenAI.DefaultModel,
		)
	}
	if cfg.Providers.Gemini.APIKey != "" {
		providers["gemini"] = llm.NewGeminiProvider(
			cfg.Providers.Gemini.APIKey,
			cfg.Providers.Gemini.DefaultModel,
		)
	}
	if cfg.Providers.Grok.APIKey != "" {
		p := llm.NewGrokProvider(
			cfg.Providers.Grok.APIKey,
			cfg.Providers.Grok.DefaultModel,
		)
		if cfg.Providers.Grok.BaseURL != "" {
			p.BaseURL = cfg.Providers.Grok.BaseURL
		}
		providers["grok"] = p
	}

	router := llm.NewRouter(providers, cfg.Providers.DefaultForTask)

	deps = &Dependencies{
		Vault:   v,
		Router:  router,
		Prompts: lib,
	}

	return nil
}

// getProvider resolves the provider for the given task type,
// respecting the --provider CLI override.
func getProvider(task llm.TaskType) (llm.Provider, error) {
	if providerOvr != "" {
		p, ok := deps.Router.Providers[providerOvr]
		if !ok {
			return nil, fmt.Errorf("unknown provider %q", providerOvr)
		}
		return p, nil
	}
	return deps.Router.ForTask(task)
}
