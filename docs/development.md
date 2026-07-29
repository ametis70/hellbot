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
