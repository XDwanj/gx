package lang

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const upstreamIssueURL = "https://github.com/ind-igo/cx/issues/new?template=language-request.yml"

var supportedLanguages = []string{
	"bash",
	"c",
	"cpp",
	"go",
	"java",
	"lua",
	"python",
	"protobuf",
	"ruby",
	"rust",
	"swift",
	"typescript",
	"zig",
}

type manifest struct {
	Installed map[string]bool `json:"installed"`
}

func Add(_ io.Writer, stderr io.Writer, languages []string) error {
	if len(languages) == 0 {
		return fmt.Errorf("gx: specify at least one language, e.g.: gx lang add rust typescript")
	}
	state, err := loadManifest()
	if err != nil {
		return err
	}

	for _, language := range languages {
		if !isSupported(language) {
			return fmt.Errorf("gx: unknown language '%s' — supported: %s", language, joinSupported())
		}
		state.Installed[language] = true
	}
	if saveErr := saveManifest(state); saveErr != nil {
		return saveErr
	}
	_, err = fmt.Fprintf(stderr, "gx: installed %d grammar(s)\n", len(languages))
	return err
}

func Remove(_ io.Writer, stderr io.Writer, languages []string) error {
	if len(languages) == 0 {
		return fmt.Errorf("gx: specify at least one language, e.g.: gx lang remove rust")
	}
	state, err := loadManifest()
	if err != nil {
		return err
	}

	for _, language := range languages {
		if state.Installed[language] {
			delete(state.Installed, language)
			if _, writeErr := fmt.Fprintf(stderr, "gx: removed %s grammar\n", language); writeErr != nil {
				return writeErr
			}
			continue
		}
		if _, writeErr := fmt.Fprintf(stderr, "gx: %s grammar not found in cache\n", language); writeErr != nil {
			return writeErr
		}
	}
	return saveManifest(state)
}

func List(stdout io.Writer, stderr io.Writer) error {
	state, err := loadManifest()
	if err != nil {
		return err
	}

	for _, language := range supportedLanguages {
		marker := "[missing]"
		if state.Installed[language] {
			marker = "[installed]"
		}
		if _, writeErr := fmt.Fprintf(stdout, "%-15s %s\n", language, marker); writeErr != nil {
			return writeErr
		}
	}
	_, err = fmt.Fprintf(
		stderr,
		"\nNeed another language? gx is derived from cx; upstream issue tracker: %s\n",
		upstreamIssueURL,
	)
	return err
}

func GrammarCacheDir() string {
	return filepath.Join(gxCacheDir(), "grammars")
}

func IsInstalled(language string) bool {
	state, err := loadManifest()
	if err != nil {
		return false
	}
	return state.Installed[language]
}

func SupportedLanguages() []string {
	result := make([]string, len(supportedLanguages))
	copy(result, supportedLanguages)
	return result
}

func gxCacheDir() string {
	base, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join(".cache", "gx")
	}
	return filepath.Join(base, "gx")
}

func loadManifest() (*manifest, error) {
	path := filepath.Join(GrammarCacheDir(), "manifest.json")
	state := &manifest{Installed: map[string]bool{}}
	bytes, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(bytes, state); err != nil {
		return nil, err
	}
	if state.Installed == nil {
		state.Installed = map[string]bool{}
	}
	return state, nil
}

func saveManifest(state *manifest) error {
	if err := os.MkdirAll(GrammarCacheDir(), 0o755); err != nil {
		return err
	}
	bytes, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(GrammarCacheDir(), "manifest.json"), bytes, 0o644)
}

func isSupported(language string) bool {
	for _, item := range supportedLanguages {
		if item == language {
			return true
		}
	}
	return false
}

func joinSupported() string {
	items := append([]string(nil), supportedLanguages...)
	sort.Strings(items)
	return strings.Join(items, ", ")
}
