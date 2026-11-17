package query

import (
	"fmt"

	"github.com/XDwanj/gx/internal/output"
)

func paginateRows[T any](rows []T, page Page) ([]T, pageInfo) {
	info := pageInfo{
		Total:  len(rows),
		Offset: page.Offset,
	}

	if page.Offset >= len(rows) {
		return rows[:0], pageInfo{
			Total:      len(rows),
			Offset:     page.Offset,
			OutOfRange: page.Offset > 0,
		}
	}

	paged := rows
	if page.Offset > 0 {
		paged = paged[page.Offset:]
	}
	if page.Limit > 0 && len(paged) > page.Limit {
		paged = paged[:page.Limit]
		info.Truncated = true
	}
	info.Returned = len(paged)
	info.NextOffset = page.Offset + len(paged)

	return paged, info
}

func resultPageFor(info pageInfo) resultPage {
	page := resultPage{
		Total:      info.Total,
		Offset:     info.Offset,
		OutOfRange: info.OutOfRange,
	}
	if info.Truncated {
		page.NextOffset = info.NextOffset
	}
	return page
}

func (service *Service) writeRows(rows any, info pageInfo) error {
	service.writePageHint(info)
	if service.json {
		return output.PrintJSON(service.stdout, resultEnvelope{Rows: rows, Page: resultPageFor(info)})
	}
	return output.PrintTOON(service.stdout, resultEnvelope{Rows: humanizeRows(rows), Page: resultPageFor(info)})
}

func (service *Service) writeEmptyPage(page Page) error {
	_, _ = fmt.Fprintln(service.stderr, "gx: no matches")
	info := pageInfo{Offset: page.Offset, OutOfRange: page.Offset > 0}
	if service.json {
		return output.PrintJSON(service.stdout, resultEnvelope{Rows: []any{}, Page: resultPageFor(info)})
	}
	return output.PrintTOON(service.stdout, resultEnvelope{Rows: []any{}, Page: resultPageFor(info)})
}

func (service *Service) writePageHint(info pageInfo) {
	service.writeScopedPageHint("", info)
}

func (service *Service) writeScopedPageHint(scope string, info pageInfo) {
	prefix := ""
	if scope != "" {
		prefix = fmt.Sprintf("%s: ", scope)
	}
	if info.OutOfRange {
		_, _ = fmt.Fprintf(service.stderr, "gx: %soffset %d exceeds %d results\n", prefix, info.Offset, info.Total)
		return
	}
	if info.Truncated {
		start := info.Offset + 1
		end := info.Offset + info.Returned
		_, _ = fmt.Fprintf(service.stderr, "gx: %sshowing %d-%d of %d results; more: --offset %d\n", prefix, start, end, info.Total, info.NextOffset)
	}
}
