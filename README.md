# SubConv Next

SubConv Next 是面向 Mihomo / Clash Meta 的现代订阅转换工具，内置轻量 Web UI。它支持多订阅源聚合、节点编辑、规则模板、随机私密订阅链接和 Docker 部署，目标是在新协议支持、现代 UI 和自托管安全模型上补齐传统 subconverter 的不足。

V1 的重点是安全、隐私会话、随机订阅链接、最终 YAML 防泄露校验，以及可控的数据生命周期。

## 特性

- 内置 Web UI：Go 单二进制服务，无需 Node.js、Vite、React 构建链或外部 CDN。
- 多订阅源聚合：支持多个上游订阅、订阅源 Emoji 标识、批量导入、订阅流量/到期信息聚合。
- 节点命名：最终 YAML 默认使用 `Emoji + 订阅源名称 + 原始节点名`，保留上游 raw node name，重复名称追加 `#2`、`#3`。
- 本地节点管理：支持节点编辑、重命名、禁用、删除、恢复、批量改名和手动节点。
- 订阅信息：聚合上游 `subscription-userinfo`，最终订阅接口返回 `Subscription-Userinfo` header。
- 隐私会话：默认打开新 workspace，不自动加载上一个用户配置。
- 本机草稿：只有用户主动点击保存时才写入浏览器 `localStorage`。
- 随机发布链接：订阅默认使用 `/s/{random-token}/mihomo.yaml`。
- 规则配置：支持规则模式 / 模板模式互斥，避免两套规则来源同时生效。
- 安全刷新：普通重新生成只覆盖 `current.yaml`，不会不断创建历史 YAML。
- 私密链接轮换：轮换 token 后旧链接立即失效。
- 输出校验：渲染后 parse YAML back，检查节点泄露、组引用和 `MATCH` 顺序。
- 日志脱敏和轮转：默认屏蔽 token、password、uuid、private-key 等敏感信息。
- Docker 部署：支持 `/data` 持久化、healthcheck、amd64/arm64 multi-arch build。
- OpenWrt planned：V1 先稳定 Docker，后续提供二进制、init.d、UCI、ipk 和 LuCI 路线。

## 不做什么

SubConv Next 专注订阅解析、节点整理和 Mihomo 配置生成，不提供：

- 代理转发或代理核心
- 节点测速
- 连通性探测
- 端口扫描
- 公网认证网关
- 修改远端订阅源内容

## 快速开始

使用发布镜像时，新建 `docker-compose.yml`：

```yaml
services:
  subconv-next:
    image: ghcr.io/OWNER/subconv-next:latest
    container_name: subconv-next
    restart: unless-stopped
    ports:
      - "9876:9876"
    volumes:
      - ./data:/data
    environment:
      - SUBCONV_HOST=0.0.0.0
      - SUBCONV_PORT=9876
      - SUBCONV_DATA_DIR=/data
      - SUBCONV_LOG_LEVEL=info
```

启动：

```sh
docker compose up -d
curl -fsS http://127.0.0.1:9876/healthz
```

从当前源码构建：

```sh
docker compose up -d --build
curl -fsS http://127.0.0.1:9876/healthz
```

常用环境变量：

```sh
SUBCONV_HOST=0.0.0.0
SUBCONV_PORT=9876
SUBCONV_DATA_DIR=/data
SUBCONV_PUBLIC_BASE_URL=https://subconv.example.com
SUBCONV_LOG_LEVEL=info
```

当前仓库的 `docker-compose.yml` 已配置为局域网可访问：

```yaml
ports:
  - "0.0.0.0:9876:9876"
```

打开 Web UI：

```text
http://<你的局域网 IP>:9876/
```

本机访问：

```text
http://127.0.0.1:9876/
```

安全提醒：局域网模式只适合可信内网。不要把 Web UI 直接暴露到公网；如需公网访问，请放到反向代理、认证网关或 VPN 后面。

## 部署模式

局域网访问：

```yaml
ports:
  - "0.0.0.0:9876:9876"
```

仅本机访问：

```yaml
ports:
  - "127.0.0.1:9876:9876"
```

持久化目录：

```yaml
volumes:
  - ./config:/config
  - ./data:/data
```

`./data` 必须保留挂载；发布数据位于 `/data/published/{publish_id}`，容器重启后订阅链接依赖这里的 `meta.json` 和 `current.yaml`。

构建多架构镜像：

```sh
docker buildx build --platform linux/amd64,linux/arm64 -t subconv-next:local .
```

## 使用流程

1. 打开 Web UI，系统会创建一个新的隐私 workspace。
2. 添加上游订阅或手动节点。
3. 为每个订阅源选择名称和 Emoji，用于在 Clash Meta / Mihomo 节点列表中生成 `{emoji} {sourceName}｜{rawNodeName}` 前缀。
4. 按需调整过滤规则、输出选项、规则模式和节点状态。
5. 点击 `生成订阅链接`。
6. 将生成的 `/s/{random-token}/mihomo.yaml` 导入 Clash Meta / Mihomo。
7. 后续点击 `重新生成配置` 会覆盖同一个 `current.yaml`，订阅链接保持不变。
8. 如怀疑链接泄露，点击 `重新生成私密链接`，旧链接会立即返回 `404`。

## 隐私会话和本机草稿

首页默认创建新 workspace，不会自动加载上一个用户的配置。

顶部状态有三种：

- `隐私会话`：默认模式。刷新页面后不会自动保留配置。
- `发现本机草稿`：浏览器里有用户手动保存过的草稿，但不会自动恢复。
- `本机草稿`：当前配置已由用户主动保存到当前浏览器。

本机草稿只会在用户点击 `保存为本机草稿` 或 `更新本机草稿` 后写入：

```text
localStorage.SUBCONV_LOCAL_DRAFT
```

丢弃草稿会删除该 localStorage key。草稿不会保存 published subscription token，也不会保存完整输出 YAML。

## 随机订阅链接

固定路径 `/sub/mihomo.yaml` 已禁用，不能直接下载 YAML。

订阅客户端使用随机发布链接：

```text
http://<host>:9876/s/{random-token}/mihomo.yaml
```

YAML 响应头：

```text
Content-Type: text/yaml; charset=utf-8
Cache-Control: no-store
X-Robots-Tag: noindex, nofollow, noarchive
X-Content-Type-Options: nosniff
```

任何持有 `/s/{token}/mihomo.yaml` 链接的人都能访问该订阅。若怀疑泄露，请在 Web UI 中重新生成私密链接；旧 token 会立即失效。

## Subscription-Userinfo

SubConv Next 拉取上游订阅时会读取 `subscription-userinfo` 响应头，并在最终订阅接口返回聚合后的 `Subscription-Userinfo`：

```text
Subscription-Userinfo: upload=...; download=...; total=...; expire=...
```

Clash / Mihomo 会使用该 header 显示已用流量、总流量和到期时间。多订阅源时，`upload`、`download`、`total` 会求和，`expire` 取最近到期时间。没有任何上游订阅信息时，不会伪造 `0B / 0B`。

## 重新生成配置 vs 重新生成私密链接

`重新生成配置`：

- 使用当前 workspace 配置重新渲染 YAML。
- 覆盖 `/data/published/{publish_id}/current.yaml`。
- `publish_id` 不变。
- token 和订阅 URL 不变。
- 不保留历史 YAML。

`重新生成私密链接`：

- 只轮换 token。
- `current.yaml` 不变。
- 旧 token 立即失效。
- 新订阅 URL 需要重新导入 Clash Meta / Mihomo。

`删除发布`：

- 删除 `/data/published/{publish_id}`。
- 当前订阅链接立即不可访问。

## 数据目录

运行时数据位于 `/data`：

```text
/data/workspaces/{workspace_hash}/config.json
/data/workspaces/{workspace_hash}/state.json

/data/published/{publish_id}/current.yaml
/data/published/{publish_id}/meta.json

/data/cache/
/data/logs/app.log
```

workspace 是网页编辑会话，默认 24 小时无访问后清理。published 是给 Clash Meta / Mihomo 使用的长期订阅发布，每个 `publish_id` 只保留一个 `current.yaml`。

默认生命周期配置：

```json
{
  "workspace_ttl_seconds": 86400,
  "workspace_cleanup_interval_seconds": 3600,
  "published_delete_if_not_accessed_days": 0
}
```

`published_delete_if_not_accessed_days = 0` 表示不自动删除仍有效的 published。

## 安全和脱敏

默认安全策略：

- 默认无登录，只适合本机或可信内网使用。
- 不建议公网裸奔；公网部署必须放在 HTTPS、反向代理、认证网关或 VPN 后面。
- `/sub/mihomo.yaml` 返回 `404`。
- `/s/{token}/mihomo.yaml` 是随机私密链接，持有者即可访问。
- `/api/config` 不返回明文 access token。
- `/api/config` 中订阅 URL query 会脱敏。
- `/api/nodes` 节点详情会屏蔽 password、uuid、private key、pre-shared key 等字段。
- `/api/logs` 和 `/data/logs/app.log` 不记录完整 `/s/{token}/mihomo.yaml`。
- 日志中 `/s/{token}/mihomo.yaml` 会显示为 `/s/<redacted>/mihomo.yaml`。
- `localStorage` 本机草稿不保存完整发布 token。
- 删除发布后旧链接失效；重新生成私密链接后旧链接失效。
- 日志文件最大 5 MB，最多保留 3 个旧文件。

更多安全边界和发版前安全检查见 [SECURITY.md](SECURITY.md) 和 [docs/security.md](docs/security.md)。

## 最终 YAML 完整性校验

渲染链路只把 `finalNodes` 交给 renderer。默认不会进入 YAML 的节点包括：

- 禁用节点
- 删除节点
- 关键词排除节点
- 非法节点
- 信息节点
- 被去重的重复节点

渲染后会重新解析 YAML，并检查：

- `proxies` 中的节点都来自 `finalNodes`
- `proxy-groups` 不引用 excluded nodes
- proxy group 引用只允许已存在的节点名、已存在的组名、`DIRECT`、`REJECT`
- `⚡ 自动选择` 只能引用真实节点，不能引用策略组或信息节点
- 不生成国家/地区策略组
- rule provider 引用都存在
- `MATCH` 规则位于最后

严格模式下，如果校验失败，本次 YAML 写入会被阻止。

## 支持协议

当前解析和渲染覆盖：

- `ss`
- `ssr`
- `vmess`
- `vless`
- `trojan`
- `hysteria2`
- `tuic`
- `anytls` experimental
- `wireguard` experimental

已覆盖的常见传输和安全组合包括：

- Reality
- xHTTP
- gRPC
- WebSocket
- HTTP/2

## HTTP API

workspace 相关 API 需要携带 `?workspace={workspace_id}`。

常用接口：

```text
GET    /healthz
POST   /api/workspaces
DELETE /api/workspaces/{workspace_id}

GET    /api/config?workspace=...
PUT    /api/config?workspace=...
GET    /api/status?workspace=...
POST   /api/refresh?workspace=...

GET    /api/published?workspace=...
POST   /api/published/{publish_id}/rotate-token
DELETE /api/published/{publish_id}

GET    /api/nodes?workspace=...
GET    /api/logs?workspace=...&tail=200

GET    /s/{random-token}/mihomo.yaml
```

最终订阅接口 `GET` 和 `HEAD` 都会返回 Clash / Mihomo 订阅元信息响应头：

```text
Content-Disposition: attachment; filename="mihomo.yaml"
Profile-Update-Interval: 24
Subscription-Userinfo: upload=...; download=...; total=...; expire=...
```

只有上游实际返回 `subscription-userinfo` 时才会透传聚合后的 `Subscription-Userinfo`，不会伪造 `0B / 0B`。

## 开发

需要 Go 1.22+。

运行测试：

```sh
go test ./...
go test -race ./...
go vet ./...
make fmt-check
```

当前前端是内嵌静态页面，没有 `package.json` 或独立 npm 构建链；如后续加入前端测试/构建，请在 Release Gate 中追加：

```sh
npm test
npm run build
```

本地运行：

```sh
go run ./cmd/subconv-next serve --config ./config/config.json
```

常用服务参数：

```sh
subconv-next serve \
  --config ./config/config.json \
  --host 0.0.0.0 \
  --port 9876 \
  --data-dir /data \
  --public-base-url http://router.lan:9876 \
  --log-level info
```

Docker 详细部署说明见 [docs/docker.md](docs/docker.md)。
配置说明见 [docs/configuration.md](docs/configuration.md)，常见问题见 [docs/troubleshooting.md](docs/troubleshooting.md)。

构建二进制：

```sh
go build -o subconv-next ./cmd/subconv-next
```

构建 Docker 镜像：

```sh
docker build -t subconv-next:local .
```

发版前请按 [docs/release-checklist.md](docs/release-checklist.md) 完成人工验收。

## OpenWrt 计划

OpenWrt 支持在 V1 中仍标记为实验性；普通部署优先推荐 Docker。

首个目标为 `rockchip/armv8`，即 `arm64 / aarch64`。默认发布包是 all-in-one IPK，包含后端二进制、`init.d`、UCI、`procd` 数据目录和 LuCI 管理壳。

包格式按设备包管理器选择：

- Kwrt/opkg 设备安装 `.ipk`，例如 `DISTRIB_ARCH='aarch64_generic'` 且 `opkg print-architecture` 包含 `aarch64_generic`。
- 官方 OpenWrt 25.12 SDK 默认启用 `CONFIG_USE_APK=y`，通常输出 `.apk`，只适合 APK-based 固件。
- 如果暂时没有 Kwrt/IPK SDK，可以使用 `scripts/package-openwrt-ipk-sdk.sh` 调用 OpenWrt SDK 自带 `scripts/ipkg-build` 从 arm64 静态二进制生成 opkg 可安装的 `.ipk`。

安装路径：

```text
/usr/bin/subconv-next
/etc/config/subconv-next
/etc/init.d/subconv-next
/etc/subconv-next/data
/usr/share/luci/menu.d/luci-app-subconv-next.json
/usr/share/rpcd/acl.d/luci-app-subconv-next.json
/www/luci-static/resources/view/subconv-next/index.js
```

基本使用：

```sh
# opkg/IPK systems, including current Kwrt 25.12.2 opkg builds:
opkg install /tmp/subconv-next_1.0.0-3_aarch64_generic.ipk

# APK-based OpenWrt 25.12 builds:
apk add --allow-untrusted /tmp/subconv-next-*.apk

curl -fsS http://127.0.0.1:9876/healthz
```

all-in-one 包安装后会按 UCI `enabled=1` 默认自动 enable/start。LuCI 入口位于 `Services / SubConv Next`，可查看状态、修改端口、启动/停止/重启服务，并打开 SubConv Next Web UI；不需要单独安装 `luci-app-subconv-next`。

当前 OpenWrt 相关说明见 [docs/openwrt-build.md](docs/openwrt-build.md)、[docs/03-openwrt-package.md](docs/03-openwrt-package.md) 和 [docs/10-luci-app.md](docs/10-luci-app.md)。

## 仓库结构

```text
cmd/subconv-next/        CLI 入口
internal/api/            HTTP API 和内置 Web UI
internal/config/         JSON/UCI 配置加载和校验
internal/fetcher/        上游订阅抓取
internal/model/          核心数据结构
internal/nodestate/      节点状态持久化
internal/parser/         协议和订阅解析
internal/pipeline/       过滤、最终节点和渲染流水线
internal/renderer/       Mihomo YAML 渲染
testdata/                测试 fixture 和 golden 文件
```

## 路线图

V1 后续方向：

- OpenWrt / LuCI 包装
- 更完整的协议兼容 fixture
- 更细粒度的诊断页
- 更多外部模板兼容性
- 更完善的发布镜像和版本签名

## 贡献

欢迎提交 issue 和 pull request。建议在 issue 中提供：

- 输入订阅类型或最小复现样例
- 预期输出和实际输出
- 相关日志，注意先确认日志中没有敏感 token 或订阅 URL
- 使用的部署方式和版本

提交 PR 前请运行：

```sh
go test ./...
go test -race ./...
go vet ./...
make fmt-check
```

## License

SubConv Next is released under the MIT License. See [LICENSE](LICENSE).
