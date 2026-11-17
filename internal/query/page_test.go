package query

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/XDwanj/gx/internal/index"
)

func TestPaginateRows(t *testing.T) {
	rows := []int{1, 2, 3, 4, 5}

	paged, info := paginateRows(rows, Page{Limit: 2})
	if len(paged) != 2 || info.Total != 5 || info.NextOffset != 2 || !info.Truncated {
		t.Fatalf("limit page = %v info %+v", paged, info)
	}

	paged, info = paginateRows(rows, Page{Limit: 2, Offset: 4})
	if len(paged) != 1 || info.NextOffset != 5 || info.Truncated {
		t.Fatalf("last page = %v info %+v", paged, info)
	}

	paged, info = paginateRows(rows, Page{Offset: 99})
	if len(paged) != 0 || !info.OutOfRange || info.Total != 5 {
		t.Fatalf("out-of-range page = %v info %+v", paged, info)
	}

	paged, info = paginateRows(rows, Page{})
	if len(paged) != 5 || info.Truncated || info.OutOfRange {
		t.Fatalf("all page = %v info %+v", paged, info)
	}
}

type envelopeProbe struct {
	Rows []SymbolRow `json:"rows"`
	Page struct {
		Total      int  `json:"total"`
		Offset     int  `json:"offset"`
		NextOffset int  `json:"next_offset"`
		OutOfRange bool `json:"out_of_range"`
	} `json:"page"`
}

func testService(t *testing.T) (*Service, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	return NewService(t.TempDir(), &stdout, &stderr, true, false), &stdout, &stderr
}

func testIndex(root string) *index.Index {
	return &index.Index{
		Root: root,
		Entries: map[string]index.FileData{
			"a.go": {
				Meta: index.FileEntry{Language: "go"},
				Symbols: []index.Symbol{
					{Name: "Alpha", Kind: index.SymbolKindFunc, Line: 1},
					{Name: "Beta", Kind: index.SymbolKindFunc, Line: 10},
				},
			},
		},
	}
}

func TestSymbolsEnvelopePagination(t *testing.T) {
	service, stdout, _ := testService(t)
	idx := testIndex(service.Root())

	if err := service.Symbols(idx, Intent{Name: "*"}, Scope{}, Page{Limit: 1}); err != nil {
		t.Fatalf("Symbols: %v", err)
	}

	var got envelopeProbe
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("output is not the documented envelope: %v\n%s", err, stdout.String())
	}
	if len(got.Rows) != 1 || got.Page.Total != 2 || got.Page.NextOffset != 1 {
		t.Fatalf("rows=%d page=%+v", len(got.Rows), got.Page)
	}
}

func TestSymbolsOutOfRangePage(t *testing.T) {
	service, stdout, _ := testService(t)
	idx := testIndex(service.Root())

	if err := service.Symbols(idx, Intent{Name: "*"}, Scope{}, Page{Offset: 50}); err != nil {
		t.Fatalf("Symbols: %v", err)
	}

	var got envelopeProbe
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("output is not the documented envelope: %v\n%s", err, stdout.String())
	}
	if len(got.Rows) != 0 || !got.Page.OutOfRange || got.Page.Total != 2 {
		t.Fatalf("rows=%d page=%+v", len(got.Rows), got.Page)
	}
}

func TestSymbolsNoMatchesStillWritesEnvelope(t *testing.T) {
	service, stdout, stderr := testService(t)
	idx := testIndex(service.Root())

	if err := service.Symbols(idx, Intent{Name: "zzz"}, Scope{}, Page{}); err != nil {
		t.Fatalf("Symbols: %v", err)
	}

	var got envelopeProbe
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("output is not the documented envelope: %v\n%s", err, stdout.String())
	}
	if got.Page.Total != 0 {
		t.Fatalf("page=%+v", got.Page)
	}
	if !strings.Contains(stderr.String(), "no matches") {
		t.Fatalf("stderr should keep the human hint, got %q", stderr.String())
	}
}
