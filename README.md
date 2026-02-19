# TUIStudio Desktop

Cross-platform desktop wrapper for [TUIStudio](https://github.com/jalonsogo/tui-studio) — a visual TUI designer.

Built with [Wails v2](https://wails.io) (Go + system WebView). Ships as:
- **macOS** — `.dmg` (~8 MB)
- **Windows** — NSIS `.exe` installer
- **Linux** — `.deb` package

---

## Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Wails CLI v2](https://wails.io/docs/gettingstarted/installation): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Linux only**: `libgtk-3-dev libwebkit2gtk-4.0-dev`

---

## Dev Workflow

```bash
# 1. Clone with submodule
git clone --recurse-submodules https://github.com/jalonsogo/tui-studio-desktop
cd tui-studio-desktop

# 2. Terminal 1 — Vite dev server (hot reload)
cd web && npm install && npm run dev
# Vite listens on http://localhost:5173

# 3. Terminal 2 — Wails desktop window
wails dev --frontenddevserverurl http://localhost:5173
```

> **Note:** Always pass `--frontenddevserverurl` explicitly — Wails v2 may not auto-detect Vite 7 dev servers.

---

## Production Build

```bash
wails build
# Output: build/bin/TUIStudio  (macOS app bundle / .exe / Linux binary)
```

---

## Release

Push a version tag to trigger the GitHub Actions matrix build:

```bash
git tag v1.0.0 && git push origin v1.0.0
```

Produces `.dmg`, `.exe`, and `.deb` attached to a GitHub Release.

---

## Updating the Web App

```bash
cd web && git pull origin main
cd .. && git add web && git commit -m "bump web submodule to latest"
```

---

## Architecture

```
tui-studio-desktop/
├── main.go               # Wails entry: window config, bridge injection
├── app.go                # Go IPC methods exposed as window.go.main.App.*
├── native-bridge/
│   └── bridge.js         # Overrides showSaveFilePicker / showOpenFilePicker / showDirectoryPicker
├── scripts/
│   └── build-web.mjs     # npm ci + npm run build inside web/
├── wails.json            # Wails config
└── web/                  # git submodule → jalonsogo/tui-studio
```

### Bridge Strategy

`OnDomReady` injects `native-bridge/bridge.js` via `runtime.WindowExecJS`. It overrides three browser APIs before any user interaction:

| Web App Call | Bridge Override | Go Method |
|---|---|---|
| `showSaveFilePicker()` | Fake writable handle | `NativeSaveDialog(name, content)` |
| `showOpenFilePicker()` | Fake readable handle | `NativeOpenDialog()` → content string |
| `showDirectoryPicker()` | Fake dir handle | `NativePickDirectory()` + `NativeWriteFile(path, content)` |

`localStorage` and `window.dispatchEvent` work natively in the Wails WebView — no overrides needed.
