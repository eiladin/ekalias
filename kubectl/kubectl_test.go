// +build test

package kubectl

import (
	"testing"

	"github.com/eiladin/ekalias/mocks"
	"github.com/stretchr/testify/suite"
)

type KubectlSuite struct {
	suite.Suite
}

func TestAWSSuite(t *testing.T) {
	suite.Run(t, new(KubectlSuite))
}

func (suite KubectlSuite) TestFindContexts() {
	e := mocks.NewExecutor()
	e.On("ExecCommand", "kubectl", "config", "get-contexts", "-o", "name").Return("a\nb\nc")
	executor = e
	mocks.ReadStdOut(func() {
		res := findContexts()
		suite.Len(res, 3)
	})
}

func (suite KubectlSuite) TestSelectContext() {
	e := mocks.NewExecutor()
	e.On("ExecCommand", "kubectl", "config", "get-contexts", "-o", "name").Return("a\nb\nc")
	e.On("ReadInput").Return("2")
	executor = e

	mocks.ReadStdOut(func() {
		res := SelectContext()
		suite.Equal("b", res)
	})
}

func (suite KubectlSuite) TestFindKubectl() {
	executor = mocks.NewExecutor()

	res := FindKubectl()
	suite.Equal("kubectl", res)
}
