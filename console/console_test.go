package console

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/eiladin/ekalias/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ConsoleSuite struct {
	suite.Suite
}

func TestConsoleSuite(t *testing.T) {
	suite.Run(t, new(ConsoleSuite))
}

func (suite ConsoleSuite) TestBuildAlias() {
	res := BuildAlias("aliasname", "profile", "context")
	suite.Equal(`alias aliasname="export AWS_PROFILE=profile && kubectl config use-context context"`, res)
}

func (suite ConsoleSuite) TestReadInput() {
	content := []byte("test input\n")
	tmpfile, err := ioutil.TempFile("", "example")
	suite.NoError(err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	suite.NoError(err)

	_, err = tmpfile.Seek(0, 0)
	suite.NoError(err)

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = tmpfile

	e := DefaultExecutor{}
	res, err := e.ReadInput()
	suite.Equal("test input", res)
	suite.NoError(err)
}

func (suite ConsoleSuite) TestExecCommand() {
	e := DefaultExecutor{}
	res, err := e.ExecCommand("echo", "hello", "world")
	suite.NoError(err)
	suite.Equal("hello world\n", res)

	res, err = e.ExecCommand("echo1")
	suite.Error(err)
	suite.Empty(res)
}

func (suite ConsoleSuite) TestExecInteractive() {
	e := DefaultExecutor{}
	path, err := exec.LookPath("echo")
	suite.NoError(err)

	res := mocks.ReadStdOut(func() {
		err := e.ExecInteractive(path, "hello", "world")
		suite.NoError(err)
	})

	suite.Equal("hello world\n", res)
}

func (suite ConsoleSuite) TestFindExecutable() {
	e := DefaultExecutor{}
	res, err := e.FindExecutable("echo")
	suite.NoError(err)
	suite.Contains(res, "echo")

	res, err = e.FindExecutable("echo1")
	suite.Error(err)
	suite.Empty(res)
}

func (suite ConsoleSuite) TestSelectValueFromList() {
	list := []string{"a", "b", "c"}
	selection := "1"
	expected := "a"

	e := mocks.NewExecutor()
	e.On("ReadInput").Return(selection, nil)

	mocks.ReadStdOut(func() {
		res, err := SelectValueFromList(e, list, "test item", func() (string, error) { return "new item", nil })
		suite.NoError(err)
		suite.Equal(res, expected)
	})
}

func (suite ConsoleSuite) TestSelectValueFromListInvalidSelection() {
	newCalls := 0
	cases := []struct {
		list             []string
		initialSelection string
		finalSelection   string
		expected         string
		errorExpected    string
		newItemFunc      func() (string, error)
	}{
		{
			list:             []string{"a", "b", "c"},
			initialSelection: "-1",
			finalSelection:   "1",
			expected:         "a",
			errorExpected:    "invalid input -- valid selections: 1-3",
		},
		{
			list:             []string{"a", "b", "c"},
			initialSelection: "a",
			finalSelection:   "1",
			expected:         "a",
			errorExpected:    "invalid input -- valid selections: 1-3",
		},
		{
			list:             []string{"a", "b", "c"},
			initialSelection: "4",
			finalSelection:   "4",
			expected:         "new item",
			errorExpected:    "",
			newItemFunc: func() (string, error) {
				if newCalls == 2 {
					return "new item", nil
				}
				newCalls++
				return "", errors.New("error")
			},
		},
	}

	for i, c := range cases {
		e := mocks.NewExecutor()
		callcounter := 0
		readInput := e.On("ReadInput").Times(2)
		readInput.RunFn = func(args mock.Arguments) {
			if callcounter == 0 {
				readInput.ReturnArguments = mock.Arguments{c.initialSelection, nil}
				callcounter = 1
			} else {
				readInput.ReturnArguments = mock.Arguments{c.finalSelection, nil}
			}
		}

		out := mocks.ReadStdOut(func() {
			res, err := SelectValueFromList(e, c.list, "test item", c.newItemFunc)
			suite.NoError(err)
			suite.Equal(res, c.expected)
		})
		if c.errorExpected != "" {
			suite.Contains(out, c.errorExpected, "case number: %d", i)
		}
	}

	e := mocks.NewExecutor()
	e.On("ReadInput").Return("", errors.New("error"))
	mocks.ReadStdOut(func() {
		res, err := SelectValueFromList(e, []string{"a", "b"}, "test", func() (string, error) {
			return "", nil
		})
		suite.Error(err)
		suite.Empty(res)
	})
}
