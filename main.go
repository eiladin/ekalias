package main

import (
	"fmt"
	"log"
	"os"

	"github.com/eiladin/ekalias/aws"
	"github.com/eiladin/ekalias/console"
	"github.com/eiladin/ekalias/kubectl"
	"github.com/logrusorgru/aurora/v3"
)

func main() {
	k := kubectl.FindKubectl()
	if k == "" {
		log.Fatal("Unable to find kubectl")
	}

	a := aws.FindAWS()
	if a == "" {
		log.Fatal("Unable to find aws cli")
	}

	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println("Usage: ekalias <alias name>")
		os.Exit(1)
	}

	awsProfile := aws.SelectProfile()
	fmt.Println("")
	kubeContext := kubectl.SelectContext()
	fmt.Println("")
	fmt.Println(aurora.Green(console.BuildAlias(args[0], awsProfile, kubeContext)))
}
