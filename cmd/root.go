package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/eiladin/ekalias/aws"
	"github.com/eiladin/ekalias/console"
	"github.com/eiladin/ekalias/kubectl"
	"github.com/logrusorgru/aurora/v3"
	"github.com/spf13/cobra"
)

func Execute(version string, args []string) {
	newRootCmd(version).Execute(args)
}

type rootCmd struct {
	cmd *cobra.Command
}

func (cmd *rootCmd) Execute(args []string) {
	cmd.cmd.SetArgs(args)

	_ = cmd.cmd.Execute()
}

func newRootCmd(version string) *rootCmd {
	var root = &rootCmd{}
	var cmd = &cobra.Command{
		Use:           "ekalias",
		Short:         "generate shell aliases for switching AWS profiles and kube contexts",
		SilenceUsage:  false,
		SilenceErrors: false,
		Version:       version,
		Args: func(cmd *cobra.Command, args []string) error {
			return validateArgs(args)
		},
		Run: func(cmd *cobra.Command, args []string) {
			k := kubectl.New(nil)
			aws := aws.New(nil)

			_, err := k.FindCli()
			if err != nil {
				log.Fatalf("Unable to find kubectl -> %s", err.Error())
			}

			_, err = aws.FindCli()
			if err != nil {
				log.Fatalf("Unable to find aws cli -> %s", err.Error())
			}

			awsProfile, err := aws.SelectProfile()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("")
			kubeContext, err := k.SelectContext()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("")
			fmt.Println(aurora.Green(console.BuildAlias(args[0], awsProfile, kubeContext)))
		},
	}

	root.cmd = cmd
	return root
}

func validateArgs(args []string) error {
	if len(args) != 1 {
		return errors.New("alias name required")
	}
	return nil
}
