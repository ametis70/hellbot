# Configuration

hellbot is configured via a YAML file. The file location is resolved in this order:

1. `--config <path>` CLI flag
2. `HELLBOT_CONFIG` environment variable
3. `./config.yml` in the working directory

## Top-level options

| Field           | Type     | Default | Description                                                                                                      |
| --------------- | -------- | ------- | ---------------------------------------------------------------------------------------------------------------- |
| `poll_interval` | duration | `60s`   | How often to poll the Helldivers API. Accepts Go duration strings: `30s`, `2m`, `1h`.                            |
| `timezone`      | string   | `UTC`   | Global display timezone (IANA format). Used by notifiers that format timestamps. Can be overridden per notifier. |
| `store`         | object   | —       | Backing store configuration. See [Store](#store). Defaults to in-memory if omitted.                              |
| `notifiers`     | list     | `[]`    | List of notifier configurations. See [Notifiers](#notifiers).                                                    |

## Store

hellbot uses a backing store to persist the last known campaign state and ongoing events across restarts. The default is an in-memory store (state is lost on restart).

```yaml
store:
  type: <string>  # required — store type (memory, valkey)
  options:        # optional — type-specific options
    ...
```

### `memory` (default)

No options. State is lost when the process exits. This is the default when `store:` is omitted entirely.

```yaml
store:
  type: memory
```

---

### `sqlite`

Persists state in a local SQLite database file. State survives restarts with no external service required.

**Options**

| Field  | Type   | Default      | Description                                                                 |
|--------|--------|--------------|-----------------------------------------------------------------------------|
| `path` | string | `hellbot.db` | File path for the SQLite database. Supports `${ENV_VAR}`. Use `:memory:` for a non-persistent in-process database. |

**Example**

```yaml
store:
  type: sqlite
  options:
    path: "./data/hellbot.db"
```

**Example — Docker Compose**

```yaml
services:
  hellbot:
    image: ghcr.io/ametis70/hellbot:latest
    volumes:
      - ./config.yml:/app/config.yml
      - hellbot-data:/app/data
    restart: unless-stopped

volumes:
  hellbot-data:
```

```yaml
store:
  type: sqlite
  options:
    path: "/app/data/hellbot.db"
```

---

### `valkey`

Persists state in a [Valkey](https://valkey.io/) or Redis server. State survives restarts with no local file required.

**Options**

| Field      | Type   | Default            | Description                                                        |
|------------|--------|--------------------|--------------------------------------------------------------------|
| `addr`     | string | `localhost:6379`   | `host:port` of the Valkey/Redis server. Supports `${ENV_VAR}`.    |
| `password` | string | —                  | Server password (if auth is required). Supports `${ENV_VAR}`.     |
| `db`       | int    | `0`                | Redis database index (0–15).                                       |

**Example**

```yaml
store:
  type: valkey
  options:
    addr: "localhost:6379"
    password: "${VALKEY_PASSWORD}"
    db: 0
```

**Example — Docker Compose with a Valkey sidecar**

```yaml
services:
  hellbot:
    image: ghcr.io/ametis70/hellbot:latest
    volumes:
      - ./config.yml:/app/config.yml
    environment:
      VALKEY_PASSWORD: "s3cr3t"
    depends_on:
      - valkey
    restart: unless-stopped

  valkey:
    image: valkey/valkey:8-alpine
    command: valkey-server --requirepass s3cr3t
    restart: unless-stopped
```

---

## Notifiers

Each notifier has the same top-level shape:

```yaml
notifiers:
  - id: <string> # required — unique name, used in logs
    type: <string> # required — notifier type (stdout, ...)
    options: # optional — type-specific options
      ...
```

Multiple notifiers of the same type are supported. The `id` must be unique across all notifiers.

If no notifiers are configured, hellbot will still run and detect events — but nothing will be sent anywhere. A warning is logged at startup.

---

### `stdout`

Prints event notifications to standard output.

**Options**

| Field | Type | Default | Description |
|---|---|---|---|
| `timezone` | string | global `timezone` | Display timezone for timestamps. Overrides the global value. |
| `templates` | object | see [Templates](#templates) | Override default message templates. |

**Example**

```yaml
notifiers:
  - id: "console"
    type: stdout
    options:
      timezone: "America/Argentina/Buenos_Aires"
```

---

### `discord`

Sends event notifications to a Discord channel. See the [Discord Developer Documentation](https://discord.com/developers/docs/intro) for instructions on creating a bot and retrieving a token and channel ID.

**Options**

| Field | Type | Required | Description |
|---|---|---|---|
| `token` | string | yes (or `token_file`) | Discord bot token. Supports `${ENV_VAR}` interpolation. |
| `token_file` | string | yes (or `token`) | Path to a file containing the bot token. |
| `channel_id` | string | yes (or `channel_id_file`) | Discord channel ID. |
| `channel_id_file` | string | yes (or `channel_id`) | Path to a file containing the channel ID. |
| `guild_id` | string | no | Discord server (guild) ID. When set, slash commands (`/status`, `/statistics`) are registered as guild commands and appear instantly. When omitted, commands are registered globally and may take up to 1 hour to propagate. |
| `templates` | object | see [Templates](#templates) | Override default message templates. |

`token` and `token_file` are mutually exclusive. Same for `channel_id` and `channel_id_file`.

Times in Discord messages use Discord's native timestamp format (`<t:UNIX:f>`), which renders in the viewer's local timezone automatically — no timezone config needed.

For available slash commands, see [commands.md](commands.md).

**Example — env var**

```yaml
notifiers:
  - id: "my-server"
    type: discord
    options:
      token: "${DISCORD_TOKEN}"
      channel_id: "123456789012345678"
```

**Example — file-based secrets**

```yaml
notifiers:
  - id: "my-server"
    type: discord
    options:
      token_file: "/run/secrets/discord-token"
      channel_id_file: "/run/secrets/discord-channel-id"
```

---

### `telegram`

Sends event notifications to a Telegram chat. See the [Telegram Bot API documentation](https://core.telegram.org/bots) for instructions on creating a bot and retrieving a token and chat ID.

**Options**

| Field | Type | Required | Description |
|---|---|---|---|
| `token` | string | yes (or `token_file`) | Telegram bot token. Supports `${ENV_VAR}` interpolation. |
| `token_file` | string | yes (or `token`) | Path to a file containing the bot token. |
| `chat_id` | string | yes (or `chat_id_file`) | Telegram chat ID. |
| `chat_id_file` | string | yes (or `chat_id`) | Path to a file containing the chat ID. |
| `timezone` | string | global `timezone` | Display timezone for timestamps. Overrides the global value. |
| `templates` | object | see [Templates](#templates) | Override default message templates. |

`token` and `token_file` are mutually exclusive. Same for `chat_id` and `chat_id_file`.

For available bot commands, see [commands.md](commands.md).

**Example**

```yaml
notifiers:
  - id: "my-group"
    type: telegram
    options:
      token: "${TELEGRAM_TOKEN}"
      chat_id: "${TELEGRAM_CHAT_ID}"
      timezone: "Europe/Lisbon"
```

---

## Templates

Every notifier supports a `templates` block to override default message formats. All fields are optional — only set what you want to change, the rest uses the adapter's defaults.

### Template keys

| Key | Trigger |
|---|---|
| `defend_region_started` | A defend event starts in a normal region |
| `defend_super_earth_started` | A defend event starts in Super Earth (region 0) |
| `defend_region_succeeded` | A defend event in a normal region is won |
| `defend_super_earth_succeeded` | A defend event in Super Earth is won |
| `defend_region_failed` | A defend event in a normal region is lost |
| `defend_super_earth_failed` | A defend event in Super Earth is lost |
| `attack_homeworld_started` | An attack event against a faction homeworld starts |
| `attack_succeeded` | An attack event is won |
| `attack_failed` | An attack event is lost |
| `war_won` | The war ends with all enemy factions defeated |
| `war_lost` | The war ends without all factions being defeated (e.g. Super Earth fell) |

### Template variables

| Variable | Description | Example |
|---|---|---|
| `{FACTION}` | Enemy faction name | `Illuminate` |
| `{SEASON}` | War (season) number — available in `war_won` and `war_lost` | `159` |
| `{REGION_NAME}` | Region name | `Orionis Region` |
| `{REGION_NUMBER}` | Region number | `5` |
| `{TOTAL_REGIONS}` | Total regions per faction | `10` |
| `{START_TIME_FORMATTED}` | Start time formatted by the adapter | `2026-07-19T19:59:01Z` |
| `{END_TIME_FORMATTED}` | End time formatted by the adapter | `2026-07-21T19:59:01Z` |
| `{START_TIME_UNIX}` | Start time as Unix timestamp | `1784501941` |
| `{END_TIME_UNIX}` | End time as Unix timestamp | `1784674741` |
| `{PLAYERS}` | Players at event start | `184` |

For Discord, use `<t:{END_TIME_UNIX}:f>` to get native Discord timestamp rendering in the viewer's local timezone.

For stdout with ANSI colors, use escape sequences in the template string directly.

### Example — custom Discord templates

```yaml
notifiers:
  - id: "my-server"
    type: discord
    options:
      token: "${DISCORD_TOKEN}"
      channel_id: "123456789012345678"
      templates:
        defend_region_started: "🛡️ **{FACTION} offensive on {REGION_NAME} ({REGION_NUMBER}/{TOTAL_REGIONS})** — ends <t:{END_TIME_UNIX}:R>"
        defend_super_earth_started: "🚨 @everyone **SUPER EARTH IS UNDER ATTACK BY THE {FACTION}!** Ends <t:{END_TIME_UNIX}:R>"
        attack_succeeded: "🎉 Victory! The {FACTION} were crushed."
```

### Example — translated messages

```yaml
notifiers:
  - id: "pt-stdout"
    type: stdout
    options:
      timezone: "America/Sao_Paulo"
      templates:
        defend_region_started: "[defesa] iniciada — {FACTION} atacando {REGION_NAME} ({REGION_NUMBER}/{TOTAL_REGIONS}), termina {END_TIME_FORMATTED}"
        defend_super_earth_started: "[defesa] iniciada — {FACTION} atacando a Super Terra, termina {END_TIME_FORMATTED}"
        attack_succeeded: "[ataque] sucesso — {FACTION} derrotados"
```

---

## Secret resolution

String values in notifier options support two forms of secret injection:

### Environment variable interpolation

Use `${VAR_NAME}` syntax. The value is replaced at startup with the environment variable's value. If the variable is not set, hellbot will fail to start.

```yaml
options:
  token: "${DISCORD_TOKEN}"
```

### File-based secrets

Use a `_file` variant of the field (e.g. `token_file`) to read the value from a file. The file contents are trimmed of leading/trailing whitespace.

```yaml
options:
  token_file: "/run/secrets/discord-token"
```

`token` and `token_file` are mutually exclusive — specifying both is an error.

---

## Full example

```yaml
poll_interval: 60s
timezone: "UTC"

store:
  type: valkey
  options:
    addr: "localhost:6379"

notifiers:
  - id: "stdout-dev"
    type: stdout
    options:
      timezone: "Europe/Lisbon"
```

---

## Validation

hellbot validates the config at startup and exits immediately if:

- A notifier is missing an `id`
- Two notifiers share the same `id`
- An unknown notifier `type` is specified
- An unknown store `type` is specified
- A timezone string is invalid
- `poll_interval` is not a valid Go duration
- A required field is missing or has conflicting values (e.g. both `token` and `token_file` set)
