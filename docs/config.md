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

| Field      | Type   | Default           | Description                                                  |
| ---------- | ------ | ----------------- | ------------------------------------------------------------ |
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
