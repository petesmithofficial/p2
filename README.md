# p2

[![CI](https://github.com/petesmithofficial/p2/actions/workflows/ci.yml/badge.svg)](https://github.com/petesmithofficial/p2/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

`p2` is a tiny Go CLI for printing powers of 2 through `2^32`.

## What it does

- `p2` prints the default list from `2^0` to `2^16`
- `p2 5` prints `2^5 = 32`
- `p2 30000` finds the closest supported power of 2 and prints `2^15 = 32,768`
- Exact midpoint ties return both matches on separate lines, for example `p2 48` prints `2^5 = 32` and `2^6 = 64`
- By default, single-result lookups also copy the raw numeric value to your clipboard
- Optional user config controls list bounds, comma formatting, and clipboard copy for single-result lookups
- `p2 --config` walks you through a simple interactive setup
- `p2 --reset` restores a clean default config file

## Install

Install into `~/.local/bin`:

```sh
./install.sh
```

This install path builds `p2` from source, so it requires Go to be installed locally.

Install into a custom location:

```sh
BINDIR="$HOME/bin" ./install.sh
```

Build manually:

```sh
go build ./cmd/p2
```

## Usage

```sh
p2
p2 5
p2 30000
p2 48
p2 --help
p2 --config
p2 --reset
```

## Config

`p2` reads an optional user config file from your platform config directory:

- macOS: `~/Library/Application Support/p2/config.json`
- Linux: `~/.config/p2/config.json`
- Windows: `%AppData%\p2\config.json`

Example:

```json
{
  "lower_bound": 5,
  "upper_bound": 8,
  "use_commas": true,
  "copy_single_to_clipboard": true
}
```

Notes:

- `lower_bound` and `upper_bound` affect bare `p2` only
- fresh installs default to `upper_bound: 16`, even when no config file exists yet
- `use_commas` changes display formatting only
- `copy_single_to_clipboard` defaults to `true`
- single-result lookups copy the raw numeric value without commas
- the config file is created when you run `p2 --config` or `p2 --reset`

## Development

```sh
go test ./...
go build ./cmd/p2
```

## License

This project is licensed under the [MIT License](LICENSE).
