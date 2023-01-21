package tester

import (
	"bytes"
	"fmt"
	"os"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

type (
	Tester struct {
		basedir  string
		commands []*tester_utils.Executable

		Want tester_utils.ExecutableResult
	}
)

func New(basedir string, commands ...*tester_utils.Executable) (*Tester, error) {
	t := &Tester{
		basedir:  basedir,
		commands: commands,
	}

	for _, cmd := range commands {
		var err error

		cmd.WorkingDir, err = os.MkdirTemp(basedir, "")
		if err != nil {
			return nil, fmt.Errorf("create temp dir: %w", err)
		}
	}

	return t, nil
}

func (t *Tester) Run(args ...interface{}) (out []byte, err error) {
	cmdline := t.parseArgs(args...)

	t.Want, err = t.commands[0].Run(cmdline...)
	if err != nil {
		return nil, fmt.Errorf("run first command: %w", err)
	}

	t.Want.Stdout = bytes.TrimSpace(t.Want.Stdout)
	t.Want.Stderr = bytes.TrimSpace(t.Want.Stderr)

	for _, cmd := range t.commands[1:] {
		err = t.runCommand(cmd, cmdline)
		if err != nil {
			return
		}
	}

	return t.Want.Stdout, nil
}

func (t *Tester) runCommand(cmd *tester_utils.Executable, cmdline []string) (err error) {
	res, err := cmd.Run(cmdline...)
	if err != nil {
		return fmt.Errorf("run command: %w", err)
	}

	if res.ExitCode != t.Want.ExitCode {
		return fmt.Errorf("exit code: %v, wanted: %v", res.ExitCode, t.Want.ExitCode)
	}

	res.Stdout = bytes.TrimSpace(res.Stdout)
	res.Stderr = bytes.TrimSpace(res.Stderr)

	if bytes.Equal(res.Stdout, t.Want.Stdout) {
		return fmt.Errorf("output:\n%v\nwanted:\n%v\n", res.Stdout, t.Want.Stderr)
	}

	return nil
}

func (t *Tester) DoOnlyArgs(i int, args ...interface{}) (out []byte, err error) {
	cmdline := t.parseArgs(args...)

	res, err := t.commands[i].Run(cmdline...)
	if err != nil {
		return nil, fmt.Errorf("run command: %w", err)
	}

	res.Stdout = bytes.TrimSpace(res.Stdout)
	res.Stderr = bytes.TrimSpace(res.Stderr)

	return res.Stdout, nil
}

func (t *Tester) Do(f func(*tester_utils.Executable) error) error {
	for _, cmd := range t.commands {
		err := f(cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Tester) parseArgs(args ...interface{}) []string {
	var cmdline []string

	for _, a := range args {
		switch a := a.(type) {
		case string:
			cmdline = append(cmdline, a)
		default:
			panic(a)
		}
	}

	return cmdline
}
