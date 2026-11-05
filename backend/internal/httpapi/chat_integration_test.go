package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"dailymeal/backend/internal/model"
)

func chatConfig(t *testing.T, agentURL string) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.Mkdir("config", 0700); err != nil {
		t.Fatal(err)
	}
	content := "default:\n  agent_url: " + agentURL + "\n  api_key: test-secret\n  base_url: http://model/v1\n  model: test-model\n  use_responses_api: true\n"
	if err := os.WriteFile("config/agent.yaml", []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestChatPersistsEachTurnWithoutLoadingHistory(t *testing.T) {
	f := setup(t)
	calls := 0
	agent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var input map[string]any
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Error(err)
		}
		if r.URL.Path != "/api/v1/chat" || r.Method != "POST" || len(input) != 2 {
			t.Errorf("unexpected request shape: %s %s", r.Method, r.URL.Path)
		}
		settings := input["model_settings"].(map[string]any)
		if settings["api_key"] != "test-secret" || settings["model"] != "test-model" || input["message"] != "查一下目标" {
			t.Error("Go did not forward the default model configuration and message")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"reply":"目标是 2000 大卡。","tool_results":[{"name":"get_goal","arguments":{},"result":{"energy_kcal":2000}}]}`))
	}))
	defer agent.Close()
	chatConfig(t, agent.URL)

	for range 2 {
		body := f.request(t, "POST", "/api/v1/chat", map[string]any{"message": "查一下目标"}, 200)
		if strings.Contains(string(body), "test-secret") || strings.Contains(string(body), "model_settings") {
			t.Fatal("response exposed model configuration")
		}
	}
	var records []model.ChatRequest
	if err := f.db.Order("id").Find(&records).Error; err != nil {
		t.Fatal(err)
	}
	if calls != 2 || len(records) != 2 || records[0].ID == records[1].ID {
		t.Fatal("each message should execute independently and create a record")
	}
	for _, record := range records {
		if record.Status != "completed" || record.Message != "查一下目标" || record.Reply != "目标是 2000 大卡。" || len(record.ToolResults) != 1 || record.Error != "" {
			t.Fatalf("incomplete persisted turn: %+v", record)
		}
	}
}

func TestFailedChatPersistsPartialToolResults(t *testing.T) {
	f := setup(t)
	agent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(502)
		_, _ = w.Write([]byte(`{"error":{"message":"upstream test-secret","tool_results":[{"name":"record_food","arguments":{},"result":{"id":12}}]}}`))
	}))
	defer agent.Close()
	chatConfig(t, agent.URL)
	body := f.request(t, "POST", "/api/v1/chat", map[string]any{"message": "记录午餐"}, 502)
	var record model.ChatRequest
	if err := f.db.First(&record).Error; err != nil {
		t.Fatal(err)
	}
	if record.Status != "failed" || len(record.ToolResults) != 1 || record.Reply != "" || record.Error == "" {
		t.Fatalf("failed turn was not saved: %+v", record)
	}
	if strings.Contains(string(body)+record.Error, "test-secret") {
		t.Fatal("upstream error exposed credentials")
	}
}
