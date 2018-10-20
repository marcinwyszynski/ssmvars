package ssmvars

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	testPrefix   = "/testPrefix"
	testKMSKeyID = "testKMSKeyID"
)

type variableRepositoryTestSuite struct {
	suite.Suite

	ctx     context.Context
	mockAPI *mockSSMAPI
	sut     ReadWriter
}

func (vs *variableRepositoryTestSuite) SetupTest() {
	vs.ctx = context.Background()
	vs.mockAPI = new(mockSSMAPI)
	vs.sut = New(vs.mockAPI, testPrefix, testKMSKeyID)
}

func (vs *variableRepositoryTestSuite) TestListVariables() {
	const scope = "scope"

	vs.mockAPI.parameters = []*ssm.Parameter{
		{
			Name:  aws.String("/testPrefix/variables/scope/PLAIN"),
			Type:  aws.String("String"),
			Value: aws.String("plain"),
		},
		{
			Name:  aws.String("/testPrefix/variables/scope/SECRET"),
			Type:  aws.String("SecureString"),
			Value: aws.String("secret"),
		},
	}

	vs.mockAPI.On(
		"GetParametersByPathPagesWithContext",
		vs.ctx,
		mock.MatchedBy(func(in interface{}) bool {
			input, ok := in.(*ssm.GetParametersByPathInput)
			if !ok {
				return false
			}
			vs.Equal("/testPrefix/variables/scope/", *input.Path)
			vs.False(*input.Recursive)
			vs.True(*input.WithDecryption)
			vs.EqualValues(10, *input.MaxResults)

			return true
		}),
		mock.AnythingOfType("func(*ssm.GetParametersByPathOutput, bool) bool"),
		[]request.Option(nil),
	).Return(nil)

	ret, err := vs.sut.ListVariables(vs.ctx, scope)
	vs.NoError(err)
	vs.Len(ret, 2)

	vs.Equal("PLAIN", ret[0].Name)
	vs.Equal("plain", ret[0].Value)
	vs.False(ret[0].WriteOnly)

	vs.Equal("SECRET", ret[1].Name)
	vs.Equal("secret", ret[1].Value)
	vs.True(ret[1].WriteOnly)
}

func (vs *variableRepositoryTestSuite) TestShowVariable() {
	const scope = "scope"
	const fullPath = "/testPrefix/variables/scope/NAME"

	vs.mockAPI.On(
		"GetParameterWithContext",
		vs.ctx,
		mock.MatchedBy(func(in interface{}) bool {
			input, ok := in.(*ssm.GetParameterInput)
			if !ok {
				return false
			}
			vs.Equal(fullPath, *input.Name)
			vs.True(*input.WithDecryption)
			return true
		}),
		[]request.Option(nil),
	).Return(
		&ssm.GetParameterOutput{
			Parameter: &ssm.Parameter{
				Name:  aws.String(fullPath),
				Type:  aws.String("SecureString"),
				Value: aws.String("secret"),
			},
		},
		nil,
	)

	ret, err := vs.sut.ShowVariable(vs.ctx, scope, "NAME")
	vs.NoError(err)

	vs.Equal("NAME", ret.Name)
	vs.Equal("secret", ret.Value)
	vs.True(ret.WriteOnly)
}

func (vs *variableRepositoryTestSuite) TestCreateVariablePlain() {
	const scope = "scope"

	variable := &Variable{Name: "NAME", Value: "value", WriteOnly: false}

	vs.mockAPI.On(
		"PutParameterWithContext",
		vs.ctx,
		mock.MatchedBy(func(in interface{}) bool {
			input, ok := in.(*ssm.PutParameterInput)
			if !ok {
				return false
			}
			vs.Equal("/testPrefix/variables/scope/NAME", *input.Name)
			vs.Equal("String", *input.Type)
			vs.Equal(variable.Value, *input.Value)
			vs.True(*input.Overwrite)
			vs.Nil(input.KeyId)
			return true
		}),
		[]request.Option(nil),
	).Return((*ssm.PutParameterOutput)(nil), nil)

	ret, err := vs.sut.CreateVariable(vs.ctx, scope, variable)
	vs.NoError(err)

	vs.Equal(variable.Name, ret.Name)
	vs.Equal(variable.Value, ret.Value)
	vs.False(ret.WriteOnly)
}

func (vs *variableRepositoryTestSuite) TestCreateVariablePlainSecret() {
	const scope = "scope"
	ctx := context.Background()

	variable := &Variable{Name: "NAME", Value: "value", WriteOnly: true}

	vs.mockAPI.On(
		"PutParameterWithContext",
		ctx,
		mock.MatchedBy(func(in interface{}) bool {
			input, ok := in.(*ssm.PutParameterInput)
			if !ok {
				return false
			}
			vs.Equal("SecureString", *input.Type)
			vs.Equal(testKMSKeyID, *input.KeyId)
			return true
		}),
		[]request.Option(nil),
	).Return((*ssm.PutParameterOutput)(nil), nil)

	ret, err := vs.sut.CreateVariable(ctx, scope, variable)
	vs.NoError(err)
	vs.True(ret.WriteOnly)
}

func (vs *variableRepositoryTestSuite) TestDeleteVariable() {
	const scope = "scope"
	const fullPath = "/testPrefix/variables/scope/NAME"
	ctx := context.Background()

	vs.mockAPI.On(
		"GetParameterWithContext",
		ctx,
		mock.MatchedBy(func(in interface{}) bool {
			input, ok := in.(*ssm.GetParameterInput)
			if !ok {
				return false
			}
			vs.Equal(fullPath, *input.Name)
			vs.True(*input.WithDecryption)
			return true
		}),
		[]request.Option(nil),
	).Return(
		&ssm.GetParameterOutput{
			Parameter: &ssm.Parameter{
				Name:  aws.String(fullPath),
				Type:  aws.String("SecureString"),
				Value: aws.String("secret"),
			},
		},
		nil,
	)

	vs.mockAPI.On(
		"DeleteParameterWithContext",
		ctx,
		mock.MatchedBy(func(in interface{}) bool {
			input, ok := in.(*ssm.DeleteParameterInput)
			if !ok {
				return false
			}
			vs.Equal(fullPath, *input.Name)
			return true
		}),
		[]request.Option(nil),
	).Return((*ssm.DeleteParameterOutput)(nil), nil)

	ret, err := vs.sut.DeleteVariable(ctx, scope, "NAME")
	vs.NoError(err)

	vs.Equal("NAME", ret.Name)
	vs.Equal("secret", ret.Value)
	vs.True(ret.WriteOnly)
}

func (vs *variableRepositoryTestSuite) TestResetOK() {
	const scope = "scope"

	vs.mockAPI.parameters = []*ssm.Parameter{{Name: aws.String("bacon")}}

	vs.withPagination(nil)
	vs.withDeletion([]*string{aws.String("bacon")}, nil)

	vs.NoError(vs.sut.Reset(vs.ctx, scope))
}

func (vs *variableRepositoryTestSuite) TestResetListError() {
	const scope = "scope"

	vs.withPagination(errors.New("bacon"))

	vs.EqualError(vs.sut.Reset(vs.ctx, scope), "bacon")
}

func (vs *variableRepositoryTestSuite) TestResetDeleteError() {
	const scope = "scope"

	vs.mockAPI.parameters = []*ssm.Parameter{{Name: aws.String("bacon")}}

	vs.withPagination(nil)
	vs.withDeletion([]*string{aws.String("bacon")}, errors.New("bacon"))

	vs.EqualError(vs.sut.Reset(vs.ctx, scope), "bacon")
}

func (vs *variableRepositoryTestSuite) withPagination(err error) {
	vs.mockAPI.On(
		"GetParametersByPathPagesWithContext",
		vs.ctx,
		mock.AnythingOfType("*ssm.GetParametersByPathInput"),
		mock.AnythingOfType("func(*ssm.GetParametersByPathOutput, bool) bool"),
		[]request.Option(nil),
	).Return(err)
}

func (vs *variableRepositoryTestSuite) withDeletion(names []*string, err error) {
	vs.mockAPI.On(
		"DeleteParametersWithContext",
		vs.ctx,
		&ssm.DeleteParametersInput{Names: names},
		[]request.Option(nil),
	).Return((*ssm.DeleteParametersOutput)(nil), err)
}

func TestVariableRepository(t *testing.T) {
	suite.Run(t, new(variableRepositoryTestSuite))
}
