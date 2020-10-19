package console

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/logrusorgru/aurora/v3"
)

type Executor interface {
	ReadInput() (string, error)
	ExecCommand(string, ...string) (string, error)
	ExecInteractive(string, ...string) error
	FindExecutable(string) (string, error)
}

type DefaultExecutor struct{}

var _ Executor = DefaultExecutor{}

func SelectValueFromList(e Executor, list []string, description string, newFunc func() (string, error)) (string, error) {
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
		r, err := e.ReadInput()
		if err != nil {
			return "", err
		}
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
	return result, nil
}

func BuildAlias(aliasname, awsProfile, kubeContext string) string {
	return fmt.Sprintf(`alias %s="export AWS_PROFILE=%s && kubectl config use-context %s"`, aliasname, awsProfile, kubeContext)
}

func (e DefaultExecutor) ReadInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	r, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.Replace(r, "\n", "", -1), nil
}

func (e DefaultExecutor) ExecCommand(name string, arg ...string) (string, error) {
	out, err := exec.Command(name, arg...).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
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

func (e DefaultExecutor) FindExecutable(name string) (string, error) {
	p, err := exec.LookPath(name)
	if err != nil {
		return "", err
	}
	return p, nil
}
