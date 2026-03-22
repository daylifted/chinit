package main

import (
	"strings"
)

// buildArgv splits the command string and appends any extra args.
func buildArgv(args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, errBadArgs
	}
	parts := strings.Fields(args[0])
	if len(parts) == 0 {
		return nil, errBadArgs
	}
	return append(parts, args[1:]...), nil
}
