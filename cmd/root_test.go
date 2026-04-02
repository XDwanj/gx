package cmd

import "testing"

func TestRootCommandUsesGx(t *testing.T) {
	if rootCmd.Use != "gx" {
		t.Fatalf("expected root command name gx, got %q", rootCmd.Use)
	}
}

func TestRootCommandExposesVerboseFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("verbose")
	if flag == nil {
		t.Fatalf("expected verbose flag to be registered")
	}
	if flag.DefValue != "false" {
		t.Fatalf("expected verbose default false, got %q", flag.DefValue)
	}
}
