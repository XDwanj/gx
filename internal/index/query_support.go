package index

import (
	"os"
	"path/filepath"

	"github.com/XDwanj/gx/internal/language"
)

type IgnoreMatcher struct {
	root   string
	states map[string]walkState
}

func NewIgnoreMatcher(root string) *IgnoreMatcher {
	cleanRoot := filepath.Clean(root)
	return &IgnoreMatcher{
		root: cleanRoot,
		states: map[string]walkState{
			cleanRoot: buildWalkState(cleanRoot, cleanRoot, walkState{}),
		},
	}
}

func (matcher *IgnoreMatcher) Matches(relPath string, isDir bool) bool {
	if matcher == nil {
		return false
	}

	cleanRelPath := filepath.Clean(relPath)
	state, ok := matcher.stateForPath(cleanRelPath)
	if !ok {
		return false
	}
	return state.matches(cleanRelPath, isDir)
}

func (matcher *IgnoreMatcher) stateForPath(relPath string) (walkState, bool) {
	dir := matcher.root
	if relPath != "." {
		dir = filepath.Join(matcher.root, filepath.Dir(relPath))
	}
	return matcher.stateForDir(dir)
}

func (matcher *IgnoreMatcher) stateForDir(dir string) (walkState, bool) {
	cleanDir := filepath.Clean(dir)
	if state, ok := matcher.states[cleanDir]; ok {
		return state, true
	}
	if cleanDir == matcher.root {
		return matcher.states[matcher.root], true
	}

	parentState, ok := matcher.stateForDir(filepath.Dir(cleanDir))
	if !ok {
		return walkState{}, false
	}

	info, err := os.Stat(cleanDir)
	if err != nil || !info.IsDir() {
		return walkState{}, false
	}

	state := buildWalkState(matcher.root, cleanDir, parentState)
	matcher.states[cleanDir] = state
	return state, true
}

func LoadFileData(root string, relPath string) (FileData, bool, error) {
	cleanRoot := filepath.Clean(root)
	cleanRelPath := filepath.Clean(relPath)
	absPath := filepath.Join(cleanRoot, cleanRelPath)

	info, err := os.Stat(absPath)
	if err != nil {
		return FileData{}, false, err
	}
	if info.IsDir() {
		return FileData{}, false, nil
	}

	languageName := language.DetectLanguage(absPath)
	if languageName == "" {
		return FileData{}, false, nil
	}

	source, err := os.ReadFile(absPath)
	if err != nil {
		return FileData{}, true, err
	}

	symbols, err := language.ParseAndExtract(languageName, source, absPath)
	if err != nil {
		return FileData{}, true, err
	}

	return FileData{
		Meta:    NewFileEntry(info.ModTime(), languageName),
		Symbols: toIndexSymbols(symbols),
	}, true, nil
}
