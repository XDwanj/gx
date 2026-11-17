package query

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/XDwanj/gx/internal/index"
	"github.com/XDwanj/gx/internal/output"
)

type overviewEnvelope struct {
	Sections []OverviewSection `json:"sections" toon:"sections"`
}

func (service *Service) Overview(idx *index.Index, paths []string, full bool, page Page) error {
	if len(paths) <= 1 {
		if len(paths) == 0 {
			return service.Symbols(idx, Intent{}, Scope{Targets: paths}, page)
		}
		return service.singleOverview(idx, paths[0], full, page)
	}

	sections := make([]OverviewSection, 0, len(paths))
	for _, target := range paths {
		section, empty, err := service.overviewSectionForTarget(idx, target, full, page)
		if err != nil {
			return err
		}
		if empty {
			continue
		}
		sections = append(sections, section)
	}

	if len(sections) == 0 {
		_, _ = fmt.Fprintln(service.stderr, "gx: no matches")
		return nil
	}

	if service.json {
		return output.PrintJSON(service.stdout, overviewEnvelope{Sections: sections})
	}
	return output.PrintTOON(service.stdout, overviewEnvelope{Sections: sections})
}

func (service *Service) singleOverview(idx *index.Index, target string, full bool, page Page) error {
	switch classifyOverviewTarget(target) {
	case overviewTargetDirectory:
		return service.DirectoryOverview(idx, target, full, page)
	case overviewTargetMarkdown:
		return service.MarkdownOverview(target)
	default:
		return service.Symbols(idx, Intent{}, Scope{Targets: []string{target}}, Page{})
	}
}

func (service *Service) overviewSectionForTarget(idx *index.Index, target string, full bool, page Page) (OverviewSection, bool, error) {
	targetLabel := displayPath(normalizeRelativePath(target, service.root))

	switch classifyOverviewTarget(target) {
	case overviewTargetDirectory:
		if idx == nil {
			return OverviewSection{}, false, fmt.Errorf("gx: overview index is required for directories")
		}
		var info pageInfo
		var rows any
		var err error
		if full {
			var fullRows []DirOverviewFullRow
			fullRows, err = service.directoryOverviewFullRows(idx, target)
			if err == nil {
				fullRows, info = paginateRows(fullRows, page)
				rows = fullRows
			}
		} else {
			var dirRows []DirOverviewRow
			dirRows, err = service.directoryOverviewRows(idx, target)
			if err == nil {
				dirRows, info = paginateRows(dirRows, page)
				rows = dirRows
			}
		}
		if err != nil {
			return OverviewSection{}, false, err
		}
		service.writeScopedPageHint(targetLabel, info)
		return OverviewSection{
			Target:     targetLabel,
			TargetKind: string(overviewTargetDirectory),
			Rows:       rows,
			Page:       resultPageFor(info),
		}, info.Returned == 0 && !info.OutOfRange, nil
	case overviewTargetMarkdown:
		rows, err := service.markdownOverviewRows(target)
		if err != nil {
			return OverviewSection{}, false, err
		}
		if len(rows) == 0 {
			_, _ = fmt.Fprintf(service.stderr, "gx: no headings found in %s\n", targetLabel)
			return OverviewSection{}, true, nil
		}
		return OverviewSection{
			Target:     targetLabel,
			TargetKind: string(overviewTargetMarkdown),
			Rows:       rows,
			Page:       resultPage{Total: len(rows)},
		}, false, nil
	default:
		if idx == nil {
			return OverviewSection{}, false, fmt.Errorf("gx: overview index is required for files")
		}
		rows, err := service.symbolRows(idx, Scope{Targets: []string{target}}, Intent{})
		if err != nil {
			return OverviewSection{}, false, err
		}
		if len(rows) == 0 {
			_, _ = fmt.Fprintf(service.stderr, "gx: no matches in %s\n", targetLabel)
			return OverviewSection{}, true, nil
		}
		return OverviewSection{
			Target:     targetLabel,
			TargetKind: string(overviewTargetFile),
			Rows:       rows,
			Page:       resultPage{Total: len(rows)},
		}, false, nil
	}
}

func classifyOverviewTarget(path string) overviewTargetKind {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return overviewTargetDirectory
	}
	if err == nil && IsMarkdownPath(path) {
		return overviewTargetMarkdown
	}
	return overviewTargetFile
}

func (service *Service) DirectoryOverview(idx *index.Index, dir string, full bool, page Page) error {
	if full {
		rows, err := service.directoryOverviewFullRows(idx, dir)
		if err != nil {
			return err
		}
		rows, info := paginateRows(rows, page)
		return service.writeRows(rows, info)
	}

	rows, err := service.directoryOverviewRows(idx, dir)
	if err != nil {
		return err
	}
	rows, info := paginateRows(rows, page)
	return service.writeRows(rows, info)
}

func (service *Service) directoryOverviewRows(idx *index.Index, dir string) ([]DirOverviewRow, error) {
	data, err := buildDirectoryOverviewData(idx, dir)
	if err != nil {
		return nil, err
	}

	rows := make([]DirOverviewRow, 0)
	for _, name := range sortedKeys(data.subdirs) {
		stats := data.subdirs[name]
		rows = append(rows, DirOverviewRow{
			File:    formatSubdir(data.relDir, name),
			Symbols: fmt.Sprintf("(%d files, %d symbols)", stats[0], stats[1]),
		})
	}

	for _, fileEntry := range data.directFiles {
		symbols := prepareSymbols(fileEntry.data.Symbols)
		if len(symbols) == 0 {
			continue
		}
		seen := map[string]bool{}
		names := make([]string, 0)
		for _, symbol := range symbols[:minInt(len(symbols), directoryOverviewMaxSymbols)] {
			if seen[symbol.Name] {
				continue
			}
			seen[symbol.Name] = true
			names = append(names, symbol.Name)
		}
		suffix := ""
		if len(symbols) > len(names) {
			suffix = fmt.Sprintf(", ... (+%d more)", len(symbols)-len(names))
		}
		rows = append(rows, DirOverviewRow{
			File:    displayPath(fileEntry.path),
			Symbols: strings.Join(names, ", ") + suffix,
		})
	}

	return rows, nil
}

func (service *Service) directoryOverviewFullRows(idx *index.Index, dir string) ([]DirOverviewFullRow, error) {
	data, err := buildDirectoryOverviewData(idx, dir)
	if err != nil {
		return nil, err
	}

	rows := make([]DirOverviewFullRow, 0)
	for _, name := range sortedKeys(data.subdirs) {
		stats := data.subdirs[name]
		rows = append(rows, DirOverviewFullRow{
			File:      formatSubdir(data.relDir, name),
			Name:      fmt.Sprintf("(%d files, %d symbols)", stats[0], stats[1]),
			Kind:      "",
			Signature: "",
		})
	}

	for _, fileEntry := range data.directFiles {
		symbols := prepareSymbols(fileEntry.data.Symbols)
		if len(symbols) == 0 {
			continue
		}
		total := len(symbols)
		for _, symbol := range symbols[:minInt(total, directoryOverviewMaxSymbols)] {
			rows = append(rows, DirOverviewFullRow{
				File:      displayPath(fileEntry.path),
				Name:      symbol.Name,
				Kind:      string(symbol.Kind),
				Signature: symbol.Signature,
			})
		}
		if total > directoryOverviewMaxSymbols {
			rows = append(rows, DirOverviewFullRow{
				File: displayPath(fileEntry.path),
				Name: fmt.Sprintf("... (+%d more)", total-directoryOverviewMaxSymbols),
			})
		}
	}

	return rows, nil
}

func buildDirectoryOverviewData(idx *index.Index, dir string) (directoryOverviewData, error) {
	relDir := normalizeRelativePath(dir, idx.Root)
	if relDir == "." {
		relDir = ""
	}

	allEntries := make([]directoryOverviewEntry, 0)
	for path, data := range idx.Entries {
		if relDir != "" && !strings.HasPrefix(path, relDir) {
			continue
		}
		if isTestFile(path) {
			continue
		}
		allEntries = append(allEntries, directoryOverviewEntry{path: path, data: data})
	}

	if len(allEntries) == 0 {
		return directoryOverviewData{}, fmt.Errorf("gx: no indexed files under %s", displayPath(relDir))
	}

	directFiles := make([]directoryOverviewEntry, 0)
	subdirs := map[string][2]int{}
	for _, item := range allEntries {
		child, ok := childComponent(item.path, relDir)
		if !ok {
			continue
		}
		nonTestCount := 0
		for _, symbol := range item.data.Symbols {
			if !symbol.IsTest {
				nonTestCount++
			}
		}

		if !strings.Contains(child, string(filepath.Separator)) && filepath.Ext(child) != "" {
			directFiles = append(directFiles, item)
			continue
		}
		stats := subdirs[child]
		stats[0]++
		stats[1] += nonTestCount
		subdirs[child] = stats
	}

	sort.Slice(directFiles, func(left int, right int) bool {
		return directFiles[left].path < directFiles[right].path
	})

	return directoryOverviewData{
		relDir:      relDir,
		directFiles: directFiles,
		subdirs:     subdirs,
	}, nil
}
