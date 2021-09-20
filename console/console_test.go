package console

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConsoleSuite struct {
	suite.Suite
}

func TestConsoleSuite(t *testing.T) {
	suite.Run(t, new(ConsoleSuite))
}

func (suite ConsoleSuite) TestNew() {
	b := bytes.Buffer{}
	cases := []struct {
		reader   io.Reader
		expected io.Reader
	}{
		{expected: nil},
		{reader: &b, expected: &b},
	}

	for _, c := range cases {
		e := New(c.reader, nil, nil).(DefaultExecutor)
		suite.Equal(c.expected, e.Stdin)
	}
}

func (suite ConsoleSuite) TestBuildAlias() {
	res := BuildAlias("aliasname", "profile", "context")
	suite.Equal(`alias aliasname="export AWS_PROFILE=profile && kubectl config use-context context"`, res)
}

func (suite ConsoleSuite) TestReadInput() {
	content := "test input"

	stdin := mockReader{
		list: []string{content},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	e := New(&stdin, &stdout, &stderr)

	res, err := e.ReadInput()
	suite.Equal("test input", res)
	suite.NoError(err)
}

func (suite ConsoleSuite) TestExecCommand() {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	e := New(nil, &stdout, &stderr)

	path, err := exec.LookPath("echo")
	suite.NoError(err)

	res, err := e.ExecCommand(path, "hello", "world")
	suite.NoError(err)
	suite.Equal("hello world\n", res)

	res, err = e.ExecCommand("echo1")
	suite.Error(err)
	suite.Empty(res)
}

func (suite ConsoleSuite) TestExecInteractive() {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	e := New(nil, &stdout, &stderr)

	path, err := exec.LookPath("echo")
	suite.NoError(err)

	err = e.ExecInteractive(path, "hello", "world")
	suite.NoError(err)
	suite.Equal("hello world\n", stdout.String())
}

func (suite ConsoleSuite) TestFindExecutable() {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	e := New(nil, &stdout, &stderr)

	res, err := e.FindExecutable("echo")
	suite.NoError(err)
	suite.Contains(res, "echo")

	res, err = e.FindExecutable("echo1")
	suite.Error(err)
	suite.Empty(res)
}

func (suite ConsoleSuite) TestSelectValueFromList() {
	cases := []struct {
		list          []string
		selection     string
		expected      string
		errorExpected bool
	}{
		{
			list:      []string{"a", "b", "c"},
			selection: "1",
			expected:  "a",
		},
		{
			errorExpected: true,
		},
	}

	for _, c := range cases {
		stdin := mockReader{
			list: []string{c.selection},
		}

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		de := New(&stdin, &stdout, &stderr)
		res, err := de.SelectValueFromList(c.list, "test item", func() (string, error) { return "new item", nil })
		if c.errorExpected {
			suite.Error(err)
		} else {
			suite.NoError(err)
			suite.Equal(res, c.expected)
		}
	}
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
		stdin := mockReader{
			list: []string{c.initialSelection, c.finalSelection},
		}
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		e := New(&stdin, &stdout, &stderr)

		res, err := e.SelectValueFromList(c.list, "test item", c.newItemFunc)
		suite.NoError(err, "case number: %d", i)
		suite.Equal(c.expected, res, "case number: %d", i)

		if c.errorExpected != "" {
			suite.Contains(stdout.String(), c.errorExpected)
		}
	}
}

type mockReader struct {
	list []string
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	if len(m.list) == 0 {
		return 0, io.EOF
	}
	n = copy(p, []byte(m.list[0]+"\n"))
	m.list = m.list[1:]
	return
}
