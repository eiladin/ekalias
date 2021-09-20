package kubectl

import (
	"strings"

	"github.com/eiladin/ekalias/aws"
	"github.com/eiladin/ekalias/console"
)

const executable = "kubectl"

type Kubectl struct {
	executor console.Executor
}

func New(e console.Executor) Kubectl {
	return Kubectl{executor: e}
}

func (k Kubectl) FindCli() (string, error) {
	return k.executor.FindExecutable(executable)
}

func (k Kubectl) findContexts() ([]string, error) {
	kubectl, err := k.FindCli()
	if err != nil {
		return []string{}, err
	}
	out, err := k.executor.ExecCommand(kubectl, "config", "get-contexts", "-o", "name")
	if err != nil {
		return []string{}, err
	}
	return strings.Split(out, "\n"), nil
}

func (k Kubectl) SelectContext() (string, error) {
	contexts, err := k.findContexts()
	if err != nil {
		return "", err
	}
	aws := aws.New(k.executor)
	return k.executor.SelectValueFromList(contexts, "Kube Context", aws.CreateKubeContext)
}
