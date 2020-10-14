package console

import (
	"errors"
	"io/ioutil"
	"log"
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
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		log.Fatal(err)
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		log.Fatal(err)
	}

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = tmpfile

	e := DefaultExecutor{}
	res := e.ReadInput()
	suite.Equal("test input", res)
}

func (suite ConsoleSuite) TestExecCommand() {
	e := DefaultExecutor{}
	res := e.ExecCommand("echo", "hello", "world")
	suite.Equal("hello world\n", res)
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
	res := e.FindExecutable("echo")
	suite.Contains(res, "echo")
}

func (suite ConsoleSuite) TestSelectValueFromList() {
	list := []string{"a", "b", "c"}
	selection := "1"
	expected := "a"

	e := mocks.NewExecutor()
	e.On("ReadInput").Return(selection)

	mocks.ReadStdOut(func() {
		res := SelectValueFromList(e, list, "test item", func() (string, error) { return "new item", nil })
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
				readInput.ReturnArguments = mock.Arguments{c.initialSelection}
				callcounter = 1
			} else {
				readInput.ReturnArguments = mock.Arguments{c.finalSelection}
			}
		}

		out := mocks.ReadStdOut(func() {
			res := SelectValueFromList(e, c.list, "test item", c.newItemFunc)
			suite.Equal(res, c.expected)
		})
		if c.errorExpected != "" {
			suite.Contains(out, c.errorExpected, "case number: %d", i)
		}
	}
}
