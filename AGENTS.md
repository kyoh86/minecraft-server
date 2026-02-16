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

## Change Hygiene

- Separate work into two phases: exploration (trial/error allowed) and convergence (final requirement only).
- Before editing, define a short final requirement and keep only changes that directly satisfy it.
- Do not keep residue from discussion or failed attempts (temporary options, negation notes, abandoned branches).
- In docs, describe only the current accepted behavior and reproducible procedure.
- Do not include process history such as "we considered X", "X is not used", or similar transient discussion context.
- If historical rationale is required, place it in a dedicated decision log, not in operational setup docs.
- Before commit, run a residue check: stale names, temporary comments, contradictory statements, and implementation/doc mismatch.
