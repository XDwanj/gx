package app

import (
	"fmt"
	"gx/internal/output"
	"io"
)

var Version = "dev"

type VersionInfo struct {
	Version string `json:"version"`
}

func CurrentVersion() VersionInfo {
	return VersionInfo{Version: Version}
}

func PrintVersion(writer io.Writer, json bool) error {
	info := CurrentVersion()
	if json {
		return output.PrintJSON(writer, info)
	}

	_, err := fmt.Fprintf(writer, "gx %s\n", info.Version)
	return err
}
