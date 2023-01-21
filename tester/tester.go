package tester

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type (
	Tester struct {
		basedir  string
		commands []*Command
		want     *Command

		Logger *log.Logger

		Timeout time.Duration
	}

	Command struct {
		Path    string
		WorkDir string

		*log.Logger

		*exec.Cmd

		stdout, stderr bytes.Buffer
		combined       bytes.Buffer

		tester *Tester
	}

	Checker interface {
		Name() string
		Check(res, wanted *Command) error
	}

	ExitCodeChecker struct{}
	OutputChecker   struct{}

	Option = func(*Command) error
)

var (
	CheckExitCode ExitCodeChecker
	CheckOutput   OutputChecker
)

func New(basedir string, commands ...*Command) (*Tester, error) {
	t := &Tester{
		basedir:  basedir,
		commands: commands,
		Logger:   log.New(os.Stdout, "", 0),
		Timeout:  5 * time.Second,
	}

	for _, cmd := range t.commands {
		wd, err := os.MkdirTemp(basedir, cmd.Name())
		if err != nil {
			return nil, fmt.Errorf("create temp dir: %w", err)
		}

		cmd.tester = t
		cmd.WorkDir = wd
	}

	t.want = t.commands[0]

	return t, nil
}

func (t *Tester) Run(args ...interface{}) (out []byte, err error) {
	cmdline, checkers := t.parseArgs(args...)

	if t.Logger != nil {
		t.Logger.Printf("Running command: %v", strings.Join(cmdline, " "))
	}

	err = t.commands[0].Run(cmdline...)
	if err != nil {
		return nil, fmt.Errorf("run first command: %w", err)
	}

	for _, cmd := range t.commands[1:] {
		err = t.runCommand(cmd, cmdline, checkers)
		if err != nil {
			return nil, fmt.Errorf("%v: %w", cmd.Name(), err)
		}
	}

	out = t.commands[0].stdout.Bytes()
	out = bytes.TrimSpace(out)

	return out, nil
}

func (t *Tester) runCommand(cmd *Command, cmdline []string, checkers []Checker) (err error) {
	err = cmd.Run(cmdline...)
	if err != nil {
		return err
	}

	if cmd.Logger != nil {
		for _, line := range bytes.Split(cmd.CombinedOutput(), []byte("\n")) {
			cmd.Logger.Printf("%s", line)
		}
	}

	for _, c := range checkers {
		err = c.Check(cmd, t.want)
		if err != nil {
			return fmt.Errorf("%v: %w", c.Name(), err)
		}
	}

	return nil
}

func (t *Tester) DoOnlyArgs(i int, args ...interface{}) (out []byte, err error) {
	cmdline, checkers := t.parseArgs(args...)

	cmd := t.commands[i]

	err = cmd.Run(cmdline...)
	if err != nil {
		return nil, fmt.Errorf("run command: %w", err)
	}

	for _, c := range checkers {
		err = c.Check(cmd, nil)
		if err != nil {
			return nil, fmt.Errorf("check %T: %w", c, err)
		}
	}

	return cmd.Stdout(), nil
}

func (t *Tester) Do(f func(*Command) error) error {
	for _, cmd := range t.commands {
		err := f(cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Tester) parseArgs(args ...interface{}) (cmdline []string, cc []Checker) {
	for _, a := range args {
		switch a := a.(type) {
		case string:
			cmdline = append(cmdline, a)
		case Checker:
			cc = append(cc, a)
		default:
			panic(a)
		}
	}

	return
}

func NewCommand(path string, opts ...Option) (*Command, error) {
	cmd := &Command{
		Path: path,
	}

	for _, o := range opts {
		err := o(cmd)
		if err != nil {
			return nil, fmt.Errorf("option: %v", err)
		}
	}

	return cmd, nil
}

func (c *Command) Run(cmdline ...string) error {
	c.stdout.Reset()
	c.stderr.Reset()
	c.combined.Reset()

	c.Cmd = exec.Command(c.Path, cmdline...)

	c.Cmd.Dir = c.WorkDir

	c.Cmd.Stdout = io.MultiWriter(&c.stdout, &c.combined)
	c.Cmd.Stderr = io.MultiWriter(&c.stderr, &c.combined)

	err := c.Cmd.Start()
	if err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	to := time.NewTimer(c.tester.Timeout)
	defer to.Stop()

	errc := make(chan error, 1)

	go func() {
		errc <- c.Cmd.Wait()
	}()

	select {
	case err = <-errc:
	case <-to.C:
		err = errors.New("timeout")
	}

	var exitError *exec.ExitError
	if errors.As(err, &exitError) { // ignore it here
		err = nil
	}

	if err != nil {
		return fmt.Errorf("wait: %w", err)
	}

	return nil
}

func (c *Command) Name() string {
	return filepath.Base(c.Path)
}

func (c *Command) ExitCode() int {
	return c.Cmd.ProcessState.ExitCode()
}

func (c *Command) Stdout() []byte {
	return bytes.TrimSpace(c.stdout.Bytes())
}

func (c *Command) Stderr() []byte {
	return bytes.TrimSpace(c.stderr.Bytes())
}

func (c *Command) CombinedOutput() []byte {
	return bytes.TrimSpace(c.combined.Bytes())
}

func (ExitCodeChecker) Name() string { return "check exit code" }

func (ExitCodeChecker) Check(res, want *Command) error {
	if code, want := res.ProcessState.ExitCode(), want.ProcessState.ExitCode(); code != want {
		return fmt.Errorf("got: %v, wanted: %v", code, want)
	}

	return nil
}

func (OutputChecker) Name() string { return "check program output" }

func (OutputChecker) Check(res, want *Command) error {
	if res, want := res.Stdout(), want.Stdout(); !bytes.Equal(res, want) {
		return fmt.Errorf("got:\n%s\nwanted:\n%s\n", res, want)
	}

	return nil
}

func WithLogger(l *log.Logger) Option {
	return func(c *Command) error {
		c.Logger = l

		return nil
	}
}
