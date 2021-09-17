package aws

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/eiladin/ekalias/console"
)

var ErrProfileSpaces = errors.New("profile name cannot have spaces")
var ErrProfileExists = errors.New("profile name already exists")
var ErrNoClusters = errors.New("no clusters in selected account/region")

type AWS struct {
	executor console.Executor
}

func New(e console.Executor) AWS {
	return AWS{executor: e}
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
	sso := false

	cli, err := aws.FindCli()
	if err != nil {
		return "", err
	}

	var newProfile string
	for newProfile == "" {
		r, err := aws.executor.PromptInput("Use SSO? (only 'yes' will be accepted to approve): ")
		if err != nil {
			return "", err
		}
		sso = r == "yes"

		r, err = aws.executor.PromptInput("AWS Profile Name: ")
		if err != nil {
			return "", err
		}

		switch {
		case len(strings.Split(r, " ")) > 1:
			return "", ErrProfileSpaces
		case aws.profileExists(r):
			return "", ErrProfileExists
		default:
			newProfile = r
		}
	}

	args := []string{"configure", "--profile", newProfile}
	if sso {
		args = append(args, "sso")
	}

	err = aws.executor.ExecInteractive(cli, args...)
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

	region, err = aws.executor.PromptInput("AWS Region: ")
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
	if len(cl.Clusters) == 0 {
		return "", ErrNoClusters
	}

	for context == "" {
		context, err = aws.executor.SelectValueFromList(cl.Clusters, "Cluster", nil)
		if err != nil {
			return "", err
		}
	}

	alias, err := aws.executor.PromptInput("Kube Context Alias: ")
	if err != nil {
		return "", err
	}

	args := []string{"eks", "update-kubeconfig", "--region", region, "--name", context}
	var contextName string

	if len(alias) > 0 {
		contextName = alias
		args = append(args, "--alias", alias)
	}

	out, err = aws.executor.ExecCommand(cli, args...)
	if err != nil {
		return "", err
	}

	if contextName == "" {
		s := strings.Split(out, " ")
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

	selectedProfile, err := aws.executor.SelectValueFromList(awsprofiles, "AWS Profile", aws.CreateProfile)
	if err != nil {
		return "", err
	}

	os.Setenv("AWS_PROFILE", selectedProfile)
	return selectedProfile, nil
}
