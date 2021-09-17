package console

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/logrusorgru/aurora/v3"
)

type Executor interface {
	PromptInput(prompt string) (string, error)
	ReadInput() (string, error)
	ExecCommand(string, ...string) (string, error)
	ExecInteractive(string, ...string) error
	FindExecutable(string) (string, error)
	SelectValueFromList([]string, string, func() (string, error)) (string, error)
}

type DefaultExecutor struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

var _ Executor = DefaultExecutor{}

func New(in io.Reader, out, err io.Writer) Executor {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	if err == nil {
		err = os.Stderr
	}
	return DefaultExecutor{Stdin: in, Stdout: out, Stderr: err}
}

func (e DefaultExecutor) SelectValueFromList(list []string, description string, newFunc func() (string, error)) (string, error) {
	var result string
	for result == "" {
		count := 0
		for _, item := range list {
			if item != "" {
				count++
				fmt.Fprintf(e.Stdout, "%d. %s\n", count, item)
			}
		}
		if newFunc != nil {
			count++
			fmt.Fprintf(e.Stdout, "%d. %s\n", count, "Create New")
		}

		r, err := e.PromptInput(fmt.Sprintf("\nSelect %s [%d-%d]: ", description, 1, count))
		if err != nil {
			return "", err
		}

		i, err := strconv.Atoi(r)
		errInvalidInput := aurora.Red(fmt.Sprintf("invalid input -- valid selections: 1-%d\n", count))
		if err != nil {
			fmt.Fprintln(e.Stdout, errInvalidInput)
		} else {
			if i > count || i < 1 {
				fmt.Fprintln(e.Stdout, errInvalidInput)
			} else {
				if i == count && newFunc != nil {
					var err error
					for result, err = newFunc(); err != nil; result, err = newFunc() {
						fmt.Fprintln(e.Stdout, aurora.Red(err))
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

func (e DefaultExecutor) PromptInput(prompt string) (string, error) {
	fmt.Fprint(e.Stdout, prompt)
	return e.ReadInput()
}

func (e DefaultExecutor) ReadInput() (string, error) {
	reader := bufio.NewReader(e.Stdin)
	r, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.Replace(r, "\n", "", -1), nil
}

func (e DefaultExecutor) ExecCommand(name string, arg ...string) (string, error) {
	cmd := &exec.Cmd{
		Path:   name,
		Args:   append([]string{name}, arg...),
		Stdin:  e.Stdin,
		Stderr: e.Stderr,
	}

	out, err := cmd.Output()
	return string(out), err
}

func (e DefaultExecutor) ExecInteractive(name string, arg ...string) error {
	cmd := &exec.Cmd{
		Path:   name,
		Args:   append([]string{name}, arg...),
		Stdin:  e.Stdin,
		Stdout: e.Stdout,
		Stderr: e.Stderr,
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
