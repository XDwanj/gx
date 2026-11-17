package query

import (
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"github.com/XDwanj/gx/internal/decide"
)

type Service struct {
	root    string
	stdout  io.Writer
	stderr  io.Writer
	json    bool
	verbose bool
	decider decide.Decider
	debugMu sync.Mutex
}

func NewService(root string, stdout io.Writer, stderr io.Writer, json bool, verbose bool) *Service {
	return &Service{
		root:    filepath.Clean(root),
		stdout:  stdout,
		stderr:  stderr,
		json:    json,
		verbose: verbose,
	}
}

func (service *Service) Root() string {
	return service.root
}

func (service *Service) debugf(format string, args ...any) {
	if !service.verbose {
		return
	}
	service.debugMu.Lock()
	defer service.debugMu.Unlock()
	_, _ = fmt.Fprintf(service.stderr, "gx: debug: "+format+"\n", args...)
}
