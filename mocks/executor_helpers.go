// +build test

package mocks

import (
	"bytes"
	"io"
	"os"
	"sync"
)

func NewExecutor() *Executor {
	e := new(Executor)
	e.On("FindExecutable", "aws").Return("aws", nil)
	e.On("FindExecutable", "kubectl").Return("kubectl", nil)
	return e
}

func ReadStdOut(f func()) string {
	r, w, _ := os.Pipe()
	stdout := os.Stdout
	stderr := os.Stderr
	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
	}()
	os.Stdout = w
	out := make(chan string)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		_, _ = io.Copy(&buf, r)
		out <- buf.String()
	}()
	wg.Wait()
	f()
	w.Close()
	return <-out
}
