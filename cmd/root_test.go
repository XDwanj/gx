package cmd

import "testing"

func TestRootCommandUsesGx(t *testing.T) {
	if rootCmd.Use != "gx" {
		t.Fatalf("expected root command name gx, got %q", rootCmd.Use)
	}
}
