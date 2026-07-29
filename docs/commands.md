# Commands

hellbot supports interactive commands on Discord (slash commands) and Telegram (bot commands). The stdout notifier does not support commands.

Commands are registered automatically at startup — no extra configuration is needed beyond setting up the notifier.

---

## Available commands

| Command | Discord | Telegram | Description |
|---|---|---|---|
| `/status` | ✅ | ✅ | War progress and current sector status per faction. Optional faction filter. |
| `/statistics` | ✅ | ✅ | Cumulative war statistics with all factions summed. |
| `/test` | ❌ | ✅ | Connectivity test — confirms the bot can send messages. |

---

## `/status`

Shows the current war progress for all factions: overall completion, points, current sector, and sector progress. Active defend and attack events are listed at the bottom.

**Usage**

- Discord: `/status` or `/status faction:bugs` (dropdown choice)
- Telegram: `/status` or `/status bugs` / `/status cyborgs` / `/status illuminate`

**Example — all factions**

```
War 160 — Status

The Bugs (active)
  [██████░░░░]  63%
  247,778 / 387,510 pts

  Sector 7/11: Higgs Region
  [███░░░░░░░]  39%
  15,272 / 38,751 pts


The Cyborgs (active)
  [████████░░]  82%
  237,152 / 289,050 pts

  Sector 9/11: Ceti System
  [██░░░░░░░░]  20%
  5,912 / 28,905 pts


The Illuminate (active)
  [██░░░░░░░░]  29%
  53,039 / 181,220 pts

  Sector 3/11: Procyon Region  ⚔️ defending Orionis Region
  [█████████░]  92%
  16,795 / 18,122 pts


Active events:
  ⚔️  The Illuminate are attacking sector 5: Orionis Region — ends 14h32m
```

**Example — filtered by faction (`/status bugs`)**

```
War 160 — Status

The Bugs (active)
  [██████░░░░]  63%
  247,778 / 387,510 pts

  Sector 7/11: Higgs Region
  [███░░░░░░░]  39%
  15,272 / 38,751 pts
```

---

## `/statistics`

Shows cumulative war statistics with all factions summed into a single total.

**Usage**

- Discord: `/statistics`
- Telegram: `/statistics`

**Example**

```
War 160 — Statistics

Players online:     12,450
Total players:      98,234
Kills:           1,234,567
Deaths:             87,654
Accidentals:         3,210
Shots fired:    45,678,900
Accuracy:               27%
Missions:           45,000 (38,000 successful, 84%)
Defend events:         120 (98 successful, 82%)
Attack events:          34 (28 successful, 82%)
Planets liberated:      42
```

---

## Discord setup notes

Slash commands require the `applications.commands` OAuth2 scope when adding the bot to your server. If you added the bot without this scope, re-invite it using the OAuth2 URL generator in the [Discord Developer Portal](https://discord.com/developers/applications) with both `bot` and `applications.commands` selected.

By default, commands are registered globally and may take up to 1 hour to appear. Set `guild_id` in the Discord notifier config to register them instantly for a specific server — see [config.md](config.md#discord) for details.

---

## Telegram setup notes

Commands are received via long-polling — the bot listens continuously while hellbot is running. The `/status` and `/statistics` commands reflect the last cached campaign state (updated every `poll_interval`).
