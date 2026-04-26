package cmd

import "github.com/spf13/pflag"

func registerDefineInFlag(flags *pflag.FlagSet, value *string) {
	flags.StringVar(value, "define-in", "", "Use AI to disambiguate matches by the definition file")
}

func resolveDefineInPath(root string, path string) (string, error) {
	if path == "" {
		return "", nil
	}
	paths, err := resolveTargetPaths(root, []string{path})
	if err != nil {
		return "", err
	}
	return paths[0], nil
}
