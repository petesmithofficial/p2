package clipboard

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
)

var ErrUnavailable = errors.New("clipboard tool not available")

type command struct {
	name string
	args []string
}

func Copy(text string) error {
	cmd, err := commandFor(runtime.GOOS, exec.LookPath, os.Getenv)
	if err != nil {
		return err
	}

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
		return fmt.Errorf("clipboard command failed: %w", err)
	}

	return nil
}

func commandFor(goos string, lookPath func(string) (string, error), getenv func(string) string) (command, error) {
	switch goos {
	case "darwin":
		if err := requireCommand("pbcopy", lookPath); err != nil {
			return command{}, err
		}
		return command{name: "pbcopy"}, nil
	case "windows":
		if err := requireCommand("clip", lookPath); err != nil {
			return command{}, err
		}
		return command{name: "clip"}, nil
	default:
		if getenv("WAYLAND_DISPLAY") != "" {
			if err := requireCommand("wl-copy", lookPath); err == nil {
				return command{name: "wl-copy"}, nil
			}
		}
		if err := requireCommand("xclip", lookPath); err == nil {
			return command{name: "xclip", args: []string{"-selection", "clipboard"}}, nil
		}
		if err := requireCommand("xsel", lookPath); err == nil {
			return command{name: "xsel", args: []string{"--clipboard", "--input"}}, nil
		}
		return command{}, ErrUnavailable
	}
}

func requireCommand(name string, lookPath func(string) (string, error)) error {
	if _, err := lookPath(name); err != nil {
		return fmt.Errorf("%w: %s", ErrUnavailable, name)
	}

	return nil
}
