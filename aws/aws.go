package aws

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/eiladin/ekalias/console"
)

var executor console.Executor = console.DefaultExecutor{}

func FindAWS() string {
	return executor.FindExecutable("aws")
}

func findProfiles() []string {
	aws := FindAWS()
	out := executor.ExecCommand(aws, "configure", "list-profiles")
	return strings.Split(out, "\n")
}

func profileExists(newProfile string) bool {
	for _, prof := range findProfiles() {
		if prof == newProfile {
			return true
		}
	}
	return false
}

func createNew() (string, error) {
	var newProfile string
	for newProfile == "" {
		fmt.Print("AWS Profile Name: ")
		r := executor.ReadInput()
		if len(strings.Split(r, " ")) > 1 {
			return "", errors.New("invalid input, profile name cannot have spaces")
		} else if profileExists(r) {
			return "", errors.New("invalid input, profile name already exists")
		} else {
			newProfile = r
		}
	}
	aws := FindAWS()

	err := executor.ExecInteractive(aws, "configure", "--profile", newProfile)
	if err != nil {
		return "", err
	}
	return newProfile, nil
}

type clusterlist struct {
	Clusters []string
}

func CreateKubeContext() (string, error) {
	var context string
	var region string
	aws := FindAWS()
	for context == "" {
		fmt.Print("AWS Region: ")
		region = executor.ReadInput()
		out := executor.ExecCommand(aws, "eks", "list-clusters", "--region", region)
		cl := clusterlist{}
		err := json.Unmarshal([]byte(out), &cl)
		if err != nil {
			return "", err
		}
		if len(cl.Clusters) > 0 {
			context = console.SelectValueFromList(executor, cl.Clusters, "Cluster", nil)
		}
	}

	fmt.Print("Kube Context Alias: ")
	alias := executor.ReadInput()
	args := []string{"eks", "update-kubeconfig", "--region", region, "--name", context}
	var contextName string

	if len(alias) > 0 {
		contextName = alias
		args = append(args, "--alias", alias)
	}
	out := executor.ExecCommand(aws, args...)
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

func SelectProfile() string {
	awsprofiles := findProfiles()
	selectedProfile := console.SelectValueFromList(executor, awsprofiles, "AWS Profile", createNew)
	os.Setenv("AWS_PROFILE", selectedProfile)
	return selectedProfile
}
