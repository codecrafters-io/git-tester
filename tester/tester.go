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

		Logger *log.Logger

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
		cmd.WorkDir = filepath.Join(basedir, cmd.Name())

		err := os.Mkdir(cmd.WorkDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("create temp dir: %w", err)
		}

		cmd.tester = t
	}

	t.want = t.commands[0]

	return t, nil
}

func (t *Tester) Logf(format string, args ...interface{}) {
	if t.Logger == nil {
		return
	}

	t.Logger.Printf(format, args...)
}

func (t *Tester) Run(args ...interface{}) error {
	cmdline, checkers := t.parseArgs(args...)

	return t.Do(func(cmd *Command, i int) error {
		return cmd.Run(cmdline...)
	}, checkers...)
}

func (t *Tester) Do(fn func(*Command, int) error, checkers ...Checker) error {
	err := fn(t.commands[0], 0)
	if err != nil {
		return fmt.Errorf("run first command: %w", err)
	}

	for i, cmd := range t.commands[1:] {
		err = t.runCommand(fn, cmd, 1+i, checkers)
		if err != nil {
			return fmt.Errorf("%v: %w", cmd.Name(), err)
		}
	}

	return nil
}

func (t *Tester) runCommand(fn func(*Command, int) error, cmd *Command, i int, checkers []Checker) (err error) {
	cmd.Reset()

	err = fn(cmd, i)
	if err != nil {
		return err
	}

	if cmd.Logger != nil && len(cmd.Combined()) != 0 {
		for _, line := range bytes.Split(cmd.Combined(), []byte("\n")) {
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

func (t *Tester) RunI(i int, args ...interface{}) (err error) {
	cmdline, checkers := t.parseArgs(args...)

	cmd := t.commands[i]

	err = cmd.Run(cmdline...)
	if err != nil {
		return fmt.Errorf("run command: %w", err)
	}

	for _, c := range checkers {
		err = c.Check(cmd, nil)
		if err != nil {
			return fmt.Errorf("check %T: %w", c, err)
		}
	}

	return nil
}

func (t *Tester) CrossRun(args ...interface{}) (err error) {
	cmdline, checkers := t.parseArgs(args...)

	return t.CrossDo(func(cmd *Command, i int) error {
		return cmd.Run(cmdline...)
	}, checkers...)
}

func (t *Tester) CrossDo(fn func(*Command, int) error, checkers ...Checker) error {
	err := fn(t.commands[0], 0)
	if err != nil {
		return fmt.Errorf("run first command: %w", err)
	}

	for i, cmd := range t.commands[1:] {
		crossCmd := &Command{
			Path:    t.commands[0].Path,
			WorkDir: cmd.WorkDir,
			Logger:  t.commands[0].Logger,
			tester:  t,
		}

		err = t.runCommand(fn, crossCmd, 1+i, checkers)
		if err != nil {
			return fmt.Errorf("%v: %w", cmd.Name(), err)
		}

		fmt.Printf("cross do wd: %v\ncommand: %v\nstdout\n%s\nstderr\n%s\n", crossCmd.WorkDir, crossCmd.Path, crossCmd.Stdout(), crossCmd.Stderr())
	}

	return nil
}

func (t *Tester) Each(f func(*Command) error) error {
	for _, cmd := range t.commands {
		err := f(cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Tester) AllStdouts() []string {
	l := make([]string, len(t.commands))

	for i, cmd := range t.commands {
		l[i] = string(cmd.Stdout())
	}

	return l
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
	c.Reset()

	c.Cmd = exec.Command(c.Path, cmdline...)

	c.Cmd.Dir = c.WorkDir

	c.Cmd.Stdout = io.MultiWriter(&c.stdout, &c.combined)
	c.Cmd.Stderr = io.MultiWriter(&c.stderr, &c.combined)

	//	defer func() {
	//		fmt.Printf("COMMAND %v %v\nSTDOUT\n%s\nSTDERR\n%s\nSTDBOTH\n%s\n", c.Name(), cmdline, c.Stdout(), c.Stderr(), c.Combined())
	//	}()

	err := c.Cmd.Start()
	if err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	timeout := 5 * time.Second

	if c.tester != nil {
		timeout = c.tester.Timeout
	}

	to := time.NewTimer(timeout)
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

func (c *Command) Combined() []byte {
	return bytes.TrimSpace(c.combined.Bytes())
}

func (c *Command) Reset() {
	c.stdout.Reset()
	c.stderr.Reset()
	c.combined.Reset()

	c.Cmd = nil
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

func WithWorkDir(wd string) Option {
	return func(c *Command) error {
		c.WorkDir = wd

		return nil
	}
}
