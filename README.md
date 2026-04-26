# SubConv Next

`subconv-next` is a Docker-first subscription converter and Mihomo YAML generator.

This repository is being implemented from the Markdown specification bundle in
`/root/subconv-next-openwrt-codex-docs`, but the active implementation target
is a Docker-oriented V1 with a native embedded Web UI and JSON config.

## Planned Commands

- `subconv-next serve --config ./config/config.json`
- `subconv-next generate --config ./config/config.json --out /data/mihomo.yaml`
- `subconv-next parse --input ./input.txt --json`
- `subconv-next version`

## Current Status

- Repository skeleton created.
- Minimal CLI with `version` command implemented.
- JSON config is the active runtime path for Docker.
- Minimal HTTP daemon implemented with `serve`, `/healthz`, and `/api/status`.
- `parse --input ... --json` is implemented against the in-repo parser.
- `generate --config ... --out ...` renders Mihomo YAML from enabled inline sources.
- `/api/parse` and `/api/generate` are available on the local API.
- `/api/refresh` and `/sub/mihomo.yaml` now regenerate from inline sources and write the configured output file.
- `/api/logs?tail=200` exposes a simple in-memory log buffer for recent refresh/render events.
- `/` serves an embedded native HTML/CSS/JS management UI from `static/`.
- `/api/config` supports reading and writing the mounted JSON config file at `/config/config.json`.

## Docker

```sh
docker compose up -d --build
curl http://127.0.0.1:9876/healthz
open http://127.0.0.1:9876/
```

The sample [config/config.json](/root/subconv-next/config/config.json:1) enables one inline
sample node so `/sub/mihomo.yaml` works immediately after boot.
