package main

import (
	"fmt"

	"github.com/eiladin/ekalias/aws"
	"github.com/eiladin/ekalias/console"
	"github.com/eiladin/ekalias/kubectl"
	"github.com/logrusorgru/aurora/v3"
)

func main() {
	awsProfile := aws.SelectProfile()
	fmt.Println("")
	kubeContext := kubectl.SelectContext()
	fmt.Println("")
	fmt.Println(aurora.Green(console.BuildAlias("test", awsProfile, kubeContext)))
}
