# Hello Example Extension

A minimal bob extension that demonstrates the WASM ABI. It:

- Subscribes to `session_start` → logs a greeting and sets the status bar key `hello` to `world`

## Prerequisites

Install [TinyGo](https://tinygo.org/getting-started/install/) v0.32 or later.

```sh
# macOS
brew install tinygo
```

## Build

```sh
cd extensions/example
tinygo build -o hello.wasm -target wasi .
```

The output is a WASI-compatible `.wasm` binary (~50–100 KB).

## Install

Copy the binary to your extensions directory:

```sh
cp hello.wasm ~/.config/bob/extensions/
```

Or point `BOB_EXTENSIONS_DIR` at any directory containing `.wasm` files:

```sh
BOB_EXTENSIONS_DIR=./my-extensions ./bob
```

## What it does

| Trigger | Action |
|---------|--------|
| Session starts | Logs "hello extension: session started" (level: info) |
| Session starts | Sets status bar: `hello = world` |

## Extension lifecycle

```
load → _init() → subscribe("session_start")
                ↓
event arrives → _on_event(ptr, len) → switch on event type
                ↓
session_start → host_call("set_status", {"key":"hello","value":"world"})
```

## Adapting this example

To create your own extension:

1. Copy this directory to a new location
2. Update `go.mod` with your module path
3. Modify `_init` to subscribe to the events you need
4. Implement your logic in `_on_event`
5. Build with `tinygo build -o myext.wasm -target wasi .`

See `bob/docs/extensions.md` for the full WASM ABI reference.
