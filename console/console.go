package console

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/logrusorgru/aurora/v3"
)

func SelectValueFromList(list []string, description string, newFunc func() string) string {
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
		reader := bufio.NewReader(os.Stdin)
		r, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		r = strings.Replace(r, "\n", "", -1)
		i, err := strconv.Atoi(r)
		errInvalidInput := aurora.Red(fmt.Sprintf("invalid input -- valid selections: 1-%d\n", max))
		if err != nil {
			fmt.Println(errInvalidInput)
		} else {
			if i > max || i < 1 {
				fmt.Println(errInvalidInput)
			} else {
				if i == max && newFunc != nil {
					result = newFunc()
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
