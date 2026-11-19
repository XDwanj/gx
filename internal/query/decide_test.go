package query

import (
	"io"
	"testing"

	"github.com/XDwanj/gx/internal/index"
)

func TestCollectSiblings(t *testing.T) {
	funcKind := index.SymbolKindFunc
	service := NewService(t.TempDir(), io.Discard, io.Discard, false, false)
	idx := &index.Index{
		Entries: map[string]index.FileData{
			"b/b.go": {Symbols: []index.Symbol{
				{Name: "List", Kind: funcKind, Line: 20, Signature: "func (PublishScheduleDevice) List()"},
				{Name: "Other", Kind: funcKind, Line: 30},
			}},
			"a/a.go": {Symbols: []index.Symbol{
				{Name: "ListAll", Kind: funcKind, Line: 40},
				{Name: "List", Kind: index.SymbolKindStruct, Line: 5},
				{Name: "List", Kind: funcKind, Line: 10, Signature: "func (PublishSchedule) List()"},
			}},
		},
	}

	siblings, err := service.collectSiblings(idx, Intent{Name: "List", Kind: &funcKind})
	if err != nil {
		t.Fatalf("collectSiblings: %v", err)
	}
	if len(siblings) != 2 {
		t.Fatalf("siblings = %+v", siblings)
	}
	if siblings[0].File != "a/a.go" || siblings[0].Line != 10 || siblings[1].File != "b/b.go" || siblings[1].Line != 20 {
		t.Fatalf("siblings not sorted by file and line: %+v", siblings)
	}
	if siblings[1].Signature != "func (PublishScheduleDevice) List()" {
		t.Fatalf("sibling signature missing: %+v", siblings[1])
	}

	globbed, err := service.collectSiblings(idx, Intent{Name: "List*"})
	if err != nil {
		t.Fatalf("collectSiblings glob: %v", err)
	}
	if len(globbed) != 4 {
		t.Fatalf("glob siblings = %+v", globbed)
	}
}
