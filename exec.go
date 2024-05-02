// Package exec provides the xk6 Modules implementation for running local commands using Javascript
package exec

import (
	"strings"
	"os"
	"fmt"
	"os/exec"
	"bufio"
	"io"
	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/exec", new(RootModule))
}

// RootModule is the global module object type. It is instantiated once per test
// run and will be used to create `k6/x/exec` module instances for each VU.
type RootModule struct{}

// EXEC represents an instance of the EXEC module for every VU.
type EXEC struct {
	vu modules.VU
}

// CommandOptions contains the options that can be passed to command.
type CommandOptions struct {
	Dir string
}

// Ensure the interfaces are implemented correctly.
var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &EXEC{}
)

// NewModuleInstance implements the modules.Module interface to return
// a new instance for each VU.
func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &EXEC{vu: vu}
}

// Exports implements the modules.Instance interface and returns the exports
// of the JS module.
func (exec *EXEC) Exports() modules.Exports {
	return modules.Exports{Default: exec}
}

// Command is a wrapper for Go exec.Command
func (*EXEC) Command(name string, args []string, option CommandOptions) string {
        var out strings.Builder
	command := exec.Command(name, args...)
	command.Env = os.Environ()
	if option.Dir != "" {
	  command.Dir = option.Dir
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		fmt.Printf("Failed creating command stdoutpipe: %s", err)
		return ""
	}
	defer stdout.Close()
	stdoutReader := bufio.NewReader(stdout)
	stderr, err := command.StderrPipe()
	if err != nil {
		fmt.Printf("Failed creating command stderrpipe: %s", err)
		return ""
	}
	defer stderr.Close()
	stderrReader := bufio.NewReader(stderr)
	if err := command.Start(); err != nil {
		fmt.Printf("Failed starting command: %s", err)
		return ""
	}
	sout := go handleReader(stdoutReader)
	out.WriteString("STDOUT:")
        out.WriteString(sout)
	serr := go handleReader(stderrReader)
	out.WriteString("STDERR:")
        out.WriteString(serr)
	if err := command.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				fmt.Printf("Exit Status: %s", status.ExitStatus())
				fmt.Printf("Err: %s", err)
				return out
			}
		}
		return out
	}
	return out
}
func handleReader(reader *bufio.Reader) error {
        var out strings.Builder
	for {
		str, err := reader.ReadString('\n')
		if len(str) == 0 && err != nil {
			if err == io.EOF {
				break
			}
			return out.String()
		}
		fmt.Print(str)
		out.WriteString(str)
		if err != nil {
			if err == io.EOF {
				break
			}
			return out.String()
		}
	}
	return out.String()
}
