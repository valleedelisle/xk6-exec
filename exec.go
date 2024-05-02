// Package exec provides the xk6 Modules implementation for running local commands using Javascript
package exec

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
        "bytes"
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
	cmd := exec.Command(name, args...)
        fmt.Printf("Command: %s %s", name, strings.Join(args, " "))
	if option.Dir != "" {
	  cmd.Dir = option.Dir
	}
        pipe, _ := cmd.StdoutPipe()
        if err := cmd.Start(); err != nil {
          fmt.Printf("Start Error: %s", err.Error())
        }
        go func(p io.ReadCloser) {
            reader := bufio.NewReader(pipe)
            line, err := reader.ReadString('\n')
            for err == nil {
		out.WriteString(line)
                fmt.Println(line)
                line, err = reader.ReadString('\n')
            }
        }(pipe)
        if err := cmd.Wait(); err != nil {
          fmt.Printf("Wait Error: %s", err.Error())
        }
	return string(out)
}
