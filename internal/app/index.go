package app

import (
	"io"

	"github.com/XDwanj/gx/internal/index"
	"github.com/XDwanj/gx/internal/output"
)

type IndexResult struct {
	Root         string `json:"root" toon:"root"`
	Cache        string `json:"cache" toon:"cache"`
	IndexedFiles int    `json:"indexed_files" toon:"indexed_files"`
}

func Preindex(root string, json bool, verbose bool, stdout io.Writer, stderr io.Writer) error {
	idx, err := index.LoadOrBuildWithOptions(root, index.LoadOptions{
		Verbose: verbose,
		Stderr:  stderr,
	})
	if err != nil {
		return err
	}

	result := IndexResult{
		Root:         idx.Root,
		Cache:        index.CachePathFor(idx.Root),
		IndexedFiles: len(idx.Entries),
	}
	if json {
		return output.PrintJSON(stdout, result)
	}
	return output.PrintTOON(stdout, result)
}
