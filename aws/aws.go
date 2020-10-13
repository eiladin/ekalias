package aws

import (
	"bufio"
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
	out, err := exec.Command(aws, "configure", "list-profiles").Output()
	if err != nil {
		log.Fatal(err)
	}
	profiles := strings.Split(string(out), "\n")
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
		reader := bufio.NewReader(os.Stdin)
		r, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		r = strings.Replace(r, "\n", "", -1)
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

func SelectProfile() string {
	awsprofiles := findProfiles()
	return console.SelectValueFromList(awsprofiles, "AWS Profile", createNew)
}
