package main

import (
	"bytes"
	"errors"
	"testing"

	"p2/internal/config"
	"p2/internal/powers"
)

func TestRunNoArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := runWithDeps(nil, bytes.NewBuffer(nil), &stdout, &stderr, testDeps(config.Default()))

	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0", exitCode)
	}

	want := powers.FormatEntries(powers.Between(0, 16), true) + "\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunExponentArg(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	var copied string
	deps := testDeps(config.Default())
	deps.copy = func(value string) error {
		copied = value
		return nil
	}

	exitCode := runWithDeps([]string{"5"}, bytes.NewBuffer(nil), &stdout, &stderr, deps)

	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0", exitCode)
	}

	if stdout.String() != "2^5 = 32\n" {
		t.Fatalf("stdout = %q, want %q", stdout.String(), "2^5 = 32\n")
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	if copied != "32" {
		t.Fatalf("copied value = %q, want %q", copied, "32")
	}
}

func TestRunClosestArg(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cfg := config.Default()
	cfg.UseCommas = false

	var copied string
	deps := testDeps(cfg)
	deps.copy = func(value string) error {
		copied = value
		return nil
	}

	exitCode := runWithDeps([]string{"30000"}, bytes.NewBuffer(nil), &stdout, &stderr, deps)

	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0", exitCode)
	}

	if stdout.String() != "2^15 = 32768\n" {
		t.Fatalf("stdout = %q, want %q", stdout.String(), "2^15 = 32768\n")
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	if copied != "32768" {
		t.Fatalf("copied value = %q, want %q", copied, "32768")
	}
}

func TestRunTieArg(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	copied := false
	deps := testDeps(config.Default())
	deps.copy = func(string) error {
		copied = true
		return nil
	}

	exitCode := runWithDeps([]string{"48"}, bytes.NewBuffer(nil), &stdout, &stderr, deps)

	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0", exitCode)
	}

	if stdout.String() != "2^5 = 32\n2^6 = 64\n" {
		t.Fatalf("stdout = %q, want %q", stdout.String(), "2^5 = 32\n2^6 = 64\n")
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	if copied {
		t.Fatal("copy was called for tie result, want no copy")
	}
}

func TestRunNoArgsRespectsBounds(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cfg := config.Default()
	cfg.LowerBound = 5
	cfg.UpperBound = 8

	exitCode := runWithDeps(nil, bytes.NewBuffer(nil), &stdout, &stderr, testDeps(cfg))
	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0", exitCode)
	}

	if stdout.String() != "2^5 = 32\n2^6 = 64\n2^7 = 128\n2^8 = 256\n" {
		t.Fatalf("stdout = %q, want %q", stdout.String(), "2^5 = 32\n2^6 = 64\n2^7 = 128\n2^8 = 256\n")
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := runWithDeps([]string{"--help"}, bytes.NewBuffer(nil), &stdout, &stderr, testDeps(config.Default()))
	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0", exitCode)
	}

	if !bytes.Contains(stdout.Bytes(), []byte("usage: p2 [integer]")) {
		t.Fatalf("stdout = %q, want help text", stdout.String())
	}

	if !bytes.Contains(stdout.Bytes(), []byte("/tmp/p2-test-config.json")) {
		t.Fatalf("stdout = %q, want config path", stdout.String())
	}

	if !bytes.Contains(stdout.Bytes(), []byte("--config")) {
		t.Fatalf("stdout = %q, want config option in help text", stdout.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunConfigError(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	deps := testDeps(config.Default())
	deps.loadConfig = func() (config.Config, string, error) {
		return config.Config{}, "/tmp/p2-test-config.json", errors.New("bad config")
	}

	exitCode := runWithDeps(nil, bytes.NewBuffer(nil), &stdout, &stderr, deps)
	if exitCode != 1 {
		t.Fatalf("run() exit code = %d, want 1", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}

	if !bytes.Contains(stderr.Bytes(), []byte("config error: bad config")) {
		t.Fatalf("stderr = %q, want config error", stderr.String())
	}
}

func TestRunClipboardWarning(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	deps := testDeps(config.Default())
	deps.copy = func(string) error {
		return errors.New("clipboard unavailable")
	}

	exitCode := runWithDeps([]string{"5"}, bytes.NewBuffer(nil), &stdout, &stderr, deps)
	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0", exitCode)
	}

	if !bytes.Contains(stderr.Bytes(), []byte("warning: failed to copy to clipboard: clipboard unavailable")) {
		t.Fatalf("stderr = %q, want warning", stderr.String())
	}
}

func TestRunConfigSetup(t *testing.T) {
	var stdin bytes.Buffer
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	stdin.WriteString("5\n8\nn\ny\n")

	var saved config.Config
	deps := testDeps(config.Default())
	deps.saveConfig = func(cfg config.Config) (string, error) {
		saved = cfg
		return "/tmp/p2-test-config.json", nil
	}

	exitCode := runWithDeps([]string{"--config"}, &stdin, &stdout, &stderr, deps)
	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0", exitCode)
	}

	want := config.Default()
	want.LowerBound = 5
	want.UpperBound = 8
	want.UseCommas = false
	want.CopySingleToClipboard = true

	if saved != want {
		t.Fatalf("saved config = %#v, want %#v", saved, want)
	}

	if !bytes.Contains(stdout.Bytes(), []byte("saved config to /tmp/p2-test-config.json")) {
		t.Fatalf("stdout = %q, want save message", stdout.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunConfigSetupFallsBackToDefaultsOnConfigError(t *testing.T) {
	var stdin bytes.Buffer
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	stdin.WriteString("\n\n\n\n")

	var saved config.Config
	deps := testDeps(config.Default())
	deps.loadConfig = func() (config.Config, string, error) {
		return config.Config{}, "/tmp/p2-test-config.json", errors.New("bad config")
	}
	deps.saveConfig = func(cfg config.Config) (string, error) {
		saved = cfg
		return "/tmp/p2-test-config.json", nil
	}

	exitCode := runWithDeps([]string{"--config"}, &stdin, &stdout, &stderr, deps)
	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0", exitCode)
	}

	if saved != config.Default() {
		t.Fatalf("saved config = %#v, want %#v", saved, config.Default())
	}

	if !bytes.Contains(stderr.Bytes(), []byte("warning: config is invalid")) {
		t.Fatalf("stderr = %q, want warning", stderr.String())
	}
}

func TestRunReset(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	var saved config.Config
	deps := testDeps(config.Default())
	deps.saveConfig = func(cfg config.Config) (string, error) {
		saved = cfg
		return "/tmp/p2-test-config.json", nil
	}

	exitCode := runWithDeps([]string{"--reset"}, bytes.NewBuffer(nil), &stdout, &stderr, deps)
	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0", exitCode)
	}

	if saved != config.Default() {
		t.Fatalf("saved config = %#v, want %#v", saved, config.Default())
	}

	if !bytes.Contains(stdout.Bytes(), []byte("reset config to defaults at /tmp/p2-test-config.json")) {
		t.Fatalf("stdout = %q, want reset message", stdout.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestDefaultConfigUpperBound(t *testing.T) {
	t.Parallel()

	if got := config.Default().UpperBound; got != 16 {
		t.Fatalf("config.Default().UpperBound = %d, want 16", got)
	}
}

func TestRunInvalidInput(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		args []string
	}{
		{name: "negative", args: []string{"-1"}},
		{name: "decimal", args: []string{"5.5"}},
		{name: "nonnumeric", args: []string{"hello"}},
		{name: "unknown flag", args: []string{"--nope"}},
		{name: "extra args", args: []string{"1", "2"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			exitCode := runWithDeps(tc.args, bytes.NewBuffer(nil), &stdout, &stderr, testDeps(config.Default()))

			if exitCode != 2 {
				t.Fatalf("run(%v) exit code = %d, want 2", tc.args, exitCode)
			}

			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}

			if stderr.Len() == 0 {
				t.Fatal("stderr is empty, want usage text")
			}
		})
	}
}

func testDeps(cfg config.Config) appDeps {
	return appDeps{
		loadConfig: func() (config.Config, string, error) {
			return cfg, "/tmp/p2-test-config.json", nil
		},
		saveConfig: func(cfg config.Config) (string, error) {
			return "/tmp/p2-test-config.json", nil
		},
		copy: func(string) error {
			return nil
		},
		configPath: func() string {
			return "/tmp/p2-test-config.json"
		},
	}
}
