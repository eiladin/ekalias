package kubectl

import (
	"strings"

	"github.com/eiladin/ekalias/aws"
	"github.com/eiladin/ekalias/console"
)

type Kubectl struct {
	executor console.Executor
}

func Create(e console.Executor) Kubectl {
	k := Kubectl{
		executor: e,
	}
	if k.executor == nil {
		k.executor = console.DefaultExecutor{}
	}
	return k
}

func (k Kubectl) FindCli() string {
	return k.executor.FindExecutable("kubectl")
}

func (k Kubectl) findContexts() []string {
	kubectl := k.FindCli()
	out := k.executor.ExecCommand(kubectl, "config", "get-contexts", "-o", "name")
	return strings.Split(out, "\n")
}

func (k Kubectl) SelectContext() string {
	contexts := k.findContexts()
	aws := aws.Create(k.executor)
	return console.SelectValueFromList(k.executor, contexts, "Kube Context", aws.CreateKubeContext)
}
