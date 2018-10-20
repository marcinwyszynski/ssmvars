package ssmvars

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/pkg/errors"
)

type repositoryImpl struct {
	ssmiface.SSMAPI
	pathPrefix string
	kmsKeyID   *string
}

// New returns a Variables ReadWriter backed by SSM Parameter Store.
func New(client ssmiface.SSMAPI, prefix, kmsKeyID string) ReadWriter {
	return newRepository(client, prefix, aws.String(kmsKeyID))
}

// NewReader returns a Variables Reader backed by SSM Parameter Store.
func NewReader(client ssmiface.SSMAPI, prefix string) Reader {
	return newRepository(client, prefix, nil)
}

// NewWriter returns a Variables Writer backed by SSM Parameter Store.
func NewWriter(client ssmiface.SSMAPI, prefix, kmsKeyID string) Writer {
	return newRepository(client, prefix, aws.String(kmsKeyID))
}

func newRepository(client ssmiface.SSMAPI, prefix string, kmsKeyID *string) *repositoryImpl {
	return &repositoryImpl{
		SSMAPI:     client,
		pathPrefix: path.Join(prefix, "variables"),
		kmsKeyID:   kmsKeyID,
	}
}

func (r *repositoryImpl) ListVariables(ctx context.Context, namespace string) ([]*Variable, error) {
	input := &ssm.GetParametersByPathInput{
		Path:           aws.String(r.namespacePrefix(namespace)),
		Recursive:      aws.Bool(false),
		WithDecryption: aws.Bool(true),
		MaxResults:     aws.Int64(10),
	}

	vars := make([]*Variable, 0)
	cursor := func(cur *ssm.GetParametersByPathOutput, _ bool) bool {
		for _, parameter := range cur.Parameters {
			vars = append(vars, r.toVariable(namespace, parameter))
		}
		return true
	}

	return vars, r.GetParametersByPathPagesWithContext(ctx, input, cursor)
}

func (r *repositoryImpl) CreateVariable(ctx context.Context, namespace string, variable *Variable) (*Variable, error) {
	input := &ssm.PutParameterInput{
		Name:      aws.String(path.Join(r.pathPrefix, namespace, variable.Name)),
		Type:      aws.String("String"),
		Value:     aws.String(variable.Value),
		Overwrite: aws.Bool(true),
	}
	if variable.WriteOnly {
		input.Type = aws.String("SecureString")
		input.KeyId = r.kmsKeyID
	}
	if _, err := r.PutParameterWithContext(ctx, input); err != nil {
		return nil, errors.Wrap(err, "couldn't put the variable into the parameter store")
	}
	return variable, nil
}

func (r *repositoryImpl) DeleteVariable(ctx context.Context, namespace, name string) (*Variable, error) {
	ret, err := r.ShowVariable(ctx, namespace, name)
	if err != nil {
		return nil, err
	}
	input := &ssm.DeleteParameterInput{Name: r.variablePath(namespace, name)}
	if _, err = r.DeleteParameterWithContext(ctx, input); err != nil {
		return nil, errors.Wrap(err, "couldn't delete variable from the parameter store")
	}
	return ret, nil
}

func (r *repositoryImpl) ShowVariable(ctx context.Context, namespace, name string) (*Variable, error) {
	input := &ssm.GetParameterInput{Name: r.variablePath(namespace, name), WithDecryption: aws.Bool(true)}
	output, err := r.GetParameterWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return r.toVariable(namespace, output.Parameter), nil
}

func (r *repositoryImpl) Reset(ctx context.Context, namespace string) error {
	input := &ssm.GetParametersByPathInput{
		Path:           aws.String(r.namespacePrefix(namespace)),
		Recursive:      aws.Bool(false),
		WithDecryption: aws.Bool(true),
		MaxResults:     aws.Int64(10),
	}

	var deleteError error

	cursor := func(cur *ssm.GetParametersByPathOutput, _ bool) bool {
		num := len(cur.Parameters)
		if num == 0 {
			return true
		}

		deleteInput := &ssm.DeleteParametersInput{Names: make([]*string, num, num)}
		for i, parameter := range cur.Parameters {
			deleteInput.Names[i] = parameter.Name
		}

		_, deleteError = r.DeleteParametersWithContext(ctx, deleteInput)
		return deleteError == nil
	}

	if err := r.GetParametersByPathPagesWithContext(ctx, input, cursor); err != nil {
		return err
	}
	if deleteError != nil {
		return deleteError
	}
	return nil
}

func (r *repositoryImpl) namespacePrefix(namespace string) string {
	return fmt.Sprintf("%s/%s/", r.pathPrefix, namespace)
}

func (r *repositoryImpl) toVariable(namespace string, parameter *ssm.Parameter) *Variable {
	return &Variable{
		Name:      r.variableName(namespace, parameter),
		Value:     *parameter.Value,
		WriteOnly: *parameter.Type == "SecureString",
	}
}

func (r *repositoryImpl) variableName(namespace string, parameter *ssm.Parameter) string {
	return strings.TrimPrefix(*parameter.Name, r.namespacePrefix(namespace))
}

func (r *repositoryImpl) variablePath(namespace, name string) *string {
	return aws.String(path.Join(r.pathPrefix, namespace, name))
}
