# Agent Instructions

## Documentation Policy

- Always document infrastructure and server configuration decisions in Markdown files under `doc/`.
- When adding or updating configuration, create or update the corresponding document in `doc/` within the same change.
- Write all records under `doc/` in Japanese.
- Respond in Japanese for all chat communication.

## Minecraft Notes

- For Minecraft Java 1.21.11 and later, be careful with renamed gamerules and do not assume legacy names still work.
- Prefer checking current gamerule names before proposing commands (tab completion on server, current docs/changelog).
- Example renames to keep in mind:
  - `doDaylightCycle` -> `advance_time`
  - `doMobSpawning` -> `spawn_mobs`
