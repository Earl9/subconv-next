# SubConv Next

SubConv Next 是一个 Docker-first 的订阅转换器和 Mihomo YAML 生成器，内置轻量 Web UI。它面向需要聚合多个上游订阅、管理本地节点编辑、并生成 Clash Meta / Mihomo 订阅链接的自托管用户。

V1 的重点是安全、隐私会话、随机订阅链接、最终 YAML 防泄露校验，以及可控的数据生命周期。

## 特性

- 内置 Web UI：Go 单二进制服务，无需 Node.js、Vite、React 构建链或外部 CDN。
- 多订阅源聚合：支持多个命名上游订阅、批量导入、订阅流量/到期信息聚合。
- 本地节点管理：支持重命名、禁用、删除、恢复、批量改名和手动节点。
- 隐私会话：默认打开新 workspace，不自动加载上一个用户配置。
- 本机草稿：只有用户主动点击保存时才写入浏览器 `localStorage`。
- 随机发布链接：订阅默认使用 `/s/{random-token}/mihomo.yaml`。
- 安全刷新：普通重新生成只覆盖 `current.yaml`，不会不断创建历史 YAML。
- 私密链接轮换：轮换 token 后旧链接立即失效。
- 输出校验：渲染后 parse YAML back，检查节点泄露、组引用和 `MATCH` 顺序。
- 日志脱敏和轮转：默认屏蔽 token、password、uuid、private-key 等敏感信息。
- Docker 部署：支持 `/data` 持久化、healthcheck、amd64/arm64 multi-arch build。

## 不做什么

SubConv Next 专注订阅解析、节点整理和 Mihomo 配置生成，不提供：

- 代理转发或代理核心
- 节点测速
- 连通性探测
- 端口扫描
- 公网认证网关
- 修改远端订阅源内容

## 快速开始

```sh
docker compose up -d --build
curl -fsS http://127.0.0.1:9876/healthz
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

构建多架构镜像：

```sh
docker buildx build --platform linux/amd64,linux/arm64 -t subconv-next:local .
```

## 使用流程

1. 打开 Web UI，系统会创建一个新的隐私 workspace。
2. 添加上游订阅或手动节点。
3. 按需调整过滤规则、输出选项、规则模式和节点状态。
4. 点击 `生成订阅链接`。
5. 将生成的 `/s/{random-token}/mihomo.yaml` 导入 Clash Meta / Mihomo。
6. 后续点击 `重新生成配置` 会覆盖同一个 `current.yaml`，订阅链接保持不变。
7. 如怀疑链接泄露，点击 `重新生成私密链接`，旧链接会立即返回 `404`。

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

- `/sub/mihomo.yaml` 返回 `404`。
- `/api/config` 不返回明文 access token。
- `/api/config` 中订阅 URL query 会脱敏。
- `/api/nodes` 节点详情会屏蔽 password、uuid、private key、pre-shared key 等字段。
- `/api/logs` 和 `/data/logs/app.log` 不记录完整 `/s/{token}/mihomo.yaml`。
- 日志中 `/s/{token}/mihomo.yaml` 会显示为 `/s/<redacted>/mihomo.yaml`。
- 日志文件最大 5 MB，最多保留 3 个旧文件。

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
- proxy group 引用的节点或组都存在
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
- `anytls`
- `wireguard`

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

## 开发

需要 Go 1.22+。

运行测试：

```sh
go test ./...
go test -race ./...
go vet ./...
make fmt-check
```

本地运行：

```sh
go run ./cmd/subconv-next serve --config ./config/config.json
```

构建二进制：

```sh
go build -o subconv-next ./cmd/subconv-next
```

构建 Docker 镜像：

```sh
docker build -t subconv-next:local .
```

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
go vet ./...
make fmt-check
```

## License

当前仓库尚未包含 `LICENSE` 文件。发布到 GitHub 前请明确选择许可证，例如 MIT、Apache-2.0 或 GPL-3.0。
