package kubectl

import (
	"strings"

	"github.com/eiladin/ekalias/aws"
	"github.com/eiladin/ekalias/console"
)

var executor console.Executor = console.DefaultExecutor{}

func FindKubectl() string {
	return executor.FindExecutable("kubectl")
}

func findContexts() []string {
	kubectl := FindKubectl()
	out := executor.ExecCommand(kubectl, "config", "get-contexts", "-o", "name")
	return strings.Split(out, "\n")
}

func SelectContext() string {
	contexts := findContexts()
	return console.SelectValueFromList(executor, contexts, "Kube Context", aws.CreateKubeContext)
}
