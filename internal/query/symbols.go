package query

import (
	"sort"

	"github.com/XDwanj/gx/internal/index"
)

func (service *Service) Symbols(idx *index.Index, intent Intent, scope Scope, page Page) error {
	if err := validateDecideIntent(intent); err != nil {
		return err
	}
	rows, resolvedIdx, err := service.symbolRowsInContext(idx, scope, intent)
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		return service.writeEmptyPage(page)
	}

	if intent.DefinedIn != "" {
		target, err := service.resolveTarget(resolvedIdx, intent)
		if err != nil {
			return err
		}
		siblings, err := service.collectSiblings(resolvedIdx, intent)
		if err != nil {
			return err
		}
		rows, err = service.filterSymbolRowsWithDecide(resolvedIdx, intent.Name, target, siblings, rows)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return service.writeEmptyPage(page)
		}
	}

	rows, info := paginateRows(rows, page)
	return service.writeRows(rows, info)
}

func (service *Service) symbolRows(idx *index.Index, scope Scope, intent Intent) ([]SymbolRow, error) {
	rows, _, err := service.symbolRowsInContext(idx, scope, intent)
	return rows, err
}

func (service *Service) symbolRowsInContext(idx *index.Index, scope Scope, intent Intent) ([]SymbolRow, *index.Index, error) {
	resolvedIdx, filter, err := service.resolveQueryContext(idx, scope)
	if err != nil {
		return nil, nil, err
	}

	rows := make([]SymbolRow, 0)
	for path, data := range resolvedIdx.Entries {
		if filter != nil && !filter.contains(path) {
			continue
		}
		for _, symbol := range data.Symbols {
			if intent.Name != "" {
				ok, err := globMatch(intent.Name, symbol.Name)
				if err != nil {
					return nil, nil, err
				}
				if !ok {
					continue
				}
			}
			if intent.Kind != nil && symbol.Kind != *intent.Kind {
				continue
			}
			rows = append(rows, SymbolRow{
				File:      displayPath(path),
				Line:      symbol.Line,
				Name:      symbol.Name,
				Kind:      string(symbol.Kind),
				Signature: symbol.Signature,
			})
		}
	}

	sort.Slice(rows, func(left int, right int) bool {
		if rows[left].File == rows[right].File {
			if rows[left].Line == rows[right].Line {
				return rows[left].Name < rows[right].Name
			}
			return rows[left].Line < rows[right].Line
		}
		return rows[left].File < rows[right].File
	})

	return rows, resolvedIdx, nil
}
