package query

import (
	"fmt"
	"gx/internal/output"
	"os"
	"path/filepath"
	"strings"
)

const (
	markdownHeadingLevelMin  = 1
	markdownHeadingLevelMax  = 6
	markdownFenceIndentLimit = 3
	markdownFenceMinRunes    = 3
	markdownIndentLimit      = 3
	markdownExtensionMD      = ".md"
	markdownExtensionLong    = ".markdown"
)

type markdownOverviewRow struct {
	Level   int    `json:"level"`
	Heading string `json:"heading"`
}

type markdownHeading struct {
	Level int
	Text  string
}

type markdownFence struct {
	Marker byte
	Count  int
	Rest   string
}

func IsMarkdownPath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case markdownExtensionMD, markdownExtensionLong:
		return true
	default:
		return false
	}
}

func (service *Service) MarkdownOverview(path string) error {
	relPath := normalizeRelativePath(path, service.root)
	absPath := filepath.Join(service.root, relPath)

	source, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("gx: read markdown file: %w", err)
	}

	headings := extractMarkdownHeadings(source)
	if len(headings) == 0 {
		_, _ = fmt.Fprintf(service.stderr, "gx: no headings found in %s\n", displayPath(relPath))
		return nil
	}

	rows := make([]markdownOverviewRow, 0, len(headings))
	for _, heading := range headings {
		rows = append(rows, markdownOverviewRow{
			Level:   heading.Level,
			Heading: heading.Text,
		})
	}

	if service.json {
		return output.PrintJSON(service.stdout, rows)
	}
	return output.PrintTOON(service.stdout, rows)
}

func extractMarkdownHeadings(source []byte) []markdownHeading {
	lines := splitLines(source)
	headings := make([]markdownHeading, 0)
	inFence := false
	activeFence := markdownFence{}

	for index := 0; index < len(lines); index++ {
		line := lines[index]

		if inFence {
			fence, ok := parseMarkdownFence(line)
			if ok && fence.Marker == activeFence.Marker && fence.Count >= activeFence.Count && strings.TrimSpace(fence.Rest) == "" {
				inFence = false
				activeFence = markdownFence{}
			}
			continue
		}

		if fence, ok := parseMarkdownFence(line); ok {
			inFence = true
			activeFence = fence
			continue
		}

		if heading, ok := parseATXHeading(line); ok {
			headings = append(headings, heading)
			continue
		}

		if index+1 >= len(lines) {
			continue
		}

		if heading, ok := parseSetextHeading(line, lines[index+1]); ok {
			headings = append(headings, heading)
			index++
		}
	}

	return headings
}

func parseATXHeading(line string) (markdownHeading, bool) {
	trimmed, indent := trimMarkdownIndent(line)
	if indent > markdownIndentLimit || trimmed == "" || trimmed[0] != '#' {
		return markdownHeading{}, false
	}

	level := countLeadingByte(trimmed, '#')
	if level < markdownHeadingLevelMin || level > markdownHeadingLevelMax {
		return markdownHeading{}, false
	}

	if len(trimmed) > level && trimmed[level] != ' ' && trimmed[level] != '\t' {
		return markdownHeading{}, false
	}

	text := trimATXHeadingText(trimmed[level:])
	if text == "" {
		return markdownHeading{}, false
	}

	return markdownHeading{Level: level, Text: text}, true
}

func parseSetextHeading(textLine string, underlineLine string) (markdownHeading, bool) {
	text := strings.TrimSpace(textLine)
	if text == "" {
		return markdownHeading{}, false
	}

	trimmed, indent := trimMarkdownIndent(underlineLine)
	if indent > markdownIndentLimit || trimmed == "" {
		return markdownHeading{}, false
	}
	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" {
		return markdownHeading{}, false
	}

	marker := trimmed[0]
	if marker != '=' && marker != '-' {
		return markdownHeading{}, false
	}
	if countLeadingByte(trimmed, marker) != len(trimmed) {
		return markdownHeading{}, false
	}

	level := 1
	if marker == '-' {
		level = 2
	}

	return markdownHeading{Level: level, Text: text}, true
}

func parseMarkdownFence(line string) (markdownFence, bool) {
	trimmed, indent := trimMarkdownIndent(line)
	if indent > markdownFenceIndentLimit || trimmed == "" {
		return markdownFence{}, false
	}

	marker := trimmed[0]
	if marker != '`' && marker != '~' {
		return markdownFence{}, false
	}

	count := countLeadingByte(trimmed, marker)
	if count < markdownFenceMinRunes {
		return markdownFence{}, false
	}

	return markdownFence{
		Marker: marker,
		Count:  count,
		Rest:   trimmed[count:],
	}, true
}

func trimMarkdownIndent(line string) (string, int) {
	indent := 0
	for indent < len(line) && line[indent] == ' ' {
		indent++
	}
	return line[indent:], indent
}

func countLeadingByte(text string, target byte) int {
	count := 0
	for count < len(text) && text[count] == target {
		count++
	}
	return count
}

func trimATXHeadingText(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ""
	}

	end := len(trimmed)
	for end > 0 && trimmed[end-1] == '#' {
		end--
	}
	if end < len(trimmed) && end > 0 && trimmed[end-1] == ' ' {
		return strings.TrimSpace(trimmed[:end])
	}
	return trimmed
}
