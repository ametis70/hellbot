# Helm

hellbot ships a Helm chart for deploying to Kubernetes. The chart is published as an OCI artifact alongside the container image.

## Prerequisites

- Kubernetes 1.21+
- Helm 3.8+ (OCI support is stable from 3.8)

## Installation

```bash
helm install hellbot oci://ghcr.io/ametis70/hellbot/chart \
  --version 1.0.0 \
  --namespace hellbot \
  --create-namespace \
  -f my-values.yaml
```

## Upgrading

```bash
helm upgrade hellbot oci://ghcr.io/ametis70/hellbot/chart \
  --version 1.1.0 \
  --namespace hellbot \
  -f my-values.yaml
```

## Configuration

All configuration is done through a values file. The sections below cover each use case. See the inline comments in the chart's `values.yaml` for the full reference.

### Basic example — stdout notifier, memory store

```yaml
notifiers:
  - id: console
    type: stdout
    options:
      timezone: "America/New_York"
```

### Persistent store

#### SQLite

SQLite is the simplest option — no external service required. The chart creates a `PersistentVolumeClaim` automatically.

```yaml
store:
  type: sqlite
  sqlite:
    path: "/data/hellbot.db"
    persistence:
      enabled: true
      size: 10Mi
      storageClassName: ""  # leave empty to use the cluster default
```

#### Valkey / Redis

```yaml
store:
  type: valkey
  valkey:
    addr: "valkey-service:6379"
    db: 0
    password:
      existingSecret: "hellbot-secrets"
      existingSecretKey: "valkey-password"
```

### Notifiers

#### Discord

```yaml
notifiers:
  - id: my-server
    type: discord
    options:
      token:
        existingSecret: "hellbot-secrets"
        existingSecretKey: "discord-token"
      channel_id:
        existingSecret: "hellbot-secrets"
        existingSecretKey: "discord-channel-id"
      guild_id: "987654321098765432"   # optional, for instant slash command registration
      templates:
        defend_super_earth_started: "🚨 @everyone Super Earth is under attack by the {FACTION}!"
```

#### Telegram

```yaml
notifiers:
  - id: my-group
    type: telegram
    options:
      token:
        existingSecret: "hellbot-secrets"
        existingSecretKey: "telegram-token"
      chat_id:
        existingSecret: "hellbot-secrets"
        existingSecretKey: "telegram-chat-id"
      timezone: "Europe/Lisbon"
```

#### Webhook

```yaml
notifiers:
  - id: my-webhook
    type: webhook
    options:
      url:
        existingSecret: "hellbot-secrets"
        existingSecretKey: "webhook-url"
      secret_header: "Authorization"
      secret_value:
        existingSecret: "hellbot-secrets"
        existingSecretKey: "webhook-secret"
```

### Secrets

Sensitive fields (`token`, `channel_id`, `chat_id`, `url`, `secret_value`, `password`) accept either a plain inline string or a reference to an existing Kubernetes Secret.

**Inline** — the chart creates and manages the Secret:

```yaml
notifiers:
  - id: my-server
    type: discord
    options:
      token: "OTE5MjEwOTE1MzAyMzUw..."
      channel_id: "123456789012345678"
```

**Existing Secret reference** — you manage the Secret, the chart mounts it:

```yaml
notifiers:
  - id: my-server
    type: discord
    options:
      token:
        existingSecret: "hellbot-secrets"
        existingSecretKey: "discord-token"
```

In both cases the value is mounted as a file under `/run/secrets/` inside the container. The config never contains plaintext secrets.

The existing Secret can be created with `kubectl`:

```bash
kubectl create secret generic hellbot-secrets \
  --namespace hellbot \
  --from-literal=discord-token="OTE5MjEwOTE1MzAyMzUw..." \
  --from-literal=discord-channel-id="123456789012345678"
```

Or managed by [External Secrets Operator](https://external-secrets.io), SOPS, or any other secrets management tool.

## FluxCD

### OCIRepository

```yaml
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: OCIRepository
metadata:
  name: hellbot
  namespace: flux-system
spec:
  interval: 1h
  url: oci://ghcr.io/ametis70/hellbot/chart
  ref:
    tag: 1.0.0
```

### HelmRelease

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: hellbot
  namespace: hellbot
spec:
  interval: 1h
  chartRef:
    kind: OCIRepository
    name: hellbot
    namespace: flux-system
  values:
    pollInterval: "60s"
    timezone: "UTC"

    store:
      type: sqlite

    notifiers:
      - id: my-server
        type: discord
        options:
          token:
            existingSecret: "hellbot-secrets"
            existingSecretKey: "discord-token"
          channel_id:
            existingSecret: "hellbot-secrets"
            existingSecretKey: "discord-channel-id"
```

With External Secrets Operator, the `hellbot-secrets` Secret is created automatically from your secrets backend before the HelmRelease is reconciled.
