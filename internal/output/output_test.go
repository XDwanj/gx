package output

import (
	"bytes"
	"testing"
)

type toonRow struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Signature string `json:"signature"`
}

type quotedRow struct {
	Name string `json:"name"`
	Note string `json:"note"`
}

func TestPrintTOONTabularRows(t *testing.T) {
	rows := []toonRow{
		{Name: "newDefinitionCmd", Kind: "fn", Signature: "func newDefinitionCmd() *cobra.Command"},
	}

	var output bytes.Buffer
	if err := PrintTOON(&output, rows); err != nil {
		t.Fatalf("print toon: %v", err)
	}

	expected := "[1]{name,kind,signature}:\n  newDefinitionCmd,fn,func newDefinitionCmd() *cobra.Command\n"
	if output.String() != expected {
		t.Fatalf("unexpected TOON output:\nexpected:\n%s\ngot:\n%s", expected, output.String())
	}
}

func TestPrintTOONQuotesDelimitedStrings(t *testing.T) {
	rows := []quotedRow{
		{Name: "alpha", Note: "contains,comma"},
	}

	var output bytes.Buffer
	if err := PrintTOON(&output, rows); err != nil {
		t.Fatalf("print toon: %v", err)
	}

	expected := "[1]{name,note}:\n  alpha,\"contains,comma\"\n"
	if output.String() != expected {
		t.Fatalf("unexpected TOON output:\nexpected:\n%s\ngot:\n%s", expected, output.String())
	}
}

func TestPrintTOONEmptySlice(t *testing.T) {
	var output bytes.Buffer
	if err := PrintTOON(&output, []toonRow{}); err != nil {
		t.Fatalf("print toon: %v", err)
	}

	expected := "[0]{}:\n"
	if output.String() != expected {
		t.Fatalf("unexpected TOON output:\nexpected:\n%s\ngot:\n%s", expected, output.String())
	}
}
