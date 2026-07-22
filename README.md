<div align="center">

# hellbot

![license](https://img.shields.io/github/license/ametis70/hellbot?style=flat-square)
![go](https://img.shields.io/badge/go-1.26-00ADD8?style=flat-square&logo=go&logoColor=white)
![nix](https://img.shields.io/badge/nix-flake-5277C3?style=flat-square&logo=nixos&logoColor=white)
![build](https://img.shields.io/github/actions/workflow/status/ametis70/hellbot/main.yml?branch=main&style=flat-square)
![tests](https://img.shields.io/github/actions/workflow/status/ametis70/hellbot/test.yml?branch=main&style=flat-square&label=tests)

hellbot monitors the [HELLDIVERS™](https://www.arrowheadgamestudios.com/aboutarrowhead/games/helldivers/) galactic campaign and broadcasts alerts when defend and attack events start or end

</div>

## What it does

- Polls the official Helldivers 1 API every 60 seconds
- Detects when defend and attack events start, succeed, or fail
- Broadcasts notifications to one or more configured receivers

## Development

```bash
# Enter the dev shell (provides Go, gopls, and gotools)
nix develop

# Or with direnv
direnv allow

# Run tests
go test ./...

# Run with race detector
go test -race ./...

# Build
go build ./cmd/hellbot
```

## Configuration

hellbot is configured via environment variables:

| Variable | Description                                          | Required               |
| -------- | ---------------------------------------------------- | ---------------------- |
| `TZ`     | Display timezone (IANA format, e.g. `Europe/Lisbon`) | No (defaults to `UTC`) |

## Running

```sh
docker run \
  -e TZ=America/Argentina/Buenos_Aires \
  ghcr.io/ametis70/hellbot:latest
```

### With Docker Compose

```yml
services:
  hellbot:
    image: ghcr.io/ametis70/hellbot:latest
    environment:
      - TZ=${TZ}
    restart: unless-stopped
```

```sh
TZ=America/Argentina/Buenos_Aires docker compose up -d
```

## Contributing

Bug reports and feature requests are welcome via issues. Pull requests are also welcome — run `go test ./...` and `gofmt` before submitting.
