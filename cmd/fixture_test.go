package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/XDwanj/gx/internal/app"
	"github.com/XDwanj/gx/internal/lang"
)

const (
	fixtureCommandSymbol     = "symbol"
	fixtureCommandDefinition = "definition"
)

type fixtureQuery struct {
	Name     string   `json:"name,omitempty"`
	Kind     string   `json:"kind,omitempty"`
	Paths    []string `json:"paths,omitempty"`
	MaxLines int      `json:"max_lines,omitempty"`
}

type fixtureCase struct {
	Language string
	Command  string
	Name     string
	Dir      string
}

type fixtureSymbolRow struct {
	File      string `json:"file,omitempty"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Signature string `json:"signature"`
}

func TestLanguageFixtures(t *testing.T) {
	fixturesRoot := filepath.Join(repoRootFromFixtureTest(t), "tests")
	cases, err := discoverFixtureCases(fixturesRoot)
	if err != nil {
		t.Fatalf("discover fixture cases: %v", err)
	}
	if len(cases) == 0 {
		t.Fatalf("expected fixture cases under %s", fixturesRoot)
	}

	for _, testCase := range cases {
		testCase := testCase
		t.Run(fmt.Sprintf("%s/%s/%s", testCase.Language, testCase.Command, testCase.Name), func(t *testing.T) {
			runFixtureCase(t, testCase)
		})
	}
}

func TestFixtureKindCoverage(t *testing.T) {
	fixturesRoot := filepath.Join(repoRootFromFixtureTest(t), "tests")
	cases, err := discoverFixtureCases(fixturesRoot)
	if err != nil {
		t.Fatalf("discover fixture cases: %v", err)
	}

	symbolKinds := map[string]bool{}
	definitionKinds := map[string]bool{}

	for _, testCase := range cases {
		query := readFixtureJSON[fixtureQuery](t, filepath.Join(testCase.Dir, "query.json"))
		if testCase.Command == fixtureCommandDefinition && query.Kind != "" {
			definitionKinds[query.Kind] = true
		}
		if testCase.Command != fixtureCommandSymbol {
			continue
		}

		expected := readFixtureJSON[[]fixtureSymbolRow](t, filepath.Join(testCase.Dir, "expected.json"))
		for _, row := range expected {
			symbolKinds[row.Kind] = true
		}
	}

	for _, kind := range publicKinds() {
		if !symbolKinds[string(kind)] {
			t.Fatalf("fixture symbol coverage is missing kind %q", kind)
		}
		if !definitionKinds[string(kind)] {
			t.Fatalf("fixture definition coverage is missing kind %q", kind)
		}
	}
}

func TestFixtureLanguageKindCoverage(t *testing.T) {
	fixturesRoot := filepath.Join(repoRootFromFixtureTest(t), "tests")
	cases, err := discoverFixtureCases(fixturesRoot)
	if err != nil {
		t.Fatalf("discover fixture cases: %v", err)
	}

	symbolKindsByLanguage := map[string]map[string]bool{}
	definitionKindsByLanguage := map[string]map[string]bool{}

	for _, testCase := range cases {
		if _, ok := symbolKindsByLanguage[testCase.Language]; !ok {
			symbolKindsByLanguage[testCase.Language] = map[string]bool{}
		}
		if _, ok := definitionKindsByLanguage[testCase.Language]; !ok {
			definitionKindsByLanguage[testCase.Language] = map[string]bool{}
		}

		query := readFixtureJSON[fixtureQuery](t, filepath.Join(testCase.Dir, "query.json"))
		if testCase.Command == fixtureCommandDefinition && query.Kind != "" {
			definitionKindsByLanguage[testCase.Language][query.Kind] = true
		}
		if testCase.Command != fixtureCommandSymbol {
			continue
		}

		expected := readFixtureJSON[[]fixtureSymbolRow](t, filepath.Join(testCase.Dir, "expected.json"))
		for _, row := range expected {
			symbolKindsByLanguage[testCase.Language][row.Kind] = true
		}
	}

	for _, support := range languageKindSupportMatrix() {
		for _, kind := range support.Kinds {
			if !symbolKindsByLanguage[support.Language][string(kind)] {
				t.Fatalf("fixture symbol coverage is missing language=%q kind=%q", support.Language, kind)
			}
			if !definitionKindsByLanguage[support.Language][string(kind)] {
				t.Fatalf("fixture definition coverage is missing language=%q kind=%q", support.Language, kind)
			}
		}
	}
}

func discoverFixtureCases(fixturesRoot string) ([]fixtureCase, error) {
	entries := make([]fixtureCase, 0)
	if _, err := os.Stat(fixturesRoot); err != nil {
		return nil, err
	}

	err := filepath.WalkDir(fixturesRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(fixturesRoot, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		parts := strings.Split(relPath, string(filepath.Separator))
		if len(parts) != 3 {
			return nil
		}

		if parts[1] != fixtureCommandSymbol && parts[1] != fixtureCommandDefinition {
			return nil
		}

		if !fixtureCaseFilesExist(path) {
			return fmt.Errorf("fixture case %s is missing query.json, expected.json, or project/", path)
		}

		entries = append(entries, fixtureCase{
			Language: parts[0],
			Command:  parts[1],
			Name:     parts[2],
			Dir:      path,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(left int, right int) bool {
		if entries[left].Language != entries[right].Language {
			return entries[left].Language < entries[right].Language
		}
		if entries[left].Command != entries[right].Command {
			return entries[left].Command < entries[right].Command
		}
		return entries[left].Name < entries[right].Name
	})
	return entries, nil
}

func fixtureCaseFilesExist(caseDir string) bool {
	requiredPaths := []string{
		filepath.Join(caseDir, "query.json"),
		filepath.Join(caseDir, "expected.json"),
		filepath.Join(caseDir, "_project"),
	}

	for _, requiredPath := range requiredPaths {
		if _, err := os.Stat(requiredPath); err != nil {
			return false
		}
	}
	return true
}

func repoRootFromFixtureTest(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Dir(filepath.Dir(file))
}

func runFixtureCase(t *testing.T, testCase fixtureCase) {
	t.Helper()

	if err := lang.Add(io.Discard, io.Discard, []string{testCase.Language}); err != nil {
		t.Fatalf("install grammar %s: %v", testCase.Language, err)
	}

	query := readFixtureJSON[fixtureQuery](t, filepath.Join(testCase.Dir, "query.json"))
	expected := readFixtureJSON[any](t, filepath.Join(testCase.Dir, "expected.json"))

	projectRoot := fixtureProjectRoot(t, testCase.Dir)
	t.Chdir(projectRoot)

	args := buildFixtureArgs(projectRoot, testCase.Command, query)
	stdout, stderr, exitCode := executeRootFixtureCommand(t, projectRoot, args...)
	if exitCode != 0 {
		t.Fatalf("fixture %s exited with %d: stderr=%q", testCase.Dir, exitCode, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("fixture %s wrote stderr: %q", testCase.Dir, stderr)
	}

	var actual any
	if err := json.Unmarshal([]byte(stdout), &actual); err != nil {
		t.Fatalf("unmarshal actual stdout: %v\nstdout=%s", err, stdout)
	}
	if !reflect.DeepEqual(actual, expected) {
		expectedBytes, marshalErr := json.MarshalIndent(expected, "", "  ")
		if marshalErr != nil {
			t.Fatalf("marshal expected JSON: %v", marshalErr)
		}
		t.Fatalf("unexpected output for %s\nactual: %s\nexpected: %s", testCase.Dir, stdout, string(expectedBytes))
	}
}

func fixtureProjectRoot(t *testing.T, caseDir string) string {
	t.Helper()

	sourceRoot := filepath.Join(caseDir, "_project")
	tempRoot := t.TempDir()
	if err := os.Mkdir(filepath.Join(tempRoot, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}

	err := filepath.WalkDir(sourceRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relPath, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		targetPath := filepath.Join(tempRoot, relPath)
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		bytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(targetPath, bytes, 0o644)
	})
	if err != nil {
		t.Fatalf("copy fixture project: %v", err)
	}
	return tempRoot
}

func buildFixtureArgs(projectRoot string, command string, query fixtureQuery) []string {
	args := []string{"--root", projectRoot, "--json"}
	switch command {
	case fixtureCommandSymbol:
		args = append(args, "symbols")
	case fixtureCommandDefinition:
		args = append(args, "definition")
	default:
		panic("unsupported fixture command: " + command)
	}

	if query.Name != "" {
		args = append(args, "--name", query.Name)
	}
	if query.Kind != "" {
		args = append(args, "--kind", query.Kind)
	}
	if command == fixtureCommandDefinition && query.MaxLines > 0 {
		args = append(args, "--max-lines", fmt.Sprintf("%d", query.MaxLines))
	}
	if len(query.Paths) > 0 {
		args = append(args, query.Paths...)
	}
	return args
}

func executeRootFixtureCommand(t *testing.T, projectRoot string, args ...string) (string, string, int) {
	t.Helper()

	previousCmd := rootCmd
	previousFlags := rootFlags
	previousShowVersion := showVersion
	rootCmd = newRootCmd()
	rootFlags = app.Flags{Root: projectRoot, JSON: true}
	showVersion = false
	t.Cleanup(func() {
		rootCmd = previousCmd
		rootFlags = previousFlags
		showVersion = previousShowVersion
	})

	rootCmd.SetArgs(args)

	var exitCode int
	stdout, stderr := captureFixtureProcessOutput(t, func() {
		exitCode = Execute()
	})
	return stdout, strings.TrimSpace(stderr), exitCode
}

func readFixtureJSON[T any](t *testing.T, path string) T {
	t.Helper()

	var value T
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(bytes, &value); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return value
}

func captureFixtureProcessOutput(t *testing.T, run func()) (string, string) {
	t.Helper()

	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}

	previousStdout := os.Stdout
	previousStderr := os.Stderr
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter
	t.Cleanup(func() {
		os.Stdout = previousStdout
		os.Stderr = previousStderr
	})

	run()

	if err := stdoutWriter.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	if err := stderrWriter.Close(); err != nil {
		t.Fatalf("close stderr writer: %v", err)
	}

	var stdout bytes.Buffer
	if _, err := stdout.ReadFrom(stdoutReader); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	var stderr bytes.Buffer
	if _, err := stderr.ReadFrom(stderrReader); err != nil {
		t.Fatalf("read stderr: %v", err)
	}
	return stdout.String(), stderr.String()
}
