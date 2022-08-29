package common

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	veleroexec "github.com/vmware-tanzu/velero/pkg/util/exec"
)

type OsCommandLine struct {
	Cmd  string
	Args []string
}

func GetListBy2Pipes(ctx context.Context, cmdline1, cmdline2, cmdline3 OsCommandLine) ([]string, error) {
	var b2 bytes.Buffer
	var errVelero, errAwk error

	c1 := exec.CommandContext(ctx, cmdline1.Cmd, cmdline1.Args...)
	c2 := exec.Command(cmdline2.Cmd, cmdline2.Args...)
	c3 := exec.Command(cmdline3.Cmd, cmdline3.Args...)
	fmt.Println(c1)
	fmt.Println(c2)
	fmt.Println(c3)
	c2.Stdin, errVelero = c1.StdoutPipe()
	if errVelero != nil {
		return nil, errVelero
	}
	c3.Stdin, errAwk = c2.StdoutPipe()
	if errAwk != nil {
		return nil, errAwk
	}
	c3.Stdout = &b2
	_ = c3.Start()
	_ = c2.Start()
	_ = c1.Run()
	_ = c2.Wait()
	_ = c3.Wait()

	fmt.Println(&b2)
	scanner := bufio.NewScanner(&b2)
	var ret []string
	for scanner.Scan() {
		fmt.Printf("line: %s\n", scanner.Text())
		ret = append(ret, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func RunShellScript(script, origin_script string) error {
	cmd, err := exec.Command("/bin/sh", script).Output()
	fmt.Println("------5----")
	if err != nil {
		fmt.Println(err)
		return errors.Wrap(err, fmt.Sprintf("Fail to run shell script %s", script))
	}
	output := string(cmd)
	fmt.Printf("Script: %s, output: %s\n", script, output)
	//~/.aws/credentials
	cpCmd := exec.CommandContext(context.TODO(), "cp",
		"aws-credential", origin_script)
	fmt.Printf("cpCmd cmd =%v\n", cpCmd)
	stdout, stderr, err := veleroexec.RunCommand(cpCmd)
	if err != nil {
		fmt.Println(stdout)
		fmt.Println(stderr)
	}

	out, err := CatFile("~/.aws/credentials")
	if err != nil {
		fmt.Println(stderr)
	}
	fmt.Println(out)

	out, err = CatFile("aws_access_key_id")
	if err != nil {
		fmt.Println(stderr)
	}
	fmt.Println(out)
	err = os.Setenv("AWS_ACCESS_KEY_ID", out)
	if err != nil {
		fmt.Println(stderr)
	}

	out, err = CatFile("aws_secret_access_key")
	if err != nil {
		fmt.Println(stderr)
	}
	fmt.Println(out)
	err = os.Setenv("AWS_SECRET_ACCESS_KEY", out)
	if err != nil {
		fmt.Println(stderr)
	}

	out, err = CatFile("aws_session_token")
	if err != nil {
		fmt.Println(stderr)
	}
	fmt.Println(out)
	err = os.Setenv("AWS_SESSION_TOKEN", out)
	if err != nil {
		fmt.Println(stderr)
	}
	return nil
}

func CatFile(file string) (string, error) {
	cpCmd := exec.CommandContext(context.TODO(), "cat", file)
	fmt.Printf("CatFile cmd =%v\n", cpCmd)
	stdout, stderr, err := veleroexec.RunCommand(cpCmd)
	fmt.Print(stdout)
	if err != nil {
		fmt.Print(stderr)
		return "", err
	}
	return stdout, nil
}
