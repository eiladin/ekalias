// +build test

package aws

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/eiladin/ekalias/console"

	"github.com/eiladin/ekalias/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AWSSuite struct {
	suite.Suite
}

func TestAWSSuite(t *testing.T) {
	suite.Run(t, new(AWSSuite))
}

func (suite AWSSuite) TestCreate() {
	e := mocks.NewExecutor()
	cases := []struct {
		executor       console.Executor
		expectedResult console.Executor
	}{
		{
			executor:       nil,
			expectedResult: console.DefaultExecutor{},
		},
		{
			executor:       e,
			expectedResult: e,
		},
	}

	for _, c := range cases {
		aws := Create(c.executor)
		suite.Equal(c.expectedResult, aws.executor)
	}
}

func (suite AWSSuite) TestFindCli() {
	a := Create(mocks.NewExecutor())
	res, err := a.FindCli()
	suite.NoError(err)
	suite.Equal("aws", res)
}

func (suite AWSSuite) TestFindProfiles() {
	cases := []struct {
		findExecutableResult string
		findExecutableError  error
		execCommandResult    string
		execCommandError     error
		expectedResultLen    int
		expectedError        bool
	}{
		{
			findExecutableResult: "aws",
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
			findExecutableResult: "aws",
			findExecutableError:  nil,
			execCommandResult:    "",
			execCommandError:     errors.New("exec command error"),
			expectedResultLen:    0,
			expectedError:        true,
		},
	}

	for _, c := range cases {
		e := new(mocks.Executor)
		e.On("FindExecutable", "aws").Return(c.findExecutableResult, c.findExecutableError)
		e.On("ExecCommand", "aws", "configure", "list-profiles").Return(c.execCommandResult, c.execCommandError)
		a := Create(e)
		mocks.ReadStdOut(func() {
			res, err := a.findProfiles()
			if c.expectedError {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
			suite.Len(res, c.expectedResultLen)
		})
	}
}

func (suite AWSSuite) TestProfileExists() {
	e := mocks.NewExecutor()
	e.On("ExecCommand", "aws", "configure", "list-profiles").Return("a\nb\nc", nil)
	a := Create(e)

	cases := []struct {
		profile string
		exists  bool
	}{
		{profile: "b", exists: true},
		{profile: "d", exists: false},
	}

	for _, c := range cases {
		res := a.profileExists(c.profile)
		suite.Equal(c.exists, res)
	}

	e = new(mocks.Executor)
	e.On("FindExecutable", "aws").Return("", errors.New("error"))
	a = Create(e)
	ex := a.profileExists("a")
	suite.False(ex)
}

func (suite AWSSuite) TestCreateProfile() {
	cases := []struct {
		existingProfiles     string
		newProfile           string
		findExecutableError  error
		execCommandError     error
		execInteractiveError error
		interactiveErr       error
		readInputErr         error
		shouldErr            bool
	}{
		{
			existingProfiles: "1\n2\n3",
			newProfile:       "a",
			interactiveErr:   nil,
			shouldErr:        false,
		},
		{
			existingProfiles: "a\n\b\nc",
			newProfile:       "a",
			interactiveErr:   nil,
			shouldErr:        true,
		},
		{
			existingProfiles: "a\n\b\nc",
			newProfile:       "a b",
			interactiveErr:   nil,
			shouldErr:        true,
		},
		{
			existingProfiles: "a\n\b\nc",
			newProfile:       "ab",
			interactiveErr:   errors.New("test"),
			shouldErr:        true,
		},
		{
			readInputErr: errors.New("test"),
			shouldErr:    true,
		},
		{
			newProfile:          "ab",
			findExecutableError: errors.New("test"),
			shouldErr:           true,
		},
	}

	for _, c := range cases {
		e := new(mocks.Executor)
		e.On("FindExecutable", "aws").Return("aws", c.findExecutableError)
		e.On("ExecCommand", "aws", "configure", "list-profiles").Return(c.existingProfiles, c.execCommandError)
		e.On("ExecInteractive", "aws", "configure", "--profile", c.newProfile).Return(c.interactiveErr, c.execInteractiveError)
		e.On("ReadInput").Return(c.newProfile, c.readInputErr)
		a := Create(e)

		mocks.ReadStdOut(func() {
			res, err := a.CreateProfile()
			if c.shouldErr {
				suite.Error(err)
			} else {
				suite.Equal(c.newProfile, res)
			}
		})
	}
}

func (suite AWSSuite) TestCreateKubeContext() {
	cases := []struct {
		region              string
		clusterSelection    string
		clusterlist         string
		selectedClusterName string
		alias               string
		findExecutableError error
		readInputError      error
		listClustersError   error
		selectClusterError  error
		aliasError          error
		updateConfigError   error
		expectedError       bool
		expectedResult      string
	}{
		{
			region:              "us-east-1",
			clusterSelection:    "1",
			clusterlist:         `{"clusters":["a","b"]}`,
			selectedClusterName: "a",
			alias:               "newalias",
			expectedResult:      "newalias",
		},
		{
			region:              "us-east-1",
			clusterSelection:    "1",
			clusterlist:         `{"clusters":["a","b"]}`,
			selectedClusterName: "a",
			expectedResult:      "arn:aws:eks:us-east-1:accountID:cluster/a",
		},
		{
			region:              "us-east-1",
			clusterSelection:    "1",
			clusterlist:         `{"clusters":["a","b"}`,
			selectedClusterName: "a",
			expectedError:       true,
		},
		{
			findExecutableError: errors.New("find executable"),
			expectedError:       true,
		},
		{
			readInputError: errors.New("read input"),
			expectedError:  true,
		},
		{
			listClustersError: errors.New("list clusters"),
			expectedError:     true,
		},
		{
			region:              "us-east-1",
			clusterSelection:    "1",
			clusterlist:         `{"clusters":["a","b"]}`,
			selectedClusterName: "a",
			selectClusterError:  errors.New("select cluster error"),
			expectedError:       true,
		},
		{
			region:              "us-east-1",
			clusterSelection:    "1",
			clusterlist:         `{"clusters":["a","b"]}`,
			selectedClusterName: "a",
			aliasError:          errors.New("select cluster error"),
			expectedError:       true,
		},
		{
			region:              "us-east-1",
			clusterSelection:    "1",
			clusterlist:         `{"clusters":["a","b"]}`,
			selectedClusterName: "a",
			updateConfigError:   errors.New("update config error"),
			expectedError:       true,
			expectedResult:      "",
		},
	}

	for _, c := range cases {
		e := new(mocks.Executor)
		inputs := []string{c.region, c.clusterSelection, c.alias}
		callcounter := 0
		readInput := e.On("ReadInput").Times(3)
		readInput.RunFn = func(args mock.Arguments) {
			if callcounter == 1 {
				readInput.ReturnArguments = mock.Arguments{inputs[callcounter], c.selectClusterError}
			} else if callcounter == 2 {
				readInput.ReturnArguments = mock.Arguments{inputs[callcounter], c.aliasError}
			} else {
				readInput.ReturnArguments = mock.Arguments{inputs[callcounter], c.readInputError}
			}
			callcounter++
		}
		e.On("FindExecutable", "aws").Return("aws", c.findExecutableError)
		e.On("ExecCommand", "aws", "eks", "list-clusters", "--region", c.region).Return(c.clusterlist, c.listClustersError)
		e.On("ExecCommand", "aws", "eks", "update-kubeconfig", "--region", c.region, "--name", c.selectedClusterName).Return(fmt.Sprintf("Updated context arn:aws:eks:%s:accountID:cluster/%s in /home/user/.kube/config", c.region, c.selectedClusterName), c.updateConfigError)
		e.On("ExecCommand", "aws", "eks", "update-kubeconfig", "--region", c.region, "--name", c.selectedClusterName, "--alias", c.alias).Return(fmt.Sprintf("Updated context %s in /home/user/.kube/config", c.alias), nil)
		a := Create(e)
		mocks.ReadStdOut(func() {
			res, err := a.CreateKubeContext()
			suite.Equal(res, c.expectedResult)
			if c.expectedError != false {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite AWSSuite) TestSelectProfile() {
	cases := []struct {
		findProfilesError  error
		listProfilesError  error
		selectProfileError error
		shouldError        bool
		expectedResult     string
	}{
		{
			expectedResult: "a",
		},
		{
			findProfilesError: errors.New("find profiles"),
			shouldError:       true,
		},
		{
			listProfilesError: errors.New("list profiles"),
			shouldError:       true,
		},
		{
			selectProfileError: errors.New("select profile"),
			shouldError:        true,
		},
	}

	for _, c := range cases {
		e := new(mocks.Executor)
		e.On("FindExecutable", "aws").Return("aws", c.findProfilesError)
		e.On("ExecCommand", "aws", "configure", "list-profiles").Return("a\nb\nc", c.listProfilesError)
		e.On("ReadInput").Return("1", c.selectProfileError)
		a := Create(e)
		mocks.ReadStdOut(func() {
			res, err := a.SelectProfile()
			if c.shouldError {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
			suite.Equal(c.expectedResult, res)
			suite.Equal(c.expectedResult, os.Getenv("AWS_PROFILE"))
			os.Unsetenv("AWS_PROFILE")
		})
	}

}
