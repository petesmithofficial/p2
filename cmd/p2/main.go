package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"

	"p2/internal/clipboard"
	"p2/internal/config"
	"p2/internal/powers"
)

type appDeps struct {
	loadConfig func() (config.Config, string, error)
	saveConfig func(config.Config) (string, error)
	copy       func(string) error
	configPath func() string
}

var defaultDeps = appDeps{
	loadConfig: config.Load,
	saveConfig: config.Save,
	copy:       clipboard.Copy,
	configPath: config.DisplayPath,
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	return runWithDeps(args, stdin, stdout, stderr, defaultDeps)
}

func runWithDeps(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, deps appDeps) int {
	switch len(args) {
	case 0:
	case 1:
	default:
		return usageError(stderr, "expected zero or one argument", deps.configPath())
	}

	if len(args) == 1 {
		switch args[0] {
		case "-h", "--help":
			_, _ = fmt.Fprint(stdout, helpText(deps.configPath()))
			return 0
		case "--config":
			return runConfigSetup(stdin, stdout, stderr, deps)
		case "--reset":
			return runConfigReset(stdout, stderr, deps)
		}
	}

	cfg, _, err := deps.loadConfig()
	if err != nil {
		return configError(stderr, err)
	}

	if len(args) == 0 {
		_, _ = fmt.Fprintln(stdout, powers.FormatEntries(powers.Between(cfg.LowerBound, cfg.UpperBound), cfg.UseCommas))
		return 0
	}

	if entries, isRange, err := parseRangeArg(args[0]); isRange {
		if err != nil {
			return usageError(stderr, err.Error(), deps.configPath())
		}

		_, _ = fmt.Fprintln(stdout, powers.FormatEntries(entries, cfg.UseCommas))
		if cfg.CopySingleToClipboard && len(entries) == 1 {
			if err := deps.copy(powers.RawUint(entries[0].Value)); err != nil {
				_, _ = fmt.Fprintf(stderr, "warning: failed to copy to clipboard: %v\n", err)
			}
		}
		return 0
	}

	input, ok := parseIntegerArg(args[0])
	if !ok {
		if looksLikeFlag(args[0]) {
			return usageError(stderr, fmt.Sprintf("unknown flag %q", args[0]), deps.configPath())
		}

		return usageError(stderr, fmt.Sprintf("invalid integer %q", args[0]), deps.configPath())
	}

	if input.Sign() < 0 {
		return usageError(stderr, "negative integers are not supported", deps.configPath())
	}

	var entries []powers.Entry
	if input.Cmp(big.NewInt(powers.MaxExponent)) <= 0 {
		entry, found := powers.ByExponent(uint(input.Uint64()))
		if !found {
			return usageError(stderr, fmt.Sprintf("unsupported exponent %q", args[0]), deps.configPath())
		}

		entries = []powers.Entry{entry}
	} else {
		entries = powers.ClosestTo(input)
	}

	_, _ = fmt.Fprintln(stdout, powers.FormatEntries(entries, cfg.UseCommas))
	if cfg.CopySingleToClipboard && len(entries) == 1 {
		if err := deps.copy(powers.RawUint(entries[0].Value)); err != nil {
			_, _ = fmt.Fprintf(stderr, "warning: failed to copy to clipboard: %v\n", err)
		}
	}

	return 0
}

func runConfigSetup(stdin io.Reader, stdout io.Writer, stderr io.Writer, deps appDeps) int {
	cfg, path, err := deps.loadConfig()
	if err != nil {
		if path == "" {
			return configError(stderr, err)
		}

		cfg = config.Default()
		_, _ = fmt.Fprintf(stderr, "warning: config is invalid, using defaults for setup: %v\n", err)
	}

	_, _ = fmt.Fprintf(stdout, "Config setup for %s\n", path)
	_, _ = fmt.Fprintln(stdout, "Press Enter to keep the current value.")

	updated, err := promptConfig(stdin, stdout, cfg)
	if err != nil {
		return configError(stderr, err)
	}

	savedPath, err := deps.saveConfig(updated)
	if err != nil {
		return configError(stderr, err)
	}

	_, _ = fmt.Fprintf(stdout, "saved config to %s\n", savedPath)
	return 0
}

func runConfigReset(stdout io.Writer, stderr io.Writer, deps appDeps) int {
	savedPath, err := deps.saveConfig(config.Default())
	if err != nil {
		return configError(stderr, err)
	}

	_, _ = fmt.Fprintf(stdout, "reset config to defaults at %s\n", savedPath)
	return 0
}

func usageError(stderr io.Writer, message, configPath string) int {
	_, _ = fmt.Fprintf(stderr, "error: %s\n\n%s", message, helpText(configPath))
	return 2
}

func configError(stderr io.Writer, err error) int {
	_, _ = fmt.Fprintf(stderr, "config error: %v\n", err)
	return 1
}

func helpText(configPath string) string {
	return fmt.Sprintf(`usage: p2 [integer|range]

Print powers of 2 from 2^0 through 2^32.

With one integer argument:
  0..32  treat the value as an exponent
  >32    treat the value as a target and print the closest power of 2

With one range argument:
  A..B   print exponents from A through B
  A-B    same as A..B
  16..5  normalizes to 5 through 16

Options:
  -h, --help  show this help text
  --config    create or edit config with an interactive setup
  --reset     write a fresh default config

Config:
  path: %s
  run p2 --config to create or edit the file
  lower_bound and upper_bound apply to bare p2 only
  use_commas defaults to true
  copy_single_to_clipboard defaults to true

Example config:
{
  "lower_bound": 0,
  "upper_bound": 16,
  "use_commas": true,
  "copy_single_to_clipboard": true
}
`, configPath)
}

func promptConfig(stdin io.Reader, stdout io.Writer, current config.Config) (config.Config, error) {
	reader := bufio.NewReader(stdin)

	lower, err := promptInt(reader, stdout, "Lower bound (inclusive)", current.LowerBound, 0, int(powers.MaxExponent))
	if err != nil {
		return config.Config{}, err
	}

	upper, err := promptInt(reader, stdout, "Upper bound (inclusive)", current.UpperBound, lower, int(powers.MaxExponent))
	if err != nil {
		return config.Config{}, err
	}

	useCommas, err := promptBool(reader, stdout, "Use commas", current.UseCommas)
	if err != nil {
		return config.Config{}, err
	}

	copySingle, err := promptBool(reader, stdout, "Copy single results to clipboard", current.CopySingleToClipboard)
	if err != nil {
		return config.Config{}, err
	}

	return config.Config{
		LowerBound:            lower,
		UpperBound:            upper,
		UseCommas:             useCommas,
		CopySingleToClipboard: copySingle,
	}, nil
}

func promptInt(reader *bufio.Reader, stdout io.Writer, label string, current, min, max int) (int, error) {
	for {
		_, _ = fmt.Fprintf(stdout, "%s [%d]: ", label, current)

		answer, err := readPromptLine(reader)
		if err != nil {
			return 0, err
		}

		if answer == "" {
			return current, nil
		}

		value, err := strconv.Atoi(answer)
		if err != nil {
			_, _ = fmt.Fprintf(stdout, "Enter a whole number between %d and %d.\n", min, max)
			continue
		}

		if value < min || value > max {
			_, _ = fmt.Fprintf(stdout, "Enter a value between %d and %d.\n", min, max)
			continue
		}

		return value, nil
	}
}

func promptBool(reader *bufio.Reader, stdout io.Writer, label string, current bool) (bool, error) {
	prompt := "[y/N]"
	if current {
		prompt = "[Y/n]"
	}

	for {
		_, _ = fmt.Fprintf(stdout, "%s %s: ", label, prompt)

		answer, err := readPromptLine(reader)
		if err != nil {
			return false, err
		}

		switch strings.ToLower(answer) {
		case "":
			return current, nil
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			_, _ = fmt.Fprintln(stdout, "Enter y or n.")
		}
	}
}

func readPromptLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) && line != "" {
			err = nil
		} else if errors.Is(err, io.EOF) {
			return "", fmt.Errorf("interactive setup canceled")
		} else {
			return "", err
		}
	}

	return strings.TrimSpace(line), err
}

func parseRangeArg(raw string) ([]powers.Entry, bool, error) {
	const (
		noSeparator = iota
		dotSeparator
		hyphenSeparator
	)

	separatorType := noSeparator
	switch {
	case strings.Contains(raw, ".."):
		separatorType = dotSeparator
	case strings.Count(raw, "-") == 1 && strings.Index(raw, "-") > 0:
		separatorType = hyphenSeparator
	default:
		return nil, false, nil
	}

	var startRaw string
	var endRaw string

	switch separatorType {
	case dotSeparator:
		if strings.Count(raw, "..") != 1 {
			return nil, true, fmt.Errorf("invalid range %q", raw)
		}
		if strings.Contains(raw, "-") {
			return nil, true, fmt.Errorf("invalid range %q", raw)
		}

		startRaw, endRaw, _ = strings.Cut(raw, "..")
	case hyphenSeparator:
		if strings.Contains(raw, ".") || strings.Count(raw, "-") != 1 {
			return nil, true, fmt.Errorf("invalid range %q", raw)
		}

		startRaw, endRaw, _ = strings.Cut(raw, "-")
	}

	start, err := parseRangeEndpoint(startRaw)
	if err != nil {
		return nil, true, fmt.Errorf("invalid range %q", raw)
	}

	end, err := parseRangeEndpoint(endRaw)
	if err != nil {
		return nil, true, fmt.Errorf("invalid range %q", raw)
	}

	if start > int(powers.MaxExponent) || end > int(powers.MaxExponent) {
		return nil, true, fmt.Errorf("range exponents must be between 0 and %d", powers.MaxExponent)
	}

	if start > end {
		start, end = end, start
	}

	return powers.Between(start, end), true, nil
}

func parseIntegerArg(raw string) (*big.Int, bool) {
	normalized, ok := normalizeIntegerArg(raw)
	if !ok {
		return nil, false
	}

	input, ok := new(big.Int).SetString(normalized, 10)
	if !ok {
		return nil, false
	}

	return input, true
}

func parseRangeEndpoint(raw string) (int, error) {
	if raw == "" || !isDigits(raw) {
		return 0, fmt.Errorf("invalid endpoint")
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func normalizeIntegerArg(raw string) (string, bool) {
	if raw == "" {
		return "", false
	}

	sign := ""
	digits := raw
	if raw[0] == '+' || raw[0] == '-' {
		sign = raw[:1]
		digits = raw[1:]
	}

	if digits == "" {
		return "", false
	}

	if !strings.Contains(digits, ",") {
		if isDigits(digits) {
			return sign + digits, true
		}
		return "", false
	}

	groups := strings.Split(digits, ",")
	if len(groups) == 0 || len(groups[0]) == 0 || len(groups[0]) > 3 || !isDigits(groups[0]) {
		return "", false
	}

	var builder strings.Builder
	builder.Grow(len(raw))
	builder.WriteString(sign)
	builder.WriteString(groups[0])

	for _, group := range groups[1:] {
		if len(group) != 3 || !isDigits(group) {
			return "", false
		}
		builder.WriteString(group)
	}

	return builder.String(), true
}

func looksLikeFlag(raw string) bool {
	return strings.HasPrefix(raw, "-") && (len(raw) < 2 || raw[1] < '0' || raw[1] > '9')
}

func isDigits(value string) bool {
	if value == "" {
		return false
	}

	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}
