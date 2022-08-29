/*
Copyright 2018 the Velero contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package exec

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

// RunCommand runs a command and returns its stdout, stderr, and its returned
// error (if any). If there are errors reading stdout or stderr, their return
// value(s) will contain the error as a string.
func RunCommand(cmd *exec.Cmd) (string, string, error) {
	out1 := os.Getenv("AWS_ACCESS_KEY_ID")
	out2 := os.Getenv("AWS_SECRET_ACCESS_KEY")
	out3 := os.Getenv("AWS_SESSION_TOKEN")
	fmt.Println("out1")
	fmt.Println(out1)
	fmt.Println("out2")
	fmt.Println(out2)
	fmt.Println("out3")
	fmt.Println(out3)
	cmd.Env = os.Environ()
	fmt.Println("Env1")
	fmt.Println(cmd.Env)
	cmd.Env = append(cmd.Env, "AWS_ACCESS_KEY="+out1)
	cmd.Env = append(cmd.Env, "AWS_SECRET_ACCESS_KEY="+out2)
	cmd.Env = append(cmd.Env, "AWS_SESSION_TOKEN="+out3)
	cmd.Env = os.Environ()
	fmt.Println("Env2")
	fmt.Println(cmd.Env)

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf

	runErr := cmd.Run()

	var stdout, stderr string

	if res, readErr := ioutil.ReadAll(stdoutBuf); readErr != nil {
		stdout = errors.Wrap(readErr, "error reading command's stdout").Error()
	} else {
		stdout = string(res)
	}

	if res, readErr := ioutil.ReadAll(stderrBuf); readErr != nil {
		stderr = errors.Wrap(readErr, "error reading command's stderr").Error()
	} else {
		stderr = string(res)
	}

	return stdout, stderr, runErr
}
