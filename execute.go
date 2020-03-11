package gexec

import (
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
)

type Commander interface {
	Exec(a ...string) (int, string, error)

	ExecAsync(stdout chan string, a ...string) int

	ExecSync(a ...string) error

	ExecBG(out string, a ...string) (int, error)

	ExecNoRes(a ...string) (int, error)
}

type Command struct {}

func (c *Command) Exec(a ...string) (int, string, error) {
	cmd := newCmd(a...)

	cmd.SysProcAttr = &syscall.SysProcAttr{}

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return 0, "", err
	}

	err = cmd.Start()
	if err != nil {
		return 0, "", err
	}

	out, err := ioutil.ReadAll(outPipe)
	if err != nil {
		return 0, "", err
	}

	return cmd.Process.Pid, string(out), nil
}

func (c *Command) ExecAsync(stdout chan string, a ...string) int {
	var pid = make(chan int, 1)

	go func() {
		cmd := newCmd(a...)

		cmd.SysProcAttr = &syscall.SysProcAttr{}

		outPipe, err := cmd.StdoutPipe()
		if err != nil {
			panic(err)
		}

		err = cmd.Start()
		if err != nil {
			panic(err)
		}

		pid <-cmd.Process.Pid

		out, err := ioutil.ReadAll(outPipe)
		if err != nil {
			panic(err)
		}

		stdout <-string(out)
	}()

	return <-pid
}

func (c *Command) ExecSync(a ...string) error {
	cmd := newCmd(a...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{}

	err := cmd.Run()

	return err
}

func (c *Command) ExecBG(out string, a ...string) (int, error) {
	cmd := newCmd(a...)

	f ,err := os.Create(out)
	if err != nil {
		return 0, err
	}

	cmd.Stdout = f
	cmd.Stderr = f
	cmd.SysProcAttr = &syscall.SysProcAttr{}

	err = cmd.Start()
	if err != nil {
		return 0, err
	}

	return cmd.Process.Pid, err
}

func (c *Command) ExecNoRes(a ...string) (int, error) {
	cmd := newCmd(a...)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	err := cmd.Start()

	return cmd.Process.Pid, err
}

func NewCommand() *Command {
	return &Command{}
}

func newCmd(a ...string) *exec.Cmd {
	first := a[0]
	args := a[1:]
	cmd := exec.Command(first, args...)

	return cmd
}

