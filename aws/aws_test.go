//go:build test
// +build test

package aws

import (
	"errors"
	"fmt"
	"os"
	"testing"

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

func (suite AWSSuite) TestNew() {
	e := new(mocks.Executor)
	aws := New(e)
	suite.Equal(e, aws.executor)
}

func (suite AWSSuite) TestFindCli() {
	e := new(mocks.Executor)
	a := New(e)
	e.On("FindExecutable", "aws").Return("aws", nil)
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
			findExecutableResult: "aws",
			execCommandError:     errors.New("exec command error"),
			expectedError:        true,
		},
	}

	for _, c := range cases {
		e := new(mocks.Executor)
		e.On("FindExecutable", "aws").Return(c.findExecutableResult, c.findExecutableError)
		e.On("ExecCommand", "aws", "configure", "list-profiles").Return(c.execCommandResult, c.execCommandError)
		a := New(e)
		res, err := a.findProfiles()
		if c.expectedError {
			suite.Error(err)
		} else {
			suite.NoError(err)
		}
		suite.Len(res, c.expectedResultLen)
	}
}

func (suite AWSSuite) TestProfileExists() {
	e := new(mocks.Executor)
	e.On("ExecCommand", "aws", "configure", "list-profiles").Return("a\nb\nc", nil)
	e.On("FindExecutable", "aws").Return("aws", nil)
	a := New(e)

	cases := []struct {
		profile string
		exists  bool
	}{
		{profile: "b", exists: true},
		{profile: "d"},
	}

	for _, c := range cases {
		res := a.profileExists(c.profile)
		suite.Equal(c.exists, res)
	}

	e = new(mocks.Executor)
	e.On("FindExecutable", "aws").Return("", errors.New("error"))
	a = New(e)
	ex := a.profileExists("a")
	suite.False(ex)
}

func (suite AWSSuite) TestCreateProfile() {
	cases := []struct {
		existingProfiles     string
		newProfile           string
		sso                  bool
		ssoError             error
		findExecutableError  error
		execCommandError     error
		execInteractiveError error
		readInputErr         error
		shouldErr            bool
	}{
		{
			existingProfiles: "1\n2\n3",
			newProfile:       "a",
		},
		{
			existingProfiles: "1\n2\n3",
			newProfile:       "a",
			sso:              true,
		},
		{
			existingProfiles: "1\n2\n3",
			newProfile:       "a",
			sso:              true,
			ssoError:         errors.New("test"),
			shouldErr:        true,
		},
		{
			existingProfiles: "a\n\b\nc",
			newProfile:       "a",
			shouldErr:        true,
		},
		{
			existingProfiles: "a\n\b\nc",
			newProfile:       "a b",
			shouldErr:        true,
		},
		{
			existingProfiles:     "a\n\b\nc",
			newProfile:           "ab",
			execInteractiveError: errors.New("test"),
			shouldErr:            true,
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
		{
			newProfile:           "ab",
			execInteractiveError: errors.New("test"),
			shouldErr:            true,
		},
	}

	for _, c := range cases {
		e := mocks.Executor{}
		e.On("FindExecutable", "aws").Return("aws", c.findExecutableError)
		e.On("ExecCommand", "aws", "configure", "list-profiles").Return(c.existingProfiles, c.execCommandError)
		e.On("ExecInteractive", "aws", "configure", "--profile", c.newProfile).Return(c.execInteractiveError)
		e.On("ExecInteractive", "aws", "configure", "--profile", c.newProfile, "sso").Return(c.execInteractiveError)
		if c.sso {
			e.On("PromptInput", "Use SSO? (only 'yes' will be accepted to approve): ").Return("yes", c.ssoError)
		} else {
			e.On("PromptInput", "Use SSO? (only 'yes' will be accepted to approve): ").Return("no", c.ssoError)
		}
		e.On("PromptInput", "AWS Profile Name: ").Return(c.newProfile, c.readInputErr)
		a := New(&e)

		res, err := a.CreateProfile()
		if c.shouldErr {
			suite.Error(err)
		} else {
			suite.Equal(c.newProfile, res)
		}
	}
}

func (suite AWSSuite) TestCreateKubeContext() {
	cases := []struct {
		region              string
		regionError         error
		clusterSelection    string
		clusterlist         string
		selectList          []string
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
			selectList:          []string{"a", "b"},
			selectedClusterName: "a",
			alias:               "newalias",
			expectedResult:      "newalias",
		},
		{
			region:              "us-east-1",
			clusterSelection:    "1",
			clusterlist:         `{"clusters":["a","b"]}`,
			selectList:          []string{"a", "b"},
			selectedClusterName: "a",
			expectedResult:      "arn:aws:eks:us-east-1:accountID:cluster/a",
		},
		{
			region:              "us-east-1",
			clusterSelection:    "1",
			clusterlist:         `{"clusters":["a","b"}`,
			selectList:          []string{"a", "b"},
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
			selectList:          []string{"a", "b"},
			selectedClusterName: "a",
			selectClusterError:  errors.New("select cluster error"),
			expectedError:       true,
		},
		{
			region:              "us-east-1",
			clusterSelection:    "1",
			clusterlist:         `{"clusters":["a","b"]}`,
			selectList:          []string{"a", "b"},
			selectedClusterName: "a",
			aliasError:          errors.New("select cluster error"),
			expectedError:       true,
		},
		{
			region:              "us-east-1",
			clusterSelection:    "1",
			clusterlist:         `{"clusters":["a","b"]}`,
			selectList:          []string{"a", "b"},
			selectedClusterName: "a",
			updateConfigError:   errors.New("update config error"),
			expectedError:       true,
		},
		{
			regionError:   errors.New("region error"),
			expectedError: true,
		},
		{
			region:        "us-east-1",
			clusterlist:   `{"clusters": []}`,
			expectedError: true,
		},
	}

	for _, c := range cases {
		e := new(mocks.Executor)
		fullClusterName := fmt.Sprintf("arn:aws:eks:%s:accountID:cluster/%s", c.region, c.selectedClusterName)
		e.On("FindExecutable", "aws").Return("aws", c.findExecutableError)
		e.On("PromptInput", "AWS Region: ").Return(c.region, c.regionError)
		e.On("PromptInput", "Kube Context Alias: ").Return(c.alias, c.aliasError)
		e.On("ExecCommand", "aws", "eks", "list-clusters", "--region", c.region).Return(c.clusterlist, c.listClustersError)
		e.On("ExecCommand", "aws", "eks", "update-kubeconfig", "--region", c.region, "--name", c.selectedClusterName).Return(fmt.Sprintf("Updated context %s in /home/user/.kube/config", fullClusterName), c.updateConfigError)
		e.On("ExecCommand", "aws", "eks", "update-kubeconfig", "--region", c.region, "--name", c.selectedClusterName, "--alias", c.alias).Return(fmt.Sprintf("Updated context %s in /home/user/.kube/config", c.alias), c.updateConfigError)
		e.On("SelectValueFromList", c.selectList, "Cluster", mock.Anything).Return(c.selectedClusterName, c.selectClusterError)
		a := New(e)

		res, err := a.CreateKubeContext()
		suite.Equal(c.expectedResult, res)
		if c.expectedError != false {
			suite.Error(err)
		} else {
			suite.NoError(err)
		}
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
		{expectedResult: "a"},
		{findProfilesError: errors.New("find profiles"), shouldError: true},
		{listProfilesError: errors.New("list profiles"), shouldError: true},
		{selectProfileError: errors.New("select profile"), shouldError: true},
	}

	for _, c := range cases {
		e := new(mocks.Executor)
		e.On("FindExecutable", "aws").Return("aws", c.findProfilesError)
		e.On("ExecCommand", "aws", "configure", "list-profiles").Return("a\nb\nc", c.listProfilesError)
		e.On("ReadInput").Return("1", c.selectProfileError)
		e.On("SelectValueFromList", []string{"a", "b", "c"}, "AWS Profile", mock.Anything).Return("a", c.selectProfileError)
		a := New(e)

		res, err := a.SelectProfile()
		if c.shouldError {
			suite.Error(err)
		} else {
			suite.NoError(err)
		}
		suite.Equal(c.expectedResult, res)
		suite.Equal(c.expectedResult, os.Getenv("AWS_PROFILE"))
		os.Unsetenv("AWS_PROFILE")
	}
}
