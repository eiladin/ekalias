package aws

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/eiladin/ekalias/console"
)

type AWS struct {
	executor console.Executor
}

func Create(e console.Executor) AWS {
	aws := AWS{
		executor: e,
	}
	if aws.executor == nil {
		aws.executor = console.DefaultExecutor{}
	}
	return aws
}

func (aws AWS) FindCli() string {
	return aws.executor.FindExecutable("aws")
}

func (aws AWS) findProfiles() []string {
	cli := aws.FindCli()
	out := aws.executor.ExecCommand(cli, "configure", "list-profiles")
	return strings.Split(out, "\n")
}

func (aws AWS) profileExists(newProfile string) bool {
	for _, prof := range aws.findProfiles() {
		if prof == newProfile {
			return true
		}
	}
	return false
}

func (aws AWS) CreateProfile() (string, error) {
	var newProfile string
	for newProfile == "" {
		fmt.Print("AWS Profile Name: ")
		r := aws.executor.ReadInput()
		if len(strings.Split(r, " ")) > 1 {
			return "", errors.New("invalid input, profile name cannot have spaces")
		} else if aws.profileExists(r) {
			return "", errors.New("invalid input, profile name already exists")
		} else {
			newProfile = r
		}
	}
	cli := aws.FindCli()

	err := aws.executor.ExecInteractive(cli, "configure", "--profile", newProfile)
	if err != nil {
		return "", err
	}
	return newProfile, nil
}

type clusterlist struct {
	Clusters []string
}

func (aws AWS) CreateKubeContext() (string, error) {
	var context string
	var region string
	cli := aws.FindCli()
	for context == "" {
		fmt.Print("AWS Region: ")
		region = aws.executor.ReadInput()
		out := aws.executor.ExecCommand(cli, "eks", "list-clusters", "--region", region)
		cl := clusterlist{}
		err := json.Unmarshal([]byte(out), &cl)
		if err != nil {
			return "", err
		}
		if len(cl.Clusters) > 0 {
			context = console.SelectValueFromList(aws.executor, cl.Clusters, "Cluster", nil)
		}
	}

	fmt.Print("Kube Context Alias: ")
	alias := aws.executor.ReadInput()
	args := []string{"eks", "update-kubeconfig", "--region", region, "--name", context}
	var contextName string

	if len(alias) > 0 {
		contextName = alias
		args = append(args, "--alias", alias)
	}
	out := aws.executor.ExecCommand(cli, args...)
	s := strings.Split(out, " ")

	if contextName == "" {
		for _, item := range s {
			if strings.HasPrefix(item, "arn:aws") {
				contextName = item
			}
		}
	}
	return contextName, nil
}

func (aws AWS) SelectProfile() string {
	awsprofiles := aws.findProfiles()
	selectedProfile := console.SelectValueFromList(aws.executor, awsprofiles, "AWS Profile", aws.CreateProfile)
	os.Setenv("AWS_PROFILE", selectedProfile)
	return selectedProfile
}
