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

## Deleted or Restored Nodes Are Not Reflected in a Subscription

Deleting or restoring nodes updates an existing publication immediately using cached subscriptions and rules. This includes restoring one node or all deleted nodes. The original link is retained, and no upstream refresh is needed. Refresh the subscription in your proxy client to download the updated configuration.

If the cached configuration cannot be rebuilt or written, the operation is rolled back and the API reports the failure. Deleting the last available node makes the existing publication unavailable; restoring a usable node makes the same link available immediately.

After a service restart, status reads the node count from the existing published YAML. A failed refresh reports its error while retaining the count and timestamp of the last downloadable output.

## Custom Rules Cannot Be Saved

Saving a custom rule checks required fields, URL and interval, format compatibility, and the syntax of every inline entry. Errors identify the field or original input line; the editor keeps the input so it can be corrected. The same validation runs when saving configuration through the API.

Direct rules also have a `校验规则` button and are checked before generation. They require a target policy, such as `DOMAIN-SUFFIX,example.com,DIRECT`. Stored invalid rules remain readable for correction, while saving and generation reject them and retain the previous published output.

Use domain patterns for `domain`, CIDR addresses such as `192.0.2.0/24` for `ipcidr`, and typed entries such as `DOMAIN-SUFFIX,example.com` for `classical`. A classical rule provider cannot contain `MATCH`, `RULE-SET`, or `SUB-RULE`. Inline rules use text or YAML, not MRS.

Remote text and YAML rules are checked before updating the cache and after cache lookup, before publication. An invalid upstream update cannot overwrite a valid cached snapshot. URL validation at save time does not fetch the remote content. MRS and local file contents are loaded by the proxy client.

## `YAML subscription contains zero proxies`

The upstream response was valid YAML with no top-level proxy entries. Some providers return an empty Clash/Mihomo configuration or a template that references remote `proxy-providers` for a client User-Agent while still providing nodes in their generic subscription format.

When the response has an empty or null top-level `proxies` field, or has a nonempty `proxy-providers` map without top-level `proxies`, SubConv Next retries once with `SubConvNext/0.1 OpenWrt`, including when the initial request uses the default `clash.meta`. Refresh the affected source after upgrading to replace its cached empty response. If it still has no nodes, check the provider's subscription status and whether its URL selects a specific client format.

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
