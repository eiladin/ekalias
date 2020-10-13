package kubectl

import (
	"log"
	"os/exec"
	"strings"

	"github.com/eiladin/ekalias/aws"
	"github.com/eiladin/ekalias/console"
)

func findKubectl() string {
	out, err := exec.LookPath("kubectl")
	if err != nil {
		log.Fatal(err)
	}
	return out
}

func findContexts() []string {
	kubectl := findKubectl()
	out := console.ExecCommand(kubectl, "config", "get-contexts", "-o", "name")
	contexts := strings.Split(out, "\n")
	return contexts
}

func SelectContext() string {
	contexts := findContexts()
	return console.SelectValueFromList(contexts, "Kube Context", aws.CreateKubeContext)
}
