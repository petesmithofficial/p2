package clipboard

import (
	"errors"
	"reflect"
	"testing"
)

func TestCommandsFor(t *testing.T) {
	t.Parallel()

	success := func(commands ...string) func(string) (string, error) {
		available := map[string]struct{}{}
		for _, command := range commands {
			available[command] = struct{}{}
		}

		return func(name string) (string, error) {
			if _, ok := available[name]; ok {
				return name, nil
			}
			return "", errors.New("missing")
		}
	}

	testCases := []struct {
		name     string
		goos     string
		env      map[string]string
		lookPath func(string) (string, error)
		want     []command
	}{
		{
			name:     "darwin",
			goos:     "darwin",
			lookPath: success("pbcopy"),
			want:     []command{{name: "pbcopy"}},
		},
		{
			name:     "windows",
			goos:     "windows",
			lookPath: success("clip"),
			want:     []command{{name: "clip"}},
		},
		{
			name:     "linux wayland prefers wl-copy",
			goos:     "linux",
			env:      map[string]string{"WAYLAND_DISPLAY": "wayland-0"},
			lookPath: success("wl-copy", "xclip"),
			want: []command{
				{name: "wl-copy"},
				{name: "xclip", args: []string{"-selection", "clipboard"}},
			},
		},
		{
			name:     "linux x11 prefers xclip",
			goos:     "linux",
			env:      map[string]string{"DISPLAY": ":0"},
			lookPath: success("wl-copy", "xclip", "xsel"),
			want: []command{
				{name: "xclip", args: []string{"-selection", "clipboard"}},
				{name: "xsel", args: []string{"--clipboard", "--input"}},
				{name: "wl-copy"},
			},
		},
		{
			name:     "linux fallback order without env",
			goos:     "linux",
			lookPath: success("wl-copy", "xclip", "xsel"),
			want: []command{
				{name: "wl-copy"},
				{name: "xclip", args: []string{"-selection", "clipboard"}},
				{name: "xsel", args: []string{"--clipboard", "--input"}},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			getenv := func(key string) string {
				if value, ok := tc.env[key]; ok {
					return value
				}
				return ""
			}

			got := commandsFor(tc.goos, getenv, tc.lookPath)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("commandsFor() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestCopyWithFallsBackToNextWorkingCommand(t *testing.T) {
	t.Parallel()

	lookPath := func(name string) (string, error) {
		switch name {
		case "wl-copy", "xclip":
			return name, nil
		default:
			return "", errors.New("missing")
		}
	}

	var attempted []command
	run := func(cmd command, text string) error {
		attempted = append(attempted, cmd)
		if cmd.name == "wl-copy" {
			return errors.New("exit status 1")
		}
		return nil
	}

	err := copyWith("linux", "32768", func(string) string { return "" }, lookPath, run)
	if err != nil {
		t.Fatalf("copyWith() error = %v, want nil", err)
	}

	want := []command{
		{name: "wl-copy"},
		{name: "xclip", args: []string{"-selection", "clipboard"}},
	}
	if !reflect.DeepEqual(attempted, want) {
		t.Fatalf("copyWith() attempted = %#v, want %#v", attempted, want)
	}
}

func TestCopyWithReturnsUnavailableWhenNoCommandsExist(t *testing.T) {
	t.Parallel()

	err := copyWith(
		"linux",
		"32768",
		func(string) string { return "" },
		func(string) (string, error) { return "", errors.New("missing") },
		func(command, string) error { return nil },
	)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("copyWith() error = %v, want %v", err, ErrUnavailable)
	}
}
