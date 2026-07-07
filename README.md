# ctrl

`ctrl` is a personal, OS-agnostic Go terminal dashboard with a clean retro style.

The first scaffold uses:

- `tview` for terminal UI widgets
- `tcell` through `tview` for cross-platform terminal handling
- `cobra` for a CLI shape that can grow
- JSON files under the user config directory for local data

On Unix-like systems, the default todo path is:

```text
~/.config/ctrl/todos.json
```

First-launch and display preferences are stored next to it:

```text
~/.config/ctrl/config.json
```

You can override that config directory with:

```text
CTRL_CONFIG_HOME=/some/path ctrl
```

## Development

```sh
make run
make build
make check
```

## Current dashboard

- Todo panel backed by JSON loading
- Environment panel with probes for Docker daemon, Node, Python, Go, and Foundry Forge
- Usage panel with RAM and primary disk availability
- Battery panel with charge state or no-battery status
- First-launch tutorial and reusable help overlay with `?`
- Manual refresh with `r`
- Automatic dashboard refresh every five minutes
- Quit with `q` or `Ctrl+C`

Todo editing is intentionally left for the next pass.
