//go:build test
// +build test

package kubectl

import (
	"errors"
	"testing"

	"github.com/eiladin/ekalias/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type KubectlSuite struct {
	suite.Suite
}

func TestKubectlSuite(t *testing.T) {
	suite.Run(t, new(KubectlSuite))
}

func (suite KubectlSuite) TestNew() {
	e := new(mocks.Executor)
	k := New(e)
	suite.Equal(e, k.executor)
}

func (suite KubectlSuite) TestFindContexts() {
	cases := []struct {
		findExecutableResult string
		findExecutableError  error
		execCommandResult    string
		execCommandError     error
		expectedResultLen    int
		expectedError        bool
	}{
		{
			findExecutableResult: "kubectl",
			execCommandResult:    "a\nb\nc",
			expectedResultLen:    3,
		},
		{
			findExecutableResult: "",
			findExecutableError:  errors.New("find executable error"),
			execCommandResult:    "a\nb\nc",
			expectedError:        true,
		},
		{
			findExecutableResult: "kubectl",
			execCommandResult:    "",
			execCommandError:     errors.New("exec command error"),
			expectedError:        true,
		},
	}

	for _, c := range cases {
		e := new(mocks.Executor)
		e.On("FindExecutable", "kubectl").Return(c.findExecutableResult, c.findExecutableError)
		e.On("ExecCommand", "kubectl", "config", "get-contexts", "-o", "name").Return(c.execCommandResult, c.execCommandError)
		k := New(e)

		res, err := k.findContexts()
		if c.expectedError {
			suite.Error(err)
		} else {
			suite.NoError(err)
		}
		suite.Len(res, c.expectedResultLen)
	}
}

func (suite KubectlSuite) TestSelectContext() {
	e := new(mocks.Executor)
	e.On("FindExecutable", "kubectl").Return("kubectl", nil)
	e.On("ExecCommand", "kubectl", "config", "get-contexts", "-o", "name").Return("a\nb\nc", nil)
	e.On("ReadInput").Return("2", nil)
	e.On("SelectValueFromList", []string{"a", "b", "c"}, "Kube Context", mock.Anything).Return("b", nil)
	k := New(e)
	res, err := k.SelectContext()
	suite.NoError(err)
	suite.Equal("b", res)

	e = new(mocks.Executor)
	e.On("FindExecutable", "kubectl").Return("", errors.New("error"))
	k = New(e)
	res, err = k.SelectContext()
	suite.Error(err)
	suite.Empty(res)
}

func (suite KubectlSuite) TestFindCli() {
	cases := []struct {
		cmd string
		err error
	}{
		{cmd: "kubectl"},
		{err: errors.New("error")},
	}

	for _, c := range cases {
		e := new(mocks.Executor)
		e.On("FindExecutable", "kubectl").Return(c.cmd, c.err)
		k := New(e)
		res, err := k.FindCli()
		if c.err == nil {
			suite.NoError(err)
			suite.Equal(c.cmd, res)
		} else {
			suite.Error(err)
			suite.Empty(res)
		}
	}
}
