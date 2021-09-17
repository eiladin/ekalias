//go:build test
// +build test

package kubectl

import (
	"errors"
	"os"
	"testing"

	"github.com/eiladin/ekalias/console"
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
	me := mocks.NewExecutor()
	cases := []struct {
		executor       console.Executor
		expectedResult console.Executor
	}{
		{
			executor:       nil,
			expectedResult: console.DefaultExecutor{Stdin: os.Stdin},
		},
		{
			executor:       me,
			expectedResult: me,
		},
	}

	for _, c := range cases {
		k := New(c.executor)
		suite.Equal(c.expectedResult, k.executor)
	}
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
			findExecutableError:  nil,
			execCommandResult:    "a\nb\nc",
			execCommandError:     nil,
			expectedResultLen:    3,
			expectedError:        false,
		},
		{
			findExecutableResult: "",
			findExecutableError:  errors.New("find executable error"),
			execCommandResult:    "a\nb\nc",
			execCommandError:     nil,
			expectedResultLen:    0,
			expectedError:        true,
		},
		{
			findExecutableResult: "kubectl",
			findExecutableError:  nil,
			execCommandResult:    "",
			execCommandError:     errors.New("exec command error"),
			expectedResultLen:    0,
			expectedError:        true,
		},
	}

	for _, c := range cases {
		e := new(mocks.Executor)
		e.On("FindExecutable", "kubectl").Return(c.findExecutableResult, c.findExecutableError)
		e.On("ExecCommand", "kubectl", "config", "get-contexts", "-o", "name").Return(c.execCommandResult, c.execCommandError)
		k := New(e)
		mocks.ReadStdOut(func() {
			res, err := k.findContexts()
			if c.expectedError {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
			suite.Len(res, c.expectedResultLen)
		})
	}
}

func (suite KubectlSuite) TestSelectContext() {
	e := mocks.NewExecutor()
	e.On("ExecCommand", "kubectl", "config", "get-contexts", "-o", "name").Return("a\nb\nc", nil)
	e.On("ReadInput").Return("2", nil)
	e.On("SelectValueFromList", []string{"a", "b", "c"}, "Kube Context", mock.Anything).Return("b", nil)
	k := New(e)

	mocks.ReadStdOut(func() {
		res, err := k.SelectContext()
		suite.NoError(err)
		suite.Equal("b", res)
	})

	e = new(mocks.Executor)
	e.On("FindExecutable", "kubectl").Return("", errors.New("error"))
	k = New(e)
	res, err := k.SelectContext()
	suite.Error(err)
	suite.Empty(res)
}

func (suite KubectlSuite) TestFindCli() {
	k := New(mocks.NewExecutor())
	res, err := k.FindCli()
	suite.NoError(err)
	suite.Equal("kubectl", res)

	e := new(mocks.Executor)
	e.On("FindExecutable", "kubectl").Return("", errors.New("error"))
	k = New(e)
	res, err = k.FindCli()
	suite.Error(err)
	suite.Empty(res)
}
