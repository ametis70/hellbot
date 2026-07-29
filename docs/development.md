# Development

## Prerequisites

Enter the Nix dev shell to get all required tools on `$PATH` (Go, gopls, gotools, valkey):

```bash
nix develop
# or, with direnv:
direnv allow
```

## Running tests

```bash
go test ./...

# with race detector
go test -race ./...
```

## Building

```bash
go build ./cmd/hellbot
```

## Mock server

The built-in mock server lets you run hellbot end-to-end without a real Helldivers API connection. Instead of polling the actual API, it plays back a scripted war scenario — one response per poll tick — and exits automatically when the scenario is exhausted.

This is useful for verifying that your notifiers (Discord, Telegram, stdout) receive the correct messages without waiting for real in-game events.

### Scenario

The mock plays back a fixed 9-poll war sequence:

| Poll | Event |
|------|-------|
| 1 | Idle — baseline stored, no diff |
| 2 | Attack event starts → **attack started** notification |
| 3 | Attack still active |
| 4 | Attack succeeds → **attack succeeded** notification |
| 5 | Idle |
| 6 | Defend event starts → **defend started** notification |
| 7 | Defend still active |
| 8 | Defend succeeds, all factions defeated |
| 9 | New season → **war won** notification |

5 notifications total: attack started, attack succeeded, defend started, defend succeeded, war won.

### Usage

Add a `dev` block to your `config.yml` and set `poll_interval` to a short duration so the scenario plays back quickly:

```yaml
poll_interval: 1s

dev:
  mock_server: true

notifiers:
  - id: console
    type: stdout
```

Then run hellbot normally:

```bash
go run ./cmd/hellbot
```

At startup you will see:

```
level=WARN msg="dev.mock_server is enabled — using built-in war scenario, not the real API"
```

The bot will serve all 9 frames, fire the 5 notifications through your configured notifiers, then stop cleanly.

### Notes

- `dev.mock_server` is mutually exclusive with `dev.api_url` — when `mock_server` is `true`, `api_url` is ignored.
- The mock uses an in-memory store by default unless you configure a different `store` type. Using `memory` means state resets on every run, which is usually what you want for scenario testing.
- The scenario is hardcoded in `internal/adapter/api/mock/mock.go`. Edit `warScenario()` there to customise the event sequence.

---



SQLite requires no external service — it's the simplest way to test persistent state locally.

Add a `store` block to your `config.yml`:

```yaml
store:
  type: sqlite
  options:
    path: "./hellbot.db"
```

Run hellbot:

```bash
go run ./cmd/hellbot
```

You should see:

```
level=INFO msg="store initialized" type=sqlite path=./hellbot.db
```

Inspect the database with any SQLite client:

```bash
sqlite3 hellbot.db "SELECT payload FROM campaign;"
sqlite3 hellbot.db "SELECT * FROM ongoing_events;"
```

Delete the file to reset state:

```bash
rm hellbot.db
```

---

## Testing with Valkey

The dev shell includes `valkey-server` and `valkey-cli`. To test the `valkey` store locally:

**1. Start the server**

```bash
valkey-server --logfile ./valkey.log
```

This starts Valkey in the foreground on `localhost:6379`. Logs go to `./valkey.log`. Press `Ctrl+C` to stop it, or run it daemonized:

```bash
valkey-server --daemonize yes --logfile ./valkey.log
```

**2. Configure hellbot to use it**

Add a `store` block to your `config.yml`:

```yaml
store:
  type: valkey
  options:
    addr: "localhost:6379"
```

**3. Run hellbot**

```bash
go run ./cmd/hellbot
```

You should see:

```
level=INFO msg="store initialized" type=valkey addr=localhost:6379
```

**4. Inspect stored state**

```bash
# last known campaign (JSON)
valkey-cli get hellbot:campaign

# set of ongoing event keys
valkey-cli smembers hellbot:events
```

**5. Stop the daemonized server**

```bash
valkey-cli shutdown nosave
```
