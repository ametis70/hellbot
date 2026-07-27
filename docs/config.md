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
| `notifiers`     | list     | `[]`    | List of notifier configurations. See [Notifiers](#notifiers).                                                    |

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

Sends event notifications to a Discord channel.

**Prerequisites:**
1. Create an application in the [Discord Developer Portal](https://discord.com/developers/applications)
2. Enable the bot and copy the token
3. Use the OAuth2 URL generator with `bot` scope and `Send Messages` permission to add the bot to your server
4. Copy the [channel ID](https://support.discord.com/hc/en-us/articles/206346498) where alerts will be sent

**Options**

| Field | Type | Required | Description |
|---|---|---|---|
| `token` | string | yes (or `token_file`) | Discord bot token. Supports `${ENV_VAR}` interpolation. |
| `token_file` | string | yes (or `token`) | Path to a file containing the bot token. |
| `channel_id` | string | yes (or `channel_id_file`) | Discord channel ID. |
| `channel_id_file` | string | yes (or `channel_id`) | Path to a file containing the channel ID. |

`token` and `token_file` are mutually exclusive. Same for `channel_id` and `channel_id_file`.

Times in Discord messages use Discord's native timestamp format (`<t:UNIX:f>`), which renders in the viewer's local timezone automatically — no timezone config needed.

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
- A timezone string is invalid
- `poll_interval` is not a valid Go duration
- A required field is missing or has conflicting values (e.g. both `token` and `token_file` set)
