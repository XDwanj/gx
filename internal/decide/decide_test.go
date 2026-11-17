package decide

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

const searchName = "Search"

func fakeProvider(t *testing.T, selected []string, hits *atomic.Int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		content, _ := json.Marshal(map[string][]string{"selected_ids": selected})
		response := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": string(content)}},
			},
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(response)
	}))
}

func testRequest() Request {
	return Request{
		Intent: Intent{
			Command: "symbols",
			Name:    searchName,
			Target:  Target{File: "a.go", Line: 1, Name: searchName},
		},
		Candidates: []Candidate{
			{ID: "c0", File: "a.go", Line: 1, Name: searchName},
			{ID: "c1", File: "b.go", Line: 5, Name: searchName},
		},
	}
}

func TestDecideSelectsAndCaches(t *testing.T) {
	var hits atomic.Int32
	server := fakeProvider(t, []string{"c0"}, &hits)
	defer server.Close()

	root := t.TempDir()
	client := NewClient(Config{APIKey: "k", BaseURL: server.URL, Model: "m"}, root)

	verdict, err := client.Decide(context.Background(), testRequest())
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if len(verdict.SelectedIDs) != 1 || verdict.SelectedIDs[0] != "c0" {
		t.Fatalf("selected %v", verdict.SelectedIDs)
	}
	if hits.Load() != 1 {
		t.Fatalf("provider hits = %d", hits.Load())
	}

	verdict, err = client.Decide(context.Background(), testRequest())
	if err != nil {
		t.Fatalf("Decide cached: %v", err)
	}
	if len(verdict.SelectedIDs) != 1 || hits.Load() != 1 {
		t.Fatalf("cache should answer the second call without a new request, hits=%d verdict=%v", hits.Load(), verdict)
	}
}

func TestDecideRejectsUnknownID(t *testing.T) {
	var hits atomic.Int32
	server := fakeProvider(t, []string{"c9"}, &hits)
	defer server.Close()

	client := NewClient(Config{APIKey: "k", BaseURL: server.URL, Model: "m"}, t.TempDir())

	_, err := client.Decide(context.Background(), testRequest())
	if err == nil || !strings.Contains(err.Error(), "unknown candidate id") {
		t.Fatalf("expected unknown-id error, got %v", err)
	}
}

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("GX_AI_API_KEY", "")
	t.Setenv("GX_AI_BASE_URL", "")
	if _, err := ConfigFromEnv(); err == nil {
		t.Fatal("missing credentials should fail loudly")
	}
	t.Setenv("GX_AI_API_KEY", "k")
	t.Setenv("GX_AI_BASE_URL", "https://openrouter.ai/api")
	t.Setenv("GX_AI_MODEL", "")
	config, err := ConfigFromEnv()
	if err != nil {
		t.Fatalf("ConfigFromEnv: %v", err)
	}
	if config.Model == "" || normalizeBaseURL(config.BaseURL) != "https://openrouter.ai/api/v1" {
		t.Fatalf("config %+v", config)
	}
}

func TestDecodeSelectedIDs(t *testing.T) {
	ids, err := decodeSelectedIDs(`{"selected_ids":["c0","c2"]}`)
	if err != nil || len(ids) != 2 {
		t.Fatalf("decode %v %v", ids, err)
	}

	ids, err = decodeSelectedIDs(fmt.Sprintf("noise\n%s\ntail", `{"selected_ids":["c1"]}`))
	if err != nil || len(ids) != 1 || ids[0] != "c1" {
		t.Fatalf("decode with noise %v %v", ids, err)
	}

	if _, err = decodeSelectedIDs("not json at all"); err == nil {
		t.Fatal("invalid response should error")
	}
}
