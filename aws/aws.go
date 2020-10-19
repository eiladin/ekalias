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

func (aws AWS) FindCli() (string, error) {
	return aws.executor.FindExecutable("aws")
}

func (aws AWS) findProfiles() ([]string, error) {
	cli, err := aws.FindCli()
	if err != nil {
		return []string{}, err
	}
	out, err := aws.executor.ExecCommand(cli, "configure", "list-profiles")
	if err != nil {
		return []string{}, err
	}
	return strings.Split(out, "\n"), nil
}

func (aws AWS) profileExists(newProfile string) bool {
	profs, err := aws.findProfiles()
	if err != nil {
		return false
	}
	for _, prof := range profs {
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
		r, err := aws.executor.ReadInput()
		if err != nil {
			return "", err
		}
		if len(strings.Split(r, " ")) > 1 {
			return "", errors.New("invalid input, profile name cannot have spaces")
		} else if aws.profileExists(r) {
			return "", errors.New("invalid input, profile name already exists")
		} else {
			newProfile = r
		}
	}
	cli, err := aws.FindCli()
	if err != nil {
		return "", err
	}

	err = aws.executor.ExecInteractive(cli, "configure", "--profile", newProfile)
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
	cli, err := aws.FindCli()
	if err != nil {
		return "", err
	}
	for context == "" {
		fmt.Print("AWS Region: ")
		var err error
		region, err = aws.executor.ReadInput()
		if err != nil {
			return "", err
		}
		out, err := aws.executor.ExecCommand(cli, "eks", "list-clusters", "--region", region)
		if err != nil {
			return "", err
		}
		cl := clusterlist{}
		err = json.Unmarshal([]byte(out), &cl)
		if err != nil {
			return "", err
		}
		if len(cl.Clusters) > 0 {
			context, err = console.SelectValueFromList(aws.executor, cl.Clusters, "Cluster", nil)
			if err != nil {
				return "", err
			}
		}
	}

	fmt.Print("Kube Context Alias: ")
	alias, err := aws.executor.ReadInput()
	if err != nil {
		return "", err
	}
	args := []string{"eks", "update-kubeconfig", "--region", region, "--name", context}
	var contextName string

	if len(alias) > 0 {
		contextName = alias
		args = append(args, "--alias", alias)
	}
	out, err := aws.executor.ExecCommand(cli, args...)
	if err != nil {
		return "", err
	}
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

func (aws AWS) SelectProfile() (string, error) {
	awsprofiles, err := aws.findProfiles()
	if err != nil {
		return "", err
	}
	selectedProfile, err := console.SelectValueFromList(aws.executor, awsprofiles, "AWS Profile", aws.CreateProfile)
	if err != nil {
		return "", err
	}
	os.Setenv("AWS_PROFILE", selectedProfile)
	return selectedProfile, nil
}
