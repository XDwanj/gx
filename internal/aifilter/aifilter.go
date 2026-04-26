package aifilter

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const (
	promptVersion = 2
	defaultModel  = "gpt-4o-mini"
)

type Config struct {
	APIKey  string
	BaseURL string
	Model   string
}

type Selector interface {
	ProviderID() string
	Select(ctx context.Context, request SelectionRequest) ([]string, error)
}

type SelectionRequest struct {
	Command    string      `json:"command"`
	Name       string      `json:"name"`
	Target     Target      `json:"target"`
	Candidates []Candidate `json:"candidates"`
}

type Target struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	Name      string `json:"name"`
	Kind      string `json:"kind,omitempty"`
	Signature string `json:"signature,omitempty"`
	Body      string `json:"body,omitempty"`
}

type Candidate struct {
	ID        string `json:"id"`
	File      string `json:"file"`
	Line      int    `json:"line"`
	Name      string `json:"name,omitempty"`
	Kind      string `json:"kind,omitempty"`
	Caller    string `json:"caller,omitempty"`
	Callee    string `json:"callee,omitempty"`
	Signature string `json:"signature,omitempty"`
	Context   string `json:"context,omitempty"`
	Snippet   string `json:"snippet,omitempty"`
	Body      string `json:"body,omitempty"`
}

type Client struct {
	config Config
	openai openai.Client
}

func ConfigFromEnv() (Config, error) {
	config := Config{
		APIKey:  strings.TrimSpace(os.Getenv("GX_OPENAI_API_KEY")),
		BaseURL: strings.TrimSpace(os.Getenv("GX_OPENAI_BASE_URL")),
		Model:   strings.TrimSpace(os.Getenv("GX_OPENAI_MODEL")),
	}
	if config.APIKey == "" || config.BaseURL == "" {
		return Config{}, fmt.Errorf("gx: --define-in requires GX_OPENAI_API_KEY and GX_OPENAI_BASE_URL")
	}
	if config.Model == "" {
		config.Model = defaultModel
	}
	return config, nil
}

func NewClient(config Config) *Client {
	baseURL := normalizeBaseURL(config.BaseURL)
	return &Client{
		config: config,
		openai: openai.NewClient(
			option.WithAPIKey(config.APIKey),
			option.WithBaseURL(baseURL),
			option.WithRequestTimeout(60*time.Second),
		),
	}
}

func NewClientFromEnv() (*Client, error) {
	config, err := ConfigFromEnv()
	if err != nil {
		return nil, err
	}
	return NewClient(config), nil
}

func (client *Client) ProviderID() string {
	sum := sha256.Sum256([]byte(normalizeBaseURL(client.config.BaseURL) + "\x00" + client.config.Model))
	return hex.EncodeToString(sum[:])
}

func (client *Client) Select(ctx context.Context, request SelectionRequest) ([]string, error) {
	response, err := client.openai.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: client.config.Model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You disambiguate code navigation candidates. Return only JSON with selected_ids. " +
				"Use command-specific rules and preserve candidate IDs exactly."),
			openai.UserMessage(selectionPrompt(request)),
		},
		Temperature: openai.Float(0),
	})
	if err != nil {
		return nil, err
	}
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("gx: AI response did not include choices")
	}
	return decodeSelectedIDs(response.Choices[0].Message.Content)
}

func CacheKey(providerID string, request SelectionRequest) (string, error) {
	payload := struct {
		PromptVersion int              `json:"prompt_version"`
		ProviderID    string           `json:"provider_id"`
		Request       SelectionRequest `json:"request"`
	}{
		PromptVersion: promptVersion,
		ProviderID:    providerID,
		Request:       request,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func selectionPrompt(request SelectionRequest) string {
	encoded, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		encoded = []byte("{}")
	}
	return "Given this JSON request, return {\"selected_ids\":[...]}.\n" +
		"Use the target as the intended symbol. For symbols, definition, and callees, select candidates that are the same declaration as the target. " +
		"For callees, each candidate is the caller definition whose calls will be listed. " +
		"For references, select only candidates that refer to the target, including target definition sites when they are present. " +
		"When the target is a method, do not select the same method name on another receiver, class, struct, object, or type. " +
		"Preserve candidate IDs exactly. " +
		"If no candidate matches, return an empty array.\n\n" +
		string(encoded)
}

func normalizeBaseURL(baseURL string) string {
	trimmed := strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(trimmed, "/v1") {
		return trimmed
	}
	return trimmed + "/v1"
}

func decodeSelectedIDs(content string) ([]string, error) {
	var decoded struct {
		SelectedIDs []string `json:"selected_ids"`
	}
	if err := json.Unmarshal([]byte(content), &decoded); err == nil {
		return decoded.SelectedIDs, nil
	}

	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start < 0 || end < start {
		return nil, fmt.Errorf("gx: AI response was not JSON: %s", strings.TrimSpace(content))
	}
	if err := json.Unmarshal([]byte(content[start:end+1]), &decoded); err != nil {
		return nil, fmt.Errorf("gx: AI response was not valid selected_ids JSON: %w", err)
	}
	return decoded.SelectedIDs, nil
}
