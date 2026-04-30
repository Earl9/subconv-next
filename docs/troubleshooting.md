# Troubleshooting

## Clash / Mihomo Does Not Show Traffic or Expiry

Check the final subscription endpoint, not only the generation API:

```sh
curl -D - -o /dev/null "http://host:9876/s/{token}/mihomo.yaml"
```

Expected headers when upstream metadata exists:

```text
Subscription-Userinfo: upload=...; download=...; total=...; expire=...
Profile-Update-Interval: 24
Content-Disposition: attachment; filename="mihomo.yaml"
```

If `Subscription-Userinfo` is missing, check whether the upstream provider returns `subscription-userinfo` and whether the published `meta.json` contains `subscription_userinfo`.

## `curl -I` Returns 405

V1 supports `HEAD /s/{token}/mihomo.yaml`. If `curl -I` returns `405`, the running service is an older build or the request is hitting a reverse proxy route that does not forward HEAD correctly.

Upgrade the container or binary, then retry:

```sh
curl -I "http://host:9876/s/{token}/mihomo.yaml"
```

## Restoring a Draft Creates a New Link

Expected behavior: restoring a local draft should rebind to the existing `publish_id` when that publish still exists, and `重新生成配置` should not create a new link.

Check:

- The saved draft contains `publish_ref.publish_id`.
- `/data/published/{publish_id}/meta.json` still exists.
- `/data/published/{publish_id}/current.yaml` still exists.
- The publish was not deleted or token-rotated in a way that the draft no longer references.

If the publish directory is gone, the restored draft cannot reuse the old link and a new publish must be created.

## Container Restart Makes Links Invalid

Published links require persistent `/data`.

Check compose volume mapping:

```yaml
volumes:
  - ./data:/data
```

Then verify files exist:

```sh
ls -la ./data/published
find ./data/published -maxdepth 2 -type f
```

After restart, the same link should still work:

```sh
docker compose restart subconv-next
curl -I "http://127.0.0.1:9876/s/{token}/mihomo.yaml"
```

## YAML Import Fails

First check the API or server logs for `ValidateFinalConfig` errors. The final YAML validator blocks known bad output before publishing.

Common validation failures:

- `proxy-groups[].proxies` references a missing proxy or missing group.
- `⚡ 自动选择` references a group, `DIRECT`, or `REJECT` instead of real nodes only.
- A disabled, deleted, excluded, or invalid node appears in final `proxies`.
- A country or region group appears in V1 output.
- `MATCH` is not the final rule.

If the generated link returns YAML but the client rejects it, save the YAML and inspect `proxies`, `proxy-groups`, and `rules` first.

## Upstream Subscription Has No Traffic Metadata

Some providers only return `subscription-userinfo` for Clash-compatible User-Agent values.

SubConv Next defaults upstream fetches to:

```text
User-Agent: clash.meta
```

If a source config overrides User-Agent, try removing the override or setting a Clash/Mihomo-compatible value. Then regenerate the published subscription and check:

```sh
curl -D - -o /dev/null "http://host:9876/s/{token}/mihomo.yaml"
```

If the upstream provider never returns `subscription-userinfo`, SubConv Next will not fake `0B / 0B`.

## Docker Health Check Is Unhealthy

Check the service:

```sh
docker compose ps
docker logs subconv-next
curl -fsS http://127.0.0.1:9876/healthz
```

Expected shape:

```json
{"ok":true,"version":"...","data_dir":"/data","uptime_seconds":1}
```

If `/healthz` works from inside the container but not from the host, check port binding and firewall rules.
