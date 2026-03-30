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

func TestRunClosestArgAcceptsCommaSeparatedInput(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	var copied string
	deps := testDeps(config.Default())
	deps.copy = func(value string) error {
		copied = value
		return nil
	}

	exitCode := runWithDeps([]string{"30,000"}, bytes.NewBuffer(nil), &stdout, &stderr, deps)
	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0", exitCode)
	}

	if stdout.String() != "2^15 = 32,768\n" {
		t.Fatalf("stdout = %q, want %q", stdout.String(), "2^15 = 32,768\n")
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	if copied != "32768" {
		t.Fatalf("copied value = %q, want %q", copied, "32768")
	}
}

func TestRunRangeArg(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	copied := false
	deps := testDeps(config.Default())
	deps.copy = func(string) error {
		copied = true
		return nil
	}

	exitCode := runWithDeps([]string{"0..4"}, bytes.NewBuffer(nil), &stdout, &stderr, deps)
	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0", exitCode)
	}

	want := "2^0 = 1\n2^1 = 2\n2^2 = 4\n2^3 = 8\n2^4 = 16\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	if copied {
		t.Fatal("copy was called for multi-result range, want no copy")
	}
}

func TestRunRangeAliasArg(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := runWithDeps([]string{"1-3"}, bytes.NewBuffer(nil), &stdout, &stderr, testDeps(config.Default()))
	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0", exitCode)
	}

	want := "2^1 = 2\n2^2 = 4\n2^3 = 8\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunDescendingRangeArg(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := runWithDeps([]string{"16..5"}, bytes.NewBuffer(nil), &stdout, &stderr, testDeps(config.Default()))
	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0", exitCode)
	}

	want := "2^5  = 32\n2^6  = 64\n2^7  = 128\n2^8  = 256\n2^9  = 512\n2^10 = 1,024\n2^11 = 2,048\n2^12 = 4,096\n2^13 = 8,192\n2^14 = 16,384\n2^15 = 32,768\n2^16 = 65,536\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunSingleEntryRangeArgCopies(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	var copied string
	deps := testDeps(config.Default())
	deps.copy = func(value string) error {
		copied = value
		return nil
	}

	exitCode := runWithDeps([]string{"5..5"}, bytes.NewBuffer(nil), &stdout, &stderr, deps)
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

func TestRunRangeIgnoresConfiguredBounds(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cfg := config.Default()
	cfg.LowerBound = 10
	cfg.UpperBound = 12

	exitCode := runWithDeps([]string{"0..2"}, bytes.NewBuffer(nil), &stdout, &stderr, testDeps(cfg))
	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0", exitCode)
	}

	want := "2^0 = 1\n2^1 = 2\n2^2 = 4\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
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

	if !bytes.Contains(stdout.Bytes(), []byte("usage: p2 [integer|range]")) {
		t.Fatalf("stdout = %q, want help text", stdout.String())
	}

	if !bytes.Contains(stdout.Bytes(), []byte("/tmp/p2-test-config.json")) {
		t.Fatalf("stdout = %q, want config path", stdout.String())
	}

	if !bytes.Contains(stdout.Bytes(), []byte("--config")) {
		t.Fatalf("stdout = %q, want config option in help text", stdout.String())
	}

	if !bytes.Contains(stdout.Bytes(), []byte("A..B   print exponents from A through B")) {
		t.Fatalf("stdout = %q, want range help text", stdout.String())
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

func TestParseIntegerArg(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  string
		ok    bool
	}{
		{name: "plain", input: "30000", want: "30000", ok: true},
		{name: "comma separated", input: "30,000", want: "30000", ok: true},
		{name: "signed comma separated", input: "-30,000", want: "-30000", ok: true},
		{name: "single group", input: "999", want: "999", ok: true},
		{name: "bad grouping short", input: "3,00", ok: false},
		{name: "bad grouping double comma", input: "30,,000", ok: false},
		{name: "bad grouping leading comma", input: ",30000", ok: false},
		{name: "bad grouping trailing comma", input: "30,000,", ok: false},
		{name: "bad characters", input: "30,00a", ok: false},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseIntegerArg(tc.input)
			if ok != tc.ok {
				t.Fatalf("parseIntegerArg(%q) ok = %v, want %v", tc.input, ok, tc.ok)
			}

			if !tc.ok {
				return
			}

			if got.String() != tc.want {
				t.Fatalf("parseIntegerArg(%q) = %q, want %q", tc.input, got.String(), tc.want)
			}
		})
	}
}

func TestParseRangeArg(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     string
		want      string
		isRange   bool
		wantError bool
	}{
		{name: "dot range", input: "0..2", want: "2^0 = 1\n2^1 = 2\n2^2 = 4", isRange: true},
		{name: "hyphen range", input: "1-3", want: "2^1 = 2\n2^2 = 4\n2^3 = 8", isRange: true},
		{name: "descending range", input: "16..14", want: "2^14 = 16,384\n2^15 = 32,768\n2^16 = 65,536", isRange: true},
		{name: "plain integer", input: "30000", isRange: false},
		{name: "comma integer", input: "30,000", isRange: false},
		{name: "missing start", input: "..16", isRange: true, wantError: true},
		{name: "missing end", input: "5-", isRange: true, wantError: true},
		{name: "mixed separators", input: "1..2-3", isRange: true, wantError: true},
		{name: "too large", input: "33..40", isRange: true, wantError: true},
		{name: "target style", input: "30000..40000", isRange: true, wantError: true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, isRange, err := parseRangeArg(tc.input)
			if isRange != tc.isRange {
				t.Fatalf("parseRangeArg(%q) isRange = %v, want %v", tc.input, isRange, tc.isRange)
			}

			if tc.wantError {
				if err == nil {
					t.Fatalf("parseRangeArg(%q) error = nil, want error", tc.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseRangeArg(%q) error = %v, want nil", tc.input, err)
			}

			if !tc.isRange {
				return
			}

			if formatted := powers.FormatEntries(got, true); formatted != tc.want {
				t.Fatalf("parseRangeArg(%q) = %q, want %q", tc.input, formatted, tc.want)
			}
		})
	}
}

func TestRunInvalidInput(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		args []string
	}{
		{name: "negative", args: []string{"-1"}},
		{name: "negative with commas", args: []string{"-30,000"}},
		{name: "decimal", args: []string{"5.5"}},
		{name: "nonnumeric", args: []string{"hello"}},
		{name: "unknown flag", args: []string{"--nope"}},
		{name: "bad comma grouping", args: []string{"3,00"}},
		{name: "bad range missing start", args: []string{"..16"}},
		{name: "bad range missing end", args: []string{"5-"}},
		{name: "bad range mixed separators", args: []string{"1..2-3"}},
		{name: "bad range too large", args: []string{"33..40"}},
		{name: "bad range target style", args: []string{"30000..40000"}},
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
