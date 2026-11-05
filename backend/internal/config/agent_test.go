package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultAgentConfig(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.Mkdir("config", 0700); err != nil {
		t.Fatal(err)
	}
	data := []byte("default:\n  agent_url: http://agent:8000\n  api_key: local-test-key\n  base_url: http://model/v1\n  model: default-model\n  use_responses_api: false\nother:\n  model: unused-model\n")
	if err := os.WriteFile(filepath.FromSlash(AgentConfigPath), data, 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadAgent()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Model != "default-model" || cfg.APIKey != "local-test-key" || cfg.AgentURL != "http://agent:8000" || cfg.UseResponsesAPI {
		t.Fatal("did not load the default configuration")
	}
	if err := os.WriteFile(AgentConfigPath, []byte("default: [invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAgent(); err == nil {
		t.Fatal("invalid YAML must return an error")
	}
}
