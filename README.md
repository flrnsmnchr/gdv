# gdv

gdv is a minimal terminal-based Git change viewer written in Go. It shows changed files from the current Git repository and enables quick diff browsing in a terminal UI using `gocui`.

## Features

- Lists modified files from `git status --porcelain`
- Opens file diffs with `git diff HEAD`
- Supports side views for old/new file contents
- Keyboard controls for navigation and exit

## Installation

1. Ensure Go 1.26+ is installed.
2. Clone the repository or place the code in a directory.
3. Run:

```bash
go install ./...
```

Or build a local binary:

```bash
go build -o gdv
```

## Usage

Run `gdv` from inside a Git repository:

```bash
gdv
```

Keyboard controls:

- `j`, `f`, `down` — move down in file list
- `k`, `d`, `up` — move up in file list
- `enter`, `space` — open selected file diff / return to file list
- `h` — show old file contents
- `l`, `g` — show new file contents
- `q`, `esc`, `Ctrl+C` — quit

## Testing

Run unit tests with:

```bash
go test ./...
```

## Notes

- `gdv` relies on Git being installed and available in `PATH`.
- It reads the repository status and diffs from the current working tree.
