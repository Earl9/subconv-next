# Release Checklist

Use this checklist before tagging a V1 release.

## Automated Checks

- [ ] `gofmt` has no changes.
- [ ] `go test ./...` passes.
- [ ] `go test -race ./...` passes.
- [ ] `go vet ./...` passes.
- [ ] `docker compose up -d --build` starts successfully.
- [ ] `curl -fsS http://127.0.0.1:9876/healthz` returns `{"ok":true,...}`.
- [ ] `docker buildx build --platform linux/amd64,linux/arm64 ...` succeeds.

## Core Flow

- [ ] Open the Web UI and confirm a new privacy workspace is created.
- [ ] Add at least two subscription sources with different names and emojis.
- [ ] Confirm node preview names use `{emoji} {sourceName}｜{rawNodeName}`.
- [ ] Confirm raw node names keep upstream tags such as `[anytls]`, `[HY2]`, region, multiplier, IPv6, and line labels.
- [ ] Confirm duplicate final proxy names get `#2`, `#3` suffixes.
- [ ] Confirm proxy-groups reference the final unique names.
- [ ] Confirm rules mode and template mode are mutually exclusive.
- [ ] Add a manual node, run parse preview, and confirm parsed/error status is visible.
- [ ] Disable manual nodes and confirm they do not participate in generation.

## Published Link

- [ ] Click `生成订阅链接`.
- [ ] Confirm `/sub/mihomo.yaml` returns `404`.
- [ ] Confirm `/s/{token}/mihomo.yaml` returns YAML.
- [ ] Import the generated link into Mihomo / Clash Meta.
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
- [ ] Published YAML responses include `Cache-Control: no-store`.
- [ ] Logs rotate and keep at most three 5 MB rotated files.

## Docker

- [ ] `docker-compose.yml` persists `/data`.
- [ ] LAN binding is intentional and documented.
- [ ] `127.0.0.1` binding option is documented for local-only deployment.
- [ ] Container healthcheck is healthy.
- [ ] amd64 image runs.
- [ ] arm64 image runs.

## Documentation

- [ ] README explains project scope, supported protocols, Docker quick start, security model, configuration, development, and roadmap.
- [ ] `docs/security.md` reflects the current implementation.
- [ ] OpenWrt docs clearly state V1 supports binary + init.d first, IPK later.
- [ ] Release notes mention breaking changes, security model, and known limitations.
