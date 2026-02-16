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
- `block_display` placement uses entity coordinates and scales from the summon origin toward positive local axes.
- For gate-like planes, define one thin axis in `transformation.scale` (for example `0.2`) and tune summon coordinates along only that axis.
- To center a scaled display on one block axis, apply offset `+(1 - scale) / 2` on that axis (example: `scale=0.2` -> `+0.4`).
- If alignment differs by gate direction, adjust per-axis (`x` for east/west gates, `z` for north/south gates) in `0.1` steps and re-apply function.
- Preferred workflow: keep a single tagged display per gate, `kill` by tag, `summon`, then `data merge` for scale to avoid stacked leftovers.
- For sign text on Java 1.21.11+, use `front_text.messages`/`back_text.messages` JSON components, not legacy `Text1`-`Text4`.
- In sign components, use snake_case event keys (`click_event`, `hover_event`), not camelCase (`clickEvent`).
- Recommended sequence: `setblock ... oak_sign[...]` then `data merge block ... {front_text:{messages:[...]}}`.
- Quick validation: run `data get block <x> <y> <z>` and confirm JSON is stored under `front_text.messages`.

## Change Hygiene

- Separate work into two phases: exploration (trial/error allowed) and convergence (final requirement only).
- Before editing, define a short final requirement and keep only changes that directly satisfy it.
- Do not keep residue from discussion or failed attempts (temporary options, negation notes, abandoned branches).
- In docs, describe only the current accepted behavior and reproducible procedure.
- Do not include process history such as "we considered X", "X is not used", or similar transient discussion context.
- If historical rationale is required, place it in a dedicated decision log, not in operational setup docs.
- Before commit, run a residue check: stale names, temporary comments, contradictory statements, and implementation/doc mismatch.
