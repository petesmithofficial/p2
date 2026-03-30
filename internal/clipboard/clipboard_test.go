package clipboard

import (
	"errors"
	"reflect"
	"testing"
)

func TestCommandFor(t *testing.T) {
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

	env := func(vars map[string]string) func(string) string {
		return func(key string) string {
			return vars[key]
		}
	}

	noEnv := env(nil)
	waylandEnv := env(map[string]string{"WAYLAND_DISPLAY": "wayland-0"})

	testCases := []struct {
		name     string
		goos     string
		lookPath func(string) (string, error)
		getenv   func(string) string
		want     command
		wantErr  error
	}{
		{
			name:     "darwin",
			goos:     "darwin",
			lookPath: success("pbcopy"),
			getenv:   noEnv,
			want:     command{name: "pbcopy"},
		},
		{
			name:     "windows",
			goos:     "windows",
			lookPath: success("clip"),
			getenv:   noEnv,
			want:     command{name: "clip"},
		},
		{
			name:     "linux wl-copy on wayland",
			goos:     "linux",
			lookPath: success("wl-copy"),
			getenv:   waylandEnv,
			want:     command{name: "wl-copy"},
		},
		{
			name:     "linux wl-copy installed but X11 falls through to xclip",
			goos:     "linux",
			lookPath: success("wl-copy", "xclip"),
			getenv:   noEnv,
			want:     command{name: "xclip", args: []string{"-selection", "clipboard"}},
		},
		{
			name:     "linux wl-copy installed but X11 falls through to xsel",
			goos:     "linux",
			lookPath: success("wl-copy", "xsel"),
			getenv:   noEnv,
			want:     command{name: "xsel", args: []string{"--clipboard", "--input"}},
		},
		{
			name:     "linux xclip fallback",
			goos:     "linux",
			lookPath: success("xclip"),
			getenv:   noEnv,
			want:     command{name: "xclip", args: []string{"-selection", "clipboard"}},
		},
		{
			name:     "linux xsel fallback",
			goos:     "linux",
			lookPath: success("xsel"),
			getenv:   noEnv,
			want:     command{name: "xsel", args: []string{"--clipboard", "--input"}},
		},
		{
			name:     "unavailable",
			goos:     "linux",
			lookPath: success(),
			getenv:   noEnv,
			wantErr:  ErrUnavailable,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := commandFor(tc.goos, tc.lookPath, tc.getenv)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("commandFor() error = %v, want %v", err, tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("commandFor() error = %v, want nil", err)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("commandFor() = %#v, want %#v", got, tc.want)
			}
		})
	}
}
