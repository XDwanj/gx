package cmd

import (
	"bytes"
	"testing"

	"github.com/XDwanj/gx/internal/app"
)

func TestVersionCommandPrintsHumanReadableVersion(t *testing.T) {
	previousVersion := app.Version
	app.Version = "v0.0.1"
	t.Cleanup(func() {
		app.Version = previousVersion
	})

	var stdout bytes.Buffer
	command := newVersionCmd()
	command.SetOut(&stdout)

	if err := command.RunE(command, nil); err != nil {
		t.Fatalf("run version command: %v", err)
	}

	if stdout.String() != "gx v0.0.1\n" {
		t.Fatalf("unexpected version output %q", stdout.String())
	}
}

func TestVersionCommandPrintsJSONVersion(t *testing.T) {
	previousVersion := app.Version
	previousFlags := rootFlags
	app.Version = "v0.0.1"
	rootFlags.JSON = true
	t.Cleanup(func() {
		app.Version = previousVersion
		rootFlags = previousFlags
	})

	var stdout bytes.Buffer
	command := newVersionCmd()
	command.SetOut(&stdout)

	if err := command.RunE(command, nil); err != nil {
		t.Fatalf("run version command: %v", err)
	}

	expected := "{\n  \"version\": \"v0.0.1\"\n}\n"
	if stdout.String() != expected {
		t.Fatalf("unexpected JSON version output %q", stdout.String())
	}
}
