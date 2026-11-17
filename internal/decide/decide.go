package decide

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	systemOnePath       = "/api/alpha/decisions"
	systemOneThreshold  = 0.5
	errorBodyLimit      = 500
)

var systemOneClient = &http.Client{Timeout: requestTimeout}

type Config struct {
	APIKey  string
	BaseURL string
	Model   string
}

// Intent 描述用户要找的符号。
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

// Debugf 接收缓存与请求进度；为 nil 表示静默。
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
	if client.usesSystemOne() {
		return client.selectIDsSystemOne(ctx, request)
	}
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

// usesSystemOne 判断配置的模型是否为 TypeSafe System One 模型（Jev），
// 它在 OpenRouter 上走 decisions API，而不是 chat completions。
func (client *Client) usesSystemOne() bool {
	return strings.HasPrefix(client.config.Model, "typesafe/") ||
		strings.HasPrefix(client.config.Model, "jev-")
}

func (client *Client) selectIDsSystemOne(ctx context.Context, request Request) ([]string, error) {
	questions := make(map[string]any, len(request.Candidates))
	for i, candidate := range request.Candidates {
		questions[systemOneQuestionKey(i)] = map[string]any{
			"type": "noul",
			"instructions": fmt.Sprintf(
				"For command %q, does the candidate with id %q at %s:%d (%s) match the target %q? %s",
				request.Intent.Command, candidate.ID, candidate.File, candidate.Line,
				candidate.Signature, request.Intent.Target.Name, systemOneRule(request.Intent.Command)),
			"criteria": map[string]string{
				"true":  "the candidate matches the target under the stated rule",
				"false": "the candidate is a different declaration or an unrelated reference",
			},
		}
	}
	payload, err := json.Marshal(map[string]any{
		"model":     client.config.Model,
		"state":     request,
		"questions": questions,
	})
	if err != nil {
		return nil, err
	}
	endpoint, err := client.systemOneURL()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+client.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := systemOneClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gx: system one decision failed: %s: %s", resp.Status, truncateForError(raw))
	}
	var decoded struct {
		Answers map[string]struct {
			Noul float64 `json:"noul"`
		} `json:"answers"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, fmt.Errorf("gx: system one response was not JSON: %w", err)
	}
	selected := make([]string, 0, len(request.Candidates))
	for i, candidate := range request.Candidates {
		if decoded.Answers[systemOneQuestionKey(i)].Noul >= systemOneThreshold {
			selected = append(selected, candidate.ID)
		}
	}
	return selected, nil
}

func (client *Client) systemOneURL() (string, error) {
	parsed, err := url.Parse(normalizeBaseURL(client.config.BaseURL))
	if err != nil || parsed.Host == "" {
		return "", fmt.Errorf("gx: invalid GX_AI_BASE_URL %q", client.config.BaseURL)
	}
	parsed.Path = systemOnePath
	parsed.RawQuery, parsed.Fragment = "", ""
	return parsed.String(), nil
}

func systemOneQuestionKey(index int) string {
	return fmt.Sprintf("q%d", index)
}

func systemOneRule(command string) string {
	if command == "references" {
		return "Select candidates that refer to the target, including the target " +
			"definition site when present. Do not match the same method name on " +
			"another receiver, class, struct, object, or type."
	}
	return "Select candidates that are the same declaration as the target. When " +
		"the target is a method, do not match the same method name on another " +
		"receiver, class, struct, object, or type."
}

func truncateForError(raw []byte) string {
	text := strings.TrimSpace(string(raw))
	if len(text) > errorBodyLimit {
		return text[:errorBodyLimit]
	}
	return text
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
