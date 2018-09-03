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

// Reader provides are 'read-only' proxy to variables stored in SSM.
type Reader interface {
	// ShowVariable retrieves an individual variable by its name.
	ShowVariable(ctx context.Context, namespace, name string) (*Variable, error)

	// ListVariables lists all variables for a given namespace. It does
	// pagination automatically, with 10 entries per page. In pathological
	// cases, this operation could take a while.
	ListVariables(ctx context.Context, namespace string) ([]*Variable, error)
}

// Writer provides are 'write-only' proxy to variables stored in SSM.
type Writer interface {
	// CreateVariable creates or updates an existing variable.
	CreateVariable(ctx context.Context, namespace string, variable *Variable) (*Variable, error)

	// DeleteVariable deletes an existing variable.
	DeleteVariable(ctx context.Context, namespace, name string) (*Variable, error)
}

// ReadWriter is a 'read-write' proxy to variables stored in SSM.
type ReadWriter interface {
	Reader
	Writer
}
