package decide

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
	defaultModel        = "gpt-4o-mini"
	promptVersion       = 4
	requestTimeout      = 60 * time.Second
	concurrencyLimit    = 256
	providerCacheKeySep = "\x00"
)

type Config struct {
	APIKey  string
	BaseURL string
	Model   string
}

// Intent describes which symbol the user wants.
type Intent struct {
	Command string `json:"command"`
	Name    string `json:"name"`
	Target  Target `json:"target"`
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
	Signature string `json:"signature,omitempty"`
	Context   string `json:"context,omitempty"`
	Snippet   string `json:"snippet,omitempty"`
	Body      string `json:"body,omitempty"`
}

type Request struct {
	Intent     Intent      `json:"intent"`
	Candidates []Candidate `json:"candidates"`
}

type Verdict struct {
	SelectedIDs []string
}

type Decider interface {
	ProviderID() string
	Decide(ctx context.Context, req Request) (Verdict, error)
}

// Debugf receives cache and request progress; nil means silent.
type Debugf func(format string, args ...any)

type Client struct {
	config  Config
	root    string
	openai  openai.Client
	limiter chan struct{}
	debugf  Debugf
}

func ConfigFromEnv() (Config, error) {
	config := Config{
		APIKey:  strings.TrimSpace(os.Getenv("GX_AI_API_KEY")),
		BaseURL: strings.TrimSpace(os.Getenv("GX_AI_BASE_URL")),
		Model:   strings.TrimSpace(os.Getenv("GX_AI_MODEL")),
	}
	if config.APIKey == "" || config.BaseURL == "" {
		return Config{}, fmt.Errorf("gx: --define-in requires GX_AI_API_KEY and GX_AI_BASE_URL")
	}
	if config.Model == "" {
		config.Model = defaultModel
	}
	return config, nil
}

func NewClient(config Config, root string) *Client {
	baseURL := normalizeBaseURL(config.BaseURL)
	return &Client{
		config: config,
		root:   root,
		openai: openai.NewClient(
			option.WithAPIKey(config.APIKey),
			option.WithBaseURL(baseURL),
			option.WithRequestTimeout(requestTimeout),
		),
		limiter: make(chan struct{}, concurrencyLimit),
	}
}

func NewClientFromEnv(root string) (*Client, error) {
	config, err := ConfigFromEnv()
	if err != nil {
		return nil, err
	}
	return NewClient(config, root), nil
}

func (client *Client) SetDebugf(debugf Debugf) {
	client.debugf = debugf
}

func (client *Client) ProviderID() string {
	sum := sha256.Sum256([]byte(normalizeBaseURL(client.config.BaseURL) + providerCacheKeySep + client.config.Model))
	return hex.EncodeToString(sum[:])
}

func (client *Client) Decide(ctx context.Context, request Request) (Verdict, error) {
	providerID := client.ProviderID()
	key, err := CacheKey(providerID, request)
	if err != nil {
		return Verdict{}, err
	}
	cache, err := OpenCache(client.root)
	if err != nil {
		return Verdict{}, err
	}
	defer func() {
		_ = cache.Close()
	}()

	selectedIDs, found, err := cache.Get(key)
	if err != nil {
		return Verdict{}, err
	}
	if !found {
		client.logf("ai cache miss command=%s name=%s target=%s candidates=%d",
			request.Intent.Command, request.Intent.Name, request.Intent.Target.Name,
			len(request.Candidates))
		client.limiter <- struct{}{}
		client.logf("ai request start command=%s name=%s target=%s candidates=%d",
			request.Intent.Command, request.Intent.Name, request.Intent.Target.Name,
			len(request.Candidates))
		defer func() {
			<-client.limiter
		}()
		selectedIDs, err = client.selectIDs(ctx, request)
		if err != nil {
			return Verdict{}, err
		}
		client.logf("ai request done command=%s name=%s target=%s selected=%d",
			request.Intent.Command, request.Intent.Name, request.Intent.Target.Name,
			len(selectedIDs))
		if err := cache.Put(key, providerID, selectedIDs); err != nil {
			return Verdict{}, err
		}
	} else {
		client.logf("ai cache hit command=%s name=%s target=%s candidates=%d selected=%d",
			request.Intent.Command, request.Intent.Name, request.Intent.Target.Name,
			len(request.Candidates), len(selectedIDs))
	}

	if err := validateSelectedIDs(request.Candidates, selectedIDs); err != nil {
		return Verdict{}, err
	}
	return Verdict{SelectedIDs: selectedIDs}, nil
}

func (client *Client) selectIDs(ctx context.Context, request Request) ([]string, error) {
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

func (client *Client) logf(format string, args ...any) {
	if client.debugf == nil {
		return
	}
	client.debugf(format, args...)
}

func CacheKey(providerID string, request Request) (string, error) {
	encoded, err := json.Marshal(struct {
		PromptVersion int     `json:"prompt_version"`
		ProviderID    string  `json:"provider_id"`
		Request       Request `json:"request"`
	}{
		PromptVersion: promptVersion,
		ProviderID:    providerID,
		Request:       request,
	})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func selectionPrompt(request Request) string {
	encoded, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		encoded = []byte("{}")
	}
	return "Given this JSON request, return {\"selected_ids\":[...]}.\n" +
		"Use intent.target as the intended symbol. For symbols and definition, select candidates that are the same declaration as the target. " +
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

func validateSelectedIDs(candidates []Candidate, selectedIDs []string) error {
	valid := make(map[string]bool, len(candidates))
	for _, candidate := range candidates {
		valid[candidate.ID] = true
	}
	for _, id := range selectedIDs {
		if !valid[id] {
			return fmt.Errorf("gx: AI selected unknown candidate id: %s", id)
		}
	}
	return nil
}
