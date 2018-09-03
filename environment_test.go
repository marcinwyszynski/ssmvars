package ssmvars

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
)

type mockReader struct {
	mock.Mock
	Reader
}

func (m *mockReader) ListVariables(ctx context.Context, namespace string) ([]*Variable, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).([]*Variable), args.Error(1)
}

func TestEnvironmentOK(t *testing.T) {
	const namespace = "namespace"
	reader := new(mockReader)
	ctx := context.Background()

	reader.On("ListVariables", ctx, namespace).Return([]*Variable{
		{Name: "FIRST", Value: "first"},
		{Name: "LAST", Value: "last"},
	}, nil)

	environment, err := Environment(ctx, reader, namespace)
	assert.NoError(t, err)

	assert.Len(t, environment, 2)
	assert.Contains(t, environment, "FIRST=first")
	assert.Contains(t, environment, "LAST=last")
}

func TestEnvironmentError(t *testing.T) {
	const namespace = "namespace"
	reader := new(mockReader)
	ctx := context.Background()

	reader.
		On("ListVariables", ctx, namespace).
		Return(([]*Variable)(nil), errors.New("bacon"))

	environment, err := Environment(ctx, reader, namespace)
	assert.EqualError(t, err, "bacon")
	assert.Nil(t, environment)
}
