package clipboard

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var ErrUnavailable = errors.New("clipboard tool not available")

type command struct {
	name string
	args []string
}

func Copy(text string) error {
	return copyWith(runtime.GOOS, text, os.Getenv, exec.LookPath, runCommand)
}

func copyWith(
	goos string,
	text string,
	getenv func(string) string,
	lookPath func(string) (string, error),
	run func(command, string) error,
) error {
	commands := commandsFor(goos, getenv, lookPath)
	if len(commands) == 0 {
		return ErrUnavailable
	}

	var failures []string
	for _, cmd := range commands {
		if err := run(cmd, text); err == nil {
			return nil
		} else {
			failures = append(failures, fmt.Sprintf("%s: %v", cmd.name, err))
		}
	}

	return fmt.Errorf("clipboard command failed: %s", strings.Join(failures, "; "))
}

func runCommand(cmd command, text string) error {
	execCmd := exec.Command(cmd.name, cmd.args...)
	stdin, err := execCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("open clipboard stdin: %w", err)
	}

	if err := execCmd.Start(); err != nil {
		_ = stdin.Close()
		return fmt.Errorf("start clipboard command: %w", err)
	}

	if _, err := io.WriteString(stdin, text); err != nil {
		_ = stdin.Close()
		_ = execCmd.Wait()
		return fmt.Errorf("write clipboard data: %w", err)
	}

	if err := stdin.Close(); err != nil {
		_ = execCmd.Wait()
		return fmt.Errorf("close clipboard stdin: %w", err)
	}

	if err := execCmd.Wait(); err != nil {
		return err
	}

	return nil
}

func commandsFor(goos string, getenv func(string) string, lookPath func(string) (string, error)) []command {
	var commands []command
	seen := map[string]struct{}{}

	addIfAvailable := func(cmd command) {
		if _, ok := seen[cmd.name]; ok {
			return
		}
		if err := requireCommand(cmd.name, lookPath); err != nil {
			return
		}
		seen[cmd.name] = struct{}{}
		commands = append(commands, cmd)
	}

	switch goos {
	case "darwin":
		addIfAvailable(command{name: "pbcopy"})
	case "windows":
		addIfAvailable(command{name: "clip"})
	default:
		if getenv("WAYLAND_DISPLAY") != "" {
			addIfAvailable(command{name: "wl-copy"})
		}
		if getenv("DISPLAY") != "" {
			addIfAvailable(command{name: "xclip", args: []string{"-selection", "clipboard"}})
			addIfAvailable(command{name: "xsel", args: []string{"--clipboard", "--input"}})
		}
		addIfAvailable(command{name: "wl-copy"})
		addIfAvailable(command{name: "xclip", args: []string{"-selection", "clipboard"}})
		addIfAvailable(command{name: "xsel", args: []string{"--clipboard", "--input"}})
	}

	return commands
}

func requireCommand(name string, lookPath func(string) (string, error)) error {
	if _, err := lookPath(name); err != nil {
		return fmt.Errorf("%w: %s", ErrUnavailable, name)
	}

	return nil
}
