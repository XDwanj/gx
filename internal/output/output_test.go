package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

type probePage struct {
	Total      int `json:"total" toon:"total"`
	NextOffset int `json:"next_offset,omitempty" toon:"next_offset,omitempty"`
}

type probeEnvelope struct {
	Rows []probeRow `json:"rows" toon:"rows"`
	Page probePage  `json:"page" toon:"page"`
}

type probeRow struct {
	File string `json:"file" toon:"file"`
	Line int    `json:"line" toon:"line"`
}

func TestPrintJSONEnvelope(t *testing.T) {
	var buf bytes.Buffer
	err := PrintJSON(&buf, probeEnvelope{
		Rows: []probeRow{{File: "a.go", Line: 1}},
		Page: probePage{Total: 5, NextOffset: 1},
	})
	if err != nil {
		t.Fatal(err)
	}

	var decoded struct {
		Rows []map[string]any `json:"rows"`
		Page map[string]any   `json:"page"`
	}
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("envelope should be valid JSON: %v\n%s", err, buf.String())
	}
	if len(decoded.Rows) != 1 || decoded.Page["total"].(float64) != 5 {
		t.Fatalf("envelope %s", buf.String())
	}
}

func TestPrintTOONEnvelope(t *testing.T) {
	var buf bytes.Buffer
	err := PrintTOON(&buf, probeEnvelope{
		Rows: []probeRow{{File: "a.go", Line: 1}},
		Page: probePage{Total: 5, NextOffset: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "rows") || !strings.Contains(out, "page") || !strings.Contains(out, "total") {
		t.Fatalf("TOON envelope should carry rows and page fields:\n%s", out)
	}
	if strings.Contains(out, "File") || strings.Contains(out, "Total") {
		t.Fatalf("toon tags should rename fields:\n%s", out)
	}
}
