package ssmvars

import (
	"context"
	"fmt"
)

// Environment is a convenience function presenting the content of a Reader in
// a way that's compatible with operations that expect os.Environment.
func Environment(ctx context.Context, reader Reader, namespace string) ([]string, error) {
	variables, err := reader.ListVariables(ctx, namespace)
	if err != nil {
		return nil, err
	}

	lines := make([]string, len(variables), len(variables))
	for index, variable := range variables {
		lines[index] = fmt.Sprintf("%s=%s", variable.Name, variable.Value)
	}

	return lines, nil
}
