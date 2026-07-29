<div align="center">

# hellbot

![license](https://img.shields.io/github/license/ametis70/hellbot?style=flat-square)
![go](https://img.shields.io/badge/go-1.26-00ADD8?style=flat-square&logo=go&logoColor=white)
![nix](https://img.shields.io/badge/nix-flake-5277C3?style=flat-square&logo=nixos&logoColor=white)
![build](https://img.shields.io/github/actions/workflow/status/ametis70/hellbot/build-and-publish.yml?branch=main&style=flat-square)
![tests](https://img.shields.io/endpoint?style=flat-square&url=https://gist.githubusercontent.com/ametis70/d4d99431f51f28cef15d3fe6ab985e4a/raw/hellbot-go-tests.json)
![coverage](https://img.shields.io/endpoint?style=flat-square&url=https://gist.githubusercontent.com/ametis70/d4d99431f51f28cef15d3fe6ab985e4a/raw/hellbot-go-coverage.json)

hellbot monitors the [HELLDIVERS™](https://www.arrowheadgamestudios.com/aboutarrowhead/games/helldivers/) galactic campaign and broadcasts alerts when defend and attack events start or end

</div>

## What it does

- Polls the official Helldivers 1 API on a configurable interval
- Detects when defend events, attack events, and wars start, succeed, or fail
- Sends notifications to one or more configured notifiers simultaneously
- Supports **Discord**, **Telegram**, **stdout**, and **webhook** as notification targets
- Persists state across restarts via a configurable store (**memory**, **SQLite**, or **Valkey/Redis**)
- Supports fully customizable message templates per notifier
- Supports per-notifier timezone overrides for timestamp formatting

## Configuration

hellbot is configured via a YAML file. See [docs/config.md](docs/config.md) for all available options and examples.

The config file location is resolved in this order:
1. `--config <path>` CLI flag
2. `HELLBOT_CONFIG` environment variable
3. `./config.yml` in the working directory

## Running

```sh
docker run \
  -v ./config.yml:/app/config.yml \
  ghcr.io/ametis70/hellbot:latest
```

### With Docker Compose

```yml
services:
  hellbot:
    image: ghcr.io/ametis70/hellbot:latest
    volumes:
      - ./config.yml:/app/config.yml
    restart: unless-stopped
```

## Development

See [docs/development.md](docs/development.md) for setup, running tests, building, and testing stores locally.

## Contributing

Bug reports and feature requests are welcome via issues. Pull requests are also welcome — run `go test ./...` and `gofmt` before submitting.
