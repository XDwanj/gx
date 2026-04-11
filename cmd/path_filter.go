package cmd

import "github.com/XDwanj/gx/internal/query"

type pathFilterFlags struct {
	include []string
	exclude []string
}

func (flags pathFilterFlags) query(targets []string) query.PathQuery {
	return query.PathQuery{
		Targets: targets,
		Include: flags.include,
		Exclude: flags.exclude,
	}
}

func registerPathFilterFlags(command flagBinder, flags *pathFilterFlags) {
	command.StringArrayVar(&flags.include, "include", nil, "Include file paths matching this glob, even if currently ignored")
	command.StringArrayVar(&flags.exclude, "exclude", nil, "Exclude indexed file paths matching this glob")
}

type flagBinder interface {
	StringArrayVar(p *[]string, name string, value []string, usage string)
}
