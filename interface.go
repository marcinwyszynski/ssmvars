package ssmvars

import "context"

// Variable represents a single configuration variable.
type Variable struct {
	// Name of the variable.
	Name string

	// Value of the variable.
	Value string

	// Whether the variable should be readable from the UI.
	WriteOnly bool
}

// VariablesRepository is a proxy to variables.
type VariablesRepository interface {
	// This can require multiple network calls to the AWS parameter store
	// so keep in mind to set the timeout of the context to be big
	// as this operation could take a while in pathological cases
	ListVariables(ctx context.Context, namespace string) ([]*Variable, error)

	CreateVariable(ctx context.Context, namespace string, variable *Variable) (*Variable, error)
	DeleteVariable(ctx context.Context, namespace, name string) (*Variable, error)
}
