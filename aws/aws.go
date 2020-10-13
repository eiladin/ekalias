package aws

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/eiladin/ekalias/console"
	"github.com/logrusorgru/aurora/v3"
)

func findAWS() string {
	out, err := exec.LookPath("aws")
	if err != nil {
		log.Fatal(err)
	}
	return out
}

func findProfiles() []string {
	aws := findAWS()
	out := console.ExecCommand(aws, "configure", "list-profiles")
	profiles := strings.Split(out, "\n")
	return profiles
}

func profileExists(newProfile string) bool {
	for _, prof := range findProfiles() {
		if prof == newProfile {
			return true
		}
	}
	return false
}

func createNew() string {
	var newProfile string
	for newProfile == "" {
		fmt.Print("AWS Profile Name: ")
		r := console.ReadInput()
		if len(strings.Split(r, " ")) > 1 {
			fmt.Println(aurora.Red("invalid input, profile name cannot have spaces"))
		} else if profileExists(r) {
			fmt.Println(aurora.Red("invalid input, profile name already exists"))
		} else {
			newProfile = r
		}
	}
	aws := findAWS()

	cmd := &exec.Cmd{
		Path:   aws,
		Args:   []string{aws, "configure", "--profile", newProfile},
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return newProfile
}

type clusterlist struct {
	Clusters []string
}

func CreateKubeContext() string {
	var context string
	var region string
	aws := findAWS()
	for context == "" {
		fmt.Print("AWS Region: ")
		region := console.ReadInput()
		out := console.ExecCommand(aws, "eks", "list-clusters", "--region", region)
		cl := clusterlist{}
		err := json.Unmarshal([]byte(out), &cl)
		if err != nil {
			log.Fatal(err)
		}
		if len(cl.Clusters) > 0 {
			context = console.SelectValueFromList(cl.Clusters, "Cluster", nil)
		}
	}

	fmt.Print("Kube Context Alias: ")
	alias := console.ReadInput()
	var out string
	if len(alias) > 0 {
		out = console.ExecCommand(aws, "eks", "update-kubeconfig", "--region", region, "--name", context, "--alias", alias)
	} else {
		out = console.ExecCommand(aws, "eks", "update-kubeconfig", "--region", region, "--name", context)
	}
	s := strings.Split(out, " ")
	if s[0] == "Added" {
		return s[3]
	} else {
		return s[2]
	}
}

func SelectProfile() string {
	awsprofiles := findProfiles()
	selectedProfile := console.SelectValueFromList(awsprofiles, "AWS Profile", createNew)
	os.Setenv("AWS_PROFILE", selectedProfile)
	return selectedProfile
}
