# SubConv Next

SubConv Next is an open source, Docker-first subscription converter and Mihomo YAML generator with a native embedded Web UI.

It is designed for people who want to:

- aggregate multiple upstream subscriptions into one Mihomo subscription
- keep local node edits without rewriting the remote subscription source
- manage everything from a lightweight Web UI
- run the whole stack as a single Go service, without a frontend build chain

## Highlights

- Native embedded Web UI
  No React, Vue, Vite, Node.js, or external CDN. The UI is served directly from the Go binary.

- Multiple named subscription sources
  Add several upstream subscriptions, batch import old URL lists, and keep per-source metadata.

- Local node overrides
  Rename, hide, recover, and preserve node edits locally by stable node ID. Remote subscription content is never rewritten.

- Dynamic Mihomo subscription output
  `/sub/mihomo.yaml` supports cache, refresh-on-request, scheduled refresh, stale fallback, and optional access token protection.

- Modern protocol support
  `ss`, `vmess`, `vless`, `trojan`, `hysteria2`, `tuic`, `anytls`, `wireguard`, plus Base64 and mixed URI lists.

- Rich Mihomo renderer
  Includes rule presets, template mode, DNS/sniffer/geodata defaults, rule-provider catalog, and Mihomo config validation.

- Per-source subscription usage metadata
  Tracks traffic and expiry per subscription source, shows aggregated usage in the UI, and emits aggregated `Subscription-Userinfo`.

- Diagnostics and YAML preview
  Built-in logs, validation warnings, node summary, and full-page YAML preview.

- Security-oriented defaults
  Secrets are masked in UI responses and logs. The service focuses on config generation only.

## What SubConv Next Is Not

SubConv Next intentionally does not do the following:

- node speed tests
- connectivity tests
- port scanning
- proxy forwarding or relay
- rewriting remote subscription source content
- shipping a proxy core

The project focuses on subscription parsing, normalization, editing, and Mihomo config generation.

## Quick Start

### Docker Compose

```sh
docker compose up -d --build
curl -fsS http://127.0.0.1:9876/healthz
```

Open:

```text
http://127.0.0.1:9876/
```

The default runtime layout is:

- config: `./config/config.json`
- state: `./data/state.json`
- cache: `./data/cache/`
- rendered output: `./data/mihomo.yaml`

### Generated Subscription URL

After configuration is saved and refreshed, clients can subscribe to:

```text
http://<host>:9876/sub/mihomo.yaml
```

If `subscription_token` is enabled:

```text
http://<host>:9876/sub/mihomo.yaml?token=<token>
```

## Web UI Features

- Subscription source manager
- Conversion parameters and filters
- Output options
- Rule mode and template mode
- Node editor with local overrides
- Per-source subscription usage view
- Full-page YAML preview
- Diagnostics and runtime logs

## Refresh Pipeline

SubConv Next keeps user edits local and reapplies them on every refresh:

```text
fetch upstream
-> parse NodeIR[]
-> append inline nodes
-> append custom nodes
-> normalize
-> dedupe
-> assign stable node IDs
-> apply node overrides
-> apply disabled nodes
-> apply source prefix / naming rules
-> render Mihomo YAML
-> validate Mihomo config
-> write output
```

This means:

- upstream subscription URLs are read-only
- local edits survive refresh
- hidden nodes stay hidden until recovered

## Supported Commands

```sh
subconv-next serve --config ./config/config.json
subconv-next generate --config ./config/config.json --out ./data/mihomo.yaml
subconv-next parse --input ./input.txt --json
subconv-next version
```

## Selected HTTP API

Configuration and status:

- `GET /healthz`
- `GET /api/status`
- `GET /api/config`
- `PUT /api/config`

Refresh and output:

- `POST /api/refresh`
- `GET /sub/mihomo.yaml`
- `GET /api/logs?tail=200`

Nodes:

- `GET /api/nodes`
- `GET /api/nodes/:id`
- `PUT /api/nodes/:id/override`
- `POST /api/nodes/:id/reset`
- `POST /api/nodes/disable`
- `POST /api/nodes/enable`
- `POST /api/nodes/custom`
- `DELETE /api/nodes/custom/:id`
- `POST /api/nodes/validate`

Subscription metadata:

- `GET /api/subscription-meta`

## Configuration Model

SubConv Next separates persistent config from mutable runtime state:

- `config.json`
  User configuration, subscriptions, render options, rule selection, service options.

- `state.json`
  Node overrides, disabled node list, custom nodes, per-source subscription metadata.

This split helps avoid rewriting the main config file for every node edit.

## Supported Input / Output

### Input

- named remote subscriptions
- inline mixed URI lists
- Base64-encoded subscription content
- manual custom nodes

### Output

- Mihomo YAML

## Development

Run tests:

```sh
go test ./...
```

Run locally:

```sh
go run ./cmd/subconv-next serve --config ./config/config.json
```

Build:

```sh
go build ./cmd/subconv-next
```

## Repository Structure

```text
cmd/subconv-next/        CLI entrypoint
internal/api/            HTTP server and embedded UI endpoints
internal/config/         JSON/UCI config loading and validation
internal/fetcher/        upstream subscription fetch and cache
internal/model/          core data structures
internal/nodestate/      state persistence
internal/parser/         protocol and subscription parsing
internal/pipeline/       refresh and render pipeline
internal/renderer/       Mihomo YAML renderer and validation
testdata/                golden files and parser fixtures
```

## Status

The project is usable today for Docker-based Mihomo subscription conversion and continues to evolve quickly.

Planned and ongoing work includes:

- further UI polish
- broader template compatibility
- OpenWrt-facing packaging improvements
- more protocol and renderer coverage

## Contributing

Issues and pull requests are welcome.

If you are filing a bug report, please include:

- your config shape
- the subscription input type
- the generated error or log output
- whether the issue is in parsing, refresh, node editing, or rendering
