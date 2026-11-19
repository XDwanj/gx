package query

import (
	"sort"
	"strings"

	"github.com/XDwanj/gx/internal/index"
	"github.com/XDwanj/gx/internal/output"
)

func (service *Service) Definition(idx *index.Index, intent Intent, scope Scope, page Page, maxLines int) error {
	if err := validateDecideIntent(intent); err != nil {
		return err
	}
	resolvedIdx, filter, err := service.resolveQueryContext(idx, scope)
	if err != nil {
		return err
	}

	matches, err := findDefinitionMatches(resolvedIdx, intent)
	if err != nil {
		return err
	}

	if filter != nil {
		filtered := make([]definitionMatch, 0)
		for _, match := range matches {
			if filter.contains(match.path) {
				filtered = append(filtered, match)
			}
		}
		matches = filtered
	}

	if len(matches) == 0 {
		return service.writeEmptyPage(page)
	}

	sortDefinitionMatches(matches)
	if intent.DefinedIn != "" {
		target, err := service.resolveTarget(resolvedIdx, intent)
		if err != nil {
			return err
		}
		siblings, err := service.collectSiblings(resolvedIdx, intent)
		if err != nil {
			return err
		}
		matches, err = service.filterDefsWithDecide(resolvedIdx, intent.Name, target, siblings, matches)
		if err != nil {
			return err
		}
		if len(matches) == 0 {
			return service.writeEmptyPage(page)
		}
	}
	matches, info := paginateRows(matches, page)

	results := make([]DefinitionResult, 0, len(matches))
	for _, match := range matches {
		body, startLine, ok := readBody(resolvedIdx.Root, match.path, match.symbol)
		if !ok {
			body = ""
			startLine = 0
		}
		lineCount := 0
		if body != "" {
			lineCount = strings.Count(body, "\n") + 1
		}
		result := DefinitionResult{
			File: displayPath(match.path),
			Line: startLine,
			Body: body,
		}
		if lineCount > maxLines && maxLines > 0 {
			lines := strings.Split(body, "\n")
			result.Truncated = true
			result.Lines = lineCount
			result.Body = strings.Join(lines[:maxLines], "\n")
		}
		results = append(results, result)
	}

	service.writePageHint(info)
	if service.json {
		return output.PrintJSON(service.stdout, resultEnvelope{Rows: results, Page: resultPageFor(info)})
	}
	return output.PrintTOON(service.stdout, resultEnvelope{Rows: results, Page: resultPageFor(info)})
}

func findDefinitionMatches(idx *index.Index, intent Intent) ([]definitionMatch, error) {
	matches := make([]definitionMatch, 0)
	for path, data := range idx.Entries {
		for _, symbol := range data.Symbols {
			ok, err := globMatch(intent.Name, symbol.Name)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			if intent.Kind != nil && symbol.Kind != *intent.Kind {
				continue
			}
			matches = append(matches, definitionMatch{path: path, symbol: symbol})
		}
	}
	return matches, nil
}

func sortDefinitionMatches(matches []definitionMatch) {
	sort.Slice(matches, func(left int, right int) bool {
		leftPriority := definitionSymbolPriority(matches[left].symbol.Kind)
		rightPriority := definitionSymbolPriority(matches[right].symbol.Kind)
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		if matches[left].path == matches[right].path {
			if matches[left].symbol.Line == matches[right].symbol.Line {
				return matches[left].symbol.Name < matches[right].symbol.Name
			}
			return matches[left].symbol.Line < matches[right].symbol.Line
		}
		return matches[left].path < matches[right].path
	})
}

func definitionSymbolPriority(kind index.SymbolKind) int {
	switch kind {
	case index.SymbolKindStruct,
		index.SymbolKindEnum,
		index.SymbolKindInterface,
		index.SymbolKindClass,
		index.SymbolKindModule,
		index.SymbolKindType,
		index.SymbolKindConst:
		return definitionPriorityTypes
	case index.SymbolKindFunc:
		return definitionPriorityCallables
	default:
		return definitionPriorityFallback
	}
}
