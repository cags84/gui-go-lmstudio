# LM Studio Controller

![CI](https://github.com/cags84/gui-go-lmstudio/actions/workflows/ci.yml/badge.svg)

A cross-platform desktop GUI for the [LM Studio](https://lmstudio.ai/) local server. It drives the bundled `lms` CLI so you can start or stop the server, choose localhost-only or LAN access, toggle CORS, and manage models — without a terminal.

**Status: early development.** Only the Wails application scaffold exists today. No LM Studio integration has been built yet.

## Why this exists

LM Studio already ships `lms`, a capable CLI for controlling its local server and models. This project wraps that CLI in a desktop GUI so you get live visibility (status, loaded models, streaming logs) alongside the actions, instead of juggling separate terminal commands.

## Roadmap

- [ ] Server control: start/stop, port selection, localhost vs. LAN binding, CORS toggle
- [ ] Model management: list, load, and unload models
- [ ] Live log stream from `lms log stream`
- [ ] System panel and tray icon
- [ ] Packaged releases for macOS, Windows, and Linux

**Later:**

- [ ] Model download with progress
- [ ] Native REST API polling (in addition to the CLI)
- [ ] LM Link support
- [ ] Chat interface

## Requirements

| Dependency | Version |
|---|---|
| LM Studio (with `lms` on `PATH`) | 0.4.x |
| Go | 1.25+ |
| Node.js (development only) | 20+ |

## Development

Install the Wails v3 CLI:

```sh
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

Install frontend dependencies:

```sh
npm --prefix frontend install
```

Run in development mode (hot reload for both frontend and backend):

```sh
wails3 dev
```

Build a production binary:

```sh
wails3 build
```

Package a distributable installer/bundle:

```sh
wails3 package
```

### Platform notes

- **Linux**: requires `libwebkit2gtk-4.1-dev` to build, and the matching runtime library (`libwebkit2gtk-4.1-0` or your distribution's equivalent) to run.
- **macOS** and **Windows**: use the system-provided webview; no extra runtime dependency.

## Security

Binding the LM Studio server to `0.0.0.0` exposes the model server to every device on your local network, with no authentication in front of it. The app will warn before enabling LAN binding.

## License

[MIT](LICENSE)
