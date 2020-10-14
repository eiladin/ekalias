package console

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/logrusorgru/aurora/v3"
)

type Executor interface {
	ReadInput() string
	ExecCommand(string, ...string) string
	ExecInteractive(string, ...string) error
	FindExecutable(string) string
}

type DefaultExecutor struct{}

var _ Executor = DefaultExecutor{}

func SelectValueFromList(e Executor, list []string, description string, newFunc func() (string, error)) string {
	var result string
	for result == "" {
		max := 0
		for _, item := range list {
			if item != "" {
				max++
				fmt.Printf("%d. %s\n", max, item)
			}
		}
		if newFunc != nil {
			max++
			fmt.Printf("%d. %s\n", max, "Create New")
		}

		fmt.Printf("\nSelect %s [%d-%d]: ", description, 1, max)
		r := e.ReadInput()
		i, err := strconv.Atoi(r)
		errInvalidInput := aurora.Red(fmt.Sprintf("invalid input -- valid selections: 1-%d\n", max))
		if err != nil {
			fmt.Println(errInvalidInput)
		} else {
			if i > max || i < 1 {
				fmt.Println(errInvalidInput)
			} else {
				if i == max && newFunc != nil {
					var err error
					result, err = newFunc()
					for err != nil {
						result, err = newFunc()
						if err != nil {
							fmt.Println(aurora.Red(err))
						}
					}
				} else {
					result = list[i-1]
				}
			}
		}
	}
	return result
}

func BuildAlias(aliasname, awsProfile, kubeContext string) string {
	return fmt.Sprintf(`alias %s="export AWS_PROFILE=%s && kubectl config use-context %s"`, aliasname, awsProfile, kubeContext)
}

func (e DefaultExecutor) ReadInput() string {
	reader := bufio.NewReader(os.Stdin)
	r, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(r, "\n", "", -1)
}

func (e DefaultExecutor) ExecCommand(name string, arg ...string) string {
	out, err := exec.Command(name, arg...).Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}

func (e DefaultExecutor) ExecInteractive(name string, arg ...string) error {
	cmd := &exec.Cmd{
		Path:   name,
		Args:   append([]string{name}, arg...),
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	return cmd.Run()
}

func (e DefaultExecutor) FindExecutable(name string) string {
	p, err := exec.LookPath(name)
	if err != nil {
		log.Fatal(err)
	}
	return p
}
