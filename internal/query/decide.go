package query

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/XDwanj/gx/internal/decide"
	"github.com/XDwanj/gx/internal/index"
)

func (service *Service) deciderFor() (decide.Decider, error) {
	if service.decider != nil {
		return service.decider, nil
	}
	client, err := decide.NewClientFromEnv(service.root)
	if err != nil {
		return nil, err
	}
	client.SetDebugf(service.debugf)
	service.decider = client
	return client, nil
}

func validateDecideIntent(intent Intent) error {
	if intent.DefinedIn != "" && intent.Name == "" {
		return fmt.Errorf("gx: --define-in requires --name")
	}
	return nil
}

func (service *Service) resolveTarget(idx *index.Index, intent Intent) (decide.Target, error) {
	relPath := normalizeRelativePath(intent.DefinedIn, idx.Root)
	data, ok := idx.Entries[relPath]
	if !ok {
		return decide.Target{}, service.fileLookupError(relPath, idx.Root)
	}

	matches := make([]index.Symbol, 0)
	for _, symbol := range data.Symbols {
		matched, err := globMatch(intent.Name, symbol.Name)
		if err != nil {
			return decide.Target{}, err
		}
		if !matched {
			continue
		}
		if intent.Kind != nil && symbol.Kind != *intent.Kind {
			continue
		}
		matches = append(matches, symbol)
	}
	if len(matches) == 0 {
		return decide.Target{}, fmt.Errorf("gx: --define-in %s has no symbol matching %q", displayScopePath(relPath), intent.Name)
	}
	if len(matches) > 1 {
		return decide.Target{}, fmt.Errorf("gx: --define-in %s matches multiple symbols for %q; narrow --name", displayScopePath(relPath), intent.Name)
	}

	body, line, ok := readBody(idx.Root, relPath, matches[0])
	if !ok {
		body = ""
		line = matches[0].Line
	}
	return decide.Target{
		File:      displayPath(relPath),
		Line:      line,
		Name:      matches[0].Name,
		Kind:      string(matches[0].Kind),
		Signature: matches[0].Signature,
		Body:      truncatePromptText(body, aiTargetBodyMaxRunes),
	}, nil
}

// collectSiblings 收集与 --name 匹配的全部声明作为 AI 判定的对照组。
// 结果按文件与行号排序，保证 Request 序列化稳定、decide 缓存键可复用。
func (service *Service) collectSiblings(idx *index.Index, intent Intent) ([]decide.Target, error) {
	paths := make([]string, 0, len(idx.Entries))
	for relPath := range idx.Entries {
		paths = append(paths, relPath)
	}
	sort.Strings(paths)

	siblings := make([]decide.Target, 0)
	for _, relPath := range paths {
		for _, symbol := range idx.Entries[relPath].Symbols {
			matched, err := globMatch(intent.Name, symbol.Name)
			if err != nil {
				return nil, err
			}
			if !matched {
				continue
			}
			if intent.Kind != nil && symbol.Kind != *intent.Kind {
				continue
			}
			siblings = append(siblings, decide.Target{
				File:      displayPath(relPath),
				Line:      symbol.Line,
				Name:      symbol.Name,
				Kind:      string(symbol.Kind),
				Signature: symbol.Signature,
			})
			if len(siblings) >= aiMaxSiblings {
				return siblings, nil
			}
		}
	}
	return siblings, nil
}

func (service *Service) selectCandidateIDs(intent decide.Intent, candidates []decide.Candidate) (map[string]bool, error) {
	decider, err := service.deciderFor()
	if err != nil {
		return nil, err
	}
	verdict, err := decider.Decide(context.Background(), decide.Request{
		Intent:     intent,
		Candidates: candidates,
	})
	if err != nil {
		return nil, err
	}
	selected := make(map[string]bool, len(verdict.SelectedIDs))
	for _, id := range verdict.SelectedIDs {
		selected[id] = true
	}
	return selected, nil
}

func filterBySelection[Row any](rows []Row, selected map[string]bool) []Row {
	filtered := make([]Row, 0, len(rows))
	for indexValue, row := range rows {
		if selected[candidateID(indexValue)] {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func (service *Service) filterSymbolRowsWithDecide(idx *index.Index, name string, target decide.Target, siblings []decide.Target, rows []SymbolRow) ([]SymbolRow, error) {
	candidates := make([]decide.Candidate, 0, len(rows))
	for indexValue, row := range rows {
		snippet, err := snippetForLine(idx.Root, row.File, row.Line)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, decide.Candidate{
			ID:        candidateID(indexValue),
			File:      row.File,
			Line:      row.Line,
			Name:      row.Name,
			Kind:      row.Kind,
			Signature: row.Signature,
			Snippet:   snippet,
		})
	}

	selected, err := service.selectCandidateIDs(decide.Intent{Command: "symbols", Name: name, Target: target, Siblings: siblings}, candidates)
	if err != nil {
		return nil, err
	}
	return filterBySelection(rows, selected), nil
}

func (service *Service) filterDefsWithDecide(
	idx *index.Index, name string, target decide.Target, siblings []decide.Target, matches []definitionMatch,
) ([]definitionMatch, error) {
	candidates := make([]decide.Candidate, 0, len(matches))
	for indexValue, match := range matches {
		body, line, ok := readBody(idx.Root, match.path, match.symbol)
		if !ok {
			body = ""
			line = match.symbol.Line
		}
		candidates = append(candidates, decide.Candidate{
			ID:        candidateID(indexValue),
			File:      displayPath(match.path),
			Line:      line,
			Name:      match.symbol.Name,
			Kind:      string(match.symbol.Kind),
			Signature: match.symbol.Signature,
			Body:      truncatePromptText(body, aiCandidateBodyMaxRunes),
		})
	}

	selected, err := service.selectCandidateIDs(decide.Intent{Command: "definition", Name: name, Target: target, Siblings: siblings}, candidates)
	if err != nil {
		return nil, err
	}
	return filterBySelection(matches, selected), nil
}

func (service *Service) filterReferenceRowsWithDecide(idx *index.Index, name string, target decide.Target, siblings []decide.Target, rows []ReferenceRow) ([]ReferenceRow, error) {
	candidates := make([]decide.Candidate, 0, len(rows))
	for indexValue, row := range rows {
		snippet, err := snippetForLine(idx.Root, row.File, row.Line)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, decide.Candidate{
			ID:      candidateID(indexValue),
			File:    row.File,
			Line:    row.Line,
			Name:    name,
			Caller:  row.Caller,
			Context: row.Context,
			Snippet: snippet,
		})
	}

	selected, err := service.selectCandidateIDs(decide.Intent{Command: "references", Name: name, Target: target, Siblings: siblings}, candidates)
	if err != nil {
		return nil, err
	}
	return filterBySelection(rows, selected), nil
}

func candidateID(indexValue int) string {
	return fmt.Sprintf("c%d", indexValue)
}

func snippetForLine(root string, file string, line int) (string, error) {
	source, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
	if err != nil {
		return "", err
	}
	lines := splitLines(source)
	if line <= 0 || line > len(lines) {
		return "", nil
	}
	start := line - aiSnippetRadius - 1
	if start < 0 {
		start = 0
	}
	end := line + aiSnippetRadius
	if end > len(lines) {
		end = len(lines)
	}
	return truncatePromptText(strings.Join(lines[start:end], "\n"), aiCandidateSnippetMaxRune), nil
}

func truncatePromptText(text string, limit int) string {
	runes := []rune(text)
	if limit <= 0 || len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "\n... truncated ..."
}
