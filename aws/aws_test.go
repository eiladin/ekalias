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

func (suite AWSSuite) TestFindCli() {
	a := Create(mocks.NewExecutor())
	res := a.FindCli()
	suite.Equal("aws", res)
}

func (suite AWSSuite) TestFindProfiles() {
	e := mocks.NewExecutor()
	e.On("ExecCommand", "aws", "configure", "list-profiles").Return("a\nb\nc")
	a := Create(e)

	res := a.findProfiles()
	suite.Equal(3, len(res))
}

func (suite AWSSuite) TestProfileExists() {
	e := mocks.NewExecutor()
	e.On("ExecCommand", "aws", "configure", "list-profiles").Return("a\nb\nc")
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
}

func (suite AWSSuite) TestCreateProfile() {
	cases := []struct {
		existingProfiles string
		newProfile       string
		interactiveErr   error
		shouldErr        bool
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
	}

	for _, c := range cases {
		e := mocks.NewExecutor()
		e.On("ExecCommand", "aws", "configure", "list-profiles").Return(c.existingProfiles)
		e.On("ExecInteractive", "aws", "configure", "--profile", c.newProfile).Return(c.interactiveErr)
		e.On("ReadInput").Return(c.newProfile)
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
		expectedError       bool
		expectedResult      string
	}{
		{
			region:              "us-east-1",
			clusterSelection:    "1",
			clusterlist:         `{"clusters":["a","b"]}`,
			selectedClusterName: "a",
			alias:               "newalias",
			expectedError:       false,
			expectedResult:      "newalias",
		},
		{
			region:              "us-east-1",
			clusterSelection:    "1",
			clusterlist:         `{"clusters":["a","b"]}`,
			selectedClusterName: "a",
			alias:               "",
			expectedError:       false,
			expectedResult:      "arn:aws:eks:us-east-1:accountID:cluster/a",
		},
		{
			region:              "us-east-1",
			clusterSelection:    "1",
			clusterlist:         `{"clusters":["a","b"}`,
			selectedClusterName: "a",
			alias:               "",
			expectedError:       true,
			expectedResult:      "",
		},
	}

	for _, c := range cases {
		e := mocks.NewExecutor()
		inputs := []string{c.region, c.clusterSelection, c.alias}
		callcounter := 0
		readInput := e.On("ReadInput").Times(3)
		readInput.RunFn = func(args mock.Arguments) {
			readInput.ReturnArguments = mock.Arguments{inputs[callcounter]}
			callcounter++
		}
		e.On("ExecCommand", "aws", "eks", "list-clusters", "--region", c.region).Return(c.clusterlist)
		e.On("ExecCommand", "aws", "eks", "update-kubeconfig", "--region", c.region, "--name", c.selectedClusterName).Return(fmt.Sprintf("Updated context arn:aws:eks:%s:accountID:cluster/%s in /home/user/.kube/config", c.region, c.selectedClusterName))
		e.On("ExecCommand", "aws", "eks", "update-kubeconfig", "--region", c.region, "--name", c.selectedClusterName, "--alias", c.alias).Return(fmt.Sprintf("Updated context %s in /home/user/.kube/config", c.alias))
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
	e := mocks.NewExecutor()
	e.On("ExecCommand", "aws", "configure", "list-profiles").Return("a\nb\nc")
	e.On("ReadInput").Return("1")
	a := Create(e)

	mocks.ReadStdOut(func() {
		res := a.SelectProfile()
		suite.Equal("a", res)
		suite.Equal("a", os.Getenv("AWS_PROFILE"))
	})
}
