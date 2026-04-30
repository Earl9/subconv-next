# Release Checklist

Use this checklist before tagging a V1 release.

## Automated Checks

- [ ] `gofmt` has no changes.
- [ ] `go test ./...` passes.
- [ ] `go test -race ./...` passes.
- [ ] `go vet ./...` passes.
- [ ] If a frontend package is added, `npm test` passes.
- [ ] If a frontend package is added, `npm run build` passes.
- [ ] `docker compose config` passes.
- [ ] `docker buildx build --platform linux/amd64,linux/arm64 ...` succeeds.

## Core Flow

- [ ] Open the Web UI and confirm a new privacy workspace is created.
- [ ] Add at least two subscription sources with different names and emojis.
- [ ] Confirm node preview names use `{emoji} {sourceName}｜{rawNodeName}`.
- [ ] Confirm raw node names keep upstream tags such as `[anytls]`, `[HY2]`, region, multiplier, IPv6, and line labels.
- [ ] Confirm duplicate final proxy names get `#2`, `#3` suffixes.
- [ ] Confirm proxy-groups reference the final unique names.
- [ ] Confirm no country or region proxy-groups are generated.
- [ ] Confirm `🚀 节点选择` contains `⚡ 自动选择`, `DIRECT`, `REJECT`, and real nodes.
- [ ] Confirm `⚡ 自动选择` contains only real nodes.
- [ ] Confirm AI / YouTube / Netflix / Google rule groups can directly select real nodes in full mode.
- [ ] Confirm rules mode and template mode are mutually exclusive.
- [ ] Add a manual node, run parse preview, and confirm parsed/error status is visible.
- [ ] Disable manual nodes and confirm they do not participate in generation.

## Published Link

- [ ] Click `生成订阅链接`.
- [ ] Confirm `/sub/mihomo.yaml` returns `404`.
- [ ] Confirm `GET /s/{token}/mihomo.yaml` returns `200`.
- [ ] Confirm `HEAD /s/{token}/mihomo.yaml` returns `200`.
- [ ] Confirm `GET /s/{token}/mihomo.yaml` returns `Content-Disposition`, `Profile-Update-Interval`, and `Subscription-Userinfo` when upstream metadata exists.
- [ ] Confirm `HEAD /s/{token}/mihomo.yaml` returns the same metadata headers and an empty body.
- [ ] Confirm no `Subscription-Userinfo` header is returned when upstream metadata is unavailable.
- [ ] Import the generated link into Mihomo / Clash Meta.
- [ ] Confirm Mihomo / Clash Meta shows traffic usage and expiry when upstream metadata exists.
- [ ] Click `重新生成配置`.
- [ ] Confirm `publish_id` and subscription URL do not change.
- [ ] Confirm `/data/published/{publish_id}` contains only `current.yaml` and `meta.json`.
- [ ] Click `重新生成私密链接`.
- [ ] Confirm the old `/s/{old-token}/mihomo.yaml` returns `404`.
- [ ] Confirm the new `/s/{new-token}/mihomo.yaml` returns YAML.
- [ ] Delete the publish and confirm the latest link returns `404`.

## Local Draft

- [ ] In privacy mode, refresh the page and confirm previous config is not auto-restored.
- [ ] Save as local draft.
- [ ] Refresh the page and confirm only `发现本机草稿` is shown.
- [ ] Restore the draft and confirm subscriptions, rules, output options, manual nodes, and node edit state are restored.
- [ ] If the saved `publish_id` still exists, confirm the original subscription link is restored.
- [ ] After restore, click `重新生成配置` and confirm no new `publish_id` is created.
- [ ] Inspect `localStorage.SUBCONV_LOCAL_DRAFT` and confirm it does not contain a full `/s/{token}/mihomo.yaml` URL or raw subscription token.
- [ ] Discard the draft and confirm the localStorage key is removed.

## Security

- [ ] `/api/config` does not expose access token or full subscription URL query secrets.
- [ ] `/api/nodes` and node detail responses mask password, uuid, private key, and pre-shared key values.
- [ ] `/api/logs` does not contain full upstream URLs or published tokens.
- [ ] `/data/logs/app.log` does not contain full upstream URLs or published tokens.
- [ ] `docker logs subconv-next` does not contain full tokens, password, uuid, private-key, pre-shared-key, Authorization, or Cookie values.
- [ ] Deleting a publish makes the old URL return `404`.
- [ ] Rotating a private link makes the old URL return `404`.
- [ ] `localStorage.SUBCONV_LOCAL_DRAFT` does not contain a full subscription token.
- [ ] Published YAML responses include `Cache-Control: no-store`.
- [ ] Published YAML responses do not expose full tokens in logs; only token hints are allowed.
- [ ] Logs rotate and keep at most three 5 MB rotated files.

## YAML Integrity

- [ ] `proxy-groups[].proxies` references only `proxies[].name`, `proxy-groups[].name`, `DIRECT`, or `REJECT`.
- [ ] `⚡ 自动选择` references only real nodes.
- [ ] Disabled, deleted, keyword-excluded, invalid, and info nodes do not appear in final `proxies`.
- [ ] Disabled, deleted, keyword-excluded, invalid, and info nodes do not appear in `proxy-groups`.
- [ ] Rule provider references and rule target groups are valid.
- [ ] `MATCH` is the final rule.

## Docker

- [ ] `docker-compose.yml` persists `/data`.
- [ ] `docker compose config` passes.
- [ ] `docker compose build` succeeds.
- [ ] `docker compose up -d` starts the service.
- [ ] `curl -fsS http://127.0.0.1:9876/healthz` returns `{"ok":true,...}`.
- [ ] `docker compose ps` shows the container as healthy.
- [ ] Stop and recreate the container, then confirm existing published links still work.
- [ ] After `docker compose restart`, a published link still returns `200`.
- [ ] LAN binding is intentional and documented.
- [ ] `127.0.0.1` binding option is documented for local-only deployment.
- [ ] Container healthcheck is healthy.
- [ ] linux/amd64 buildx succeeds.
- [ ] linux/arm64 buildx succeeds.

## OpenWrt

- [ ] `subconv-next serve --host --port --data-dir --public-base-url --log-level` starts successfully on Linux.
- [ ] Embedded Web UI assets load from the binary without external static files.
- [ ] `openwrt/subconv-next/Makefile` package skeleton is present.
- [ ] `openwrt/subconv-next/files/subconv-next.init` uses `procd`.
- [ ] `openwrt/subconv-next/files/subconv-next.config` defaults data to `/etc/subconv-next/data`.
- [ ] No LuCI work is included in V1.

## Documentation

- [ ] README explains project scope, supported protocols, Docker quick start, security model, configuration, development, and roadmap.
- [ ] `SECURITY.md` and `docs/security.md` reflect the current implementation.
- [ ] `docs/docker.md` documents `/data`, healthcheck, header verification, and arm64 build.
- [ ] `docs/openwrt-build.md` documents binary + init.d deployment.
- [ ] OpenWrt docs clearly state V1 supports binary + init.d first, IPK later.
- [ ] Release notes mention breaking changes, security model, and known limitations.
