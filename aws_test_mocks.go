package ssmvars

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/stretchr/testify/mock"
)

type mockSSMAPI struct {
	mock.Mock

	ssmiface.SSMAPI
	parameters []*ssm.Parameter
}

func (m *mockSSMAPI) DeleteParameterWithContext(ctx aws.Context, input *ssm.DeleteParameterInput, opts ...request.Option) (*ssm.DeleteParameterOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*ssm.DeleteParameterOutput), args.Error(1)
}

func (m *mockSSMAPI) DeleteParametersWithContext(ctx aws.Context, input *ssm.DeleteParametersInput, opts ...request.Option) (*ssm.DeleteParametersOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*ssm.DeleteParametersOutput), args.Error(1)
}

func (m *mockSSMAPI) GetParameterWithContext(ctx aws.Context, input *ssm.GetParameterInput, opts ...request.Option) (*ssm.GetParameterOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*ssm.GetParameterOutput), args.Error(1)
}

func (m *mockSSMAPI) GetParametersByPathPagesWithContext(ctx aws.Context, input *ssm.GetParametersByPathInput, cursor func(*ssm.GetParametersByPathOutput, bool) bool, opts ...request.Option) error {
	args := m.Called(ctx, input, cursor, opts)
	cursor(&ssm.GetParametersByPathOutput{Parameters: m.parameters}, true)
	return args.Error(0)
}

func (m *mockSSMAPI) PutParameterWithContext(ctx aws.Context, input *ssm.PutParameterInput, opts ...request.Option) (*ssm.PutParameterOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*ssm.PutParameterOutput), args.Error(1)
}
