# SubConv Next

[English](../README.md) | [简体中文](README.zh-CN.md)

SubConv Next 是面向 Mihomo / Clash Meta 的自托管订阅转换工具。单个 Go 二进制提供转换 API 和 Web UI，可聚合上游订阅、编辑节点、管理分流规则、生成经过校验的 Mihomo YAML，并发布随机私密订阅链接。

Docker 是推荐的部署方式。项目也包含基于 procd、UCI、rpcd、ACL 和 LuCI JavaScript View 的原生 OpenWrt 集成。

## 功能特性

- 聚合多个订阅源，使用稳定来源 ID、名称和可选 Emoji 前缀。
- 解析 Base64 订阅、Clash/Mihomo YAML 和单节点 URI。
- 支持节点编辑、禁用、删除、恢复、批量改名和手动节点。
- 支持内置规则预设、自定义规则、远程规则集和自定义策略组。
- 可选生成国家组和国家自动测速组。
- 聚合上游 `Subscription-Userinfo`，展示流量用量和到期时间。
- 发布随机 `/s/{token}/mihomo.yaml` 私密订阅链接。
- 支持浏览器本机草稿，不保存完整发布 token。
- 输出前校验节点引用、规则集、策略组、过滤节点和最终 `MATCH` 顺序。
- Docker 支持 `linux/amd64`、`linux/arm64` 以及 `/config`、`/data` 持久化。
- 通过 LuCI 管理 OpenWrt 服务状态、设置、备份恢复和日志。

支持协议包括 `ss`、`ssr`、`vmess`、`vless`、`trojan`、`hysteria2`、`tuic`、`anytls`、`wireguard` 和 `mieru`。

## 截图

### Web UI

![SubConv Next Web UI](assets/screenshot-main.png)

<details>
<summary>手机端 Web UI</summary>

![SubConv Next 手机端 Web UI](assets/screenshot-mobile.png)

</details>

### OpenWrt LuCI

![SubConv Next OpenWrt LuCI 概览](assets/screenshot-luci.png)

<details>
<summary>手机端 LuCI</summary>

![SubConv Next 手机端 LuCI 概览](assets/screenshot-luci-mobile.png)

</details>

## 快速开始

### Docker Compose

仓库已提供加固后的 [docker-compose.yml](../docker-compose.yml)：自动初始化 `/data` 的非特权用户权限，`/config` 只读挂载，并默认只在本机回环地址发布 Web 端口。

在仓库根目录启动并检查健康状态：

```sh
mkdir -p config data
export SUBCONV_ACCESS_TOKEN="$(openssl rand -hex 32)"
docker compose up -d
curl -fsS http://127.0.0.1:9876/healthz
```

访问 <http://127.0.0.1:9876/>，使用 `SUBCONV_ACCESS_TOKEN` 登录。请妥善保存生成的令牌，并在更新时复用，不要将其提交到仓库。

此方式需要 Docker Engine 和 Compose v2 插件。只有 Docker Engine 时，可按 [独立 Docker 启动说明](docker.md#standalone-docker) 部署。局域网访问时，在启动前设置 `SUBCONV_HOST_BIND=0.0.0.0`，并保留高强度 Access Token。本地镜像构建、更新、持久化、备份和反向代理说明见 [Docker 部署](docker.md)。

使用 `ghcr.io/earl9/subconv-next:latest` 跟随发布版本；需要可重复部署时请固定具体版本标签。

### OpenWrt / Kwrt

安装完整包即可同时获得后台服务和 LuCI 页面：

```sh
opkg install /tmp/subconv-next_<版本>-<发行号>_aarch64_generic.ipk

ubus call luci.subconv status
curl -fsS http://127.0.0.1:9876/healthz
```

`aarch64_generic` 已在 Kwrt 25.12.2 `rockchip/armv8` 上验证。其它设备请先检查 `opkg print-architecture`。

官方 OpenWrt 25.12 的 APK 固件请安装对应的完整 APK 包：`apk add --allow-untrusted /tmp/subconv-next-<版本>-r<发行号>.apk`。使用 opkg 的 Kwrt 设备不能安装 APK。

主要安装路径：

```text
/usr/bin/subconv-next
/etc/init.d/subconv-next
/etc/config/subconv-next
/etc/subconv-next/data
/usr/libexec/rpcd/luci.subconv
/usr/share/luci/menu.d/luci-app-subconv-next.json
/usr/share/rpcd/acl.d/luci-app-subconv-next.json
/www/luci-static/resources/view/subconv-next/overview.js
```

当 UCI `enabled=1` 时，安装后会启用并启动服务。LuCI 入口为 **服务 > SubConv Next**。

## 下载

发布文件通过 [GitHub Releases](https://github.com/Earl9/subconv-next/releases) 提供，不提交到仓库。发布资产包括 Linux 二进制、包含后台和 LuCI 页面的 OpenWrt IPK 与 APK 完整包，以及 `checksums.txt`。

## 安全说明

Compose 默认只把端口绑定到 `127.0.0.1`，`/config` 以只读方式挂载，并启用只读根文件系统与 capability drop。需要公开匿名转换时设置 `SUBCONV_PUBLIC_CONVERTER=true` 并保留高强度 `SUBCONV_ACCESS_TOKEN`：访客无需登录，每个访客使用独立的高强度随机工作区，转换接口按来源限速，未列入公开转换白名单的路由仍由管理令牌保护。公开模式会限制空闲工作区最多 6 小时，并清理连续 30 天未访问的发布链接。公网入口仍应由反向代理终止 HTTPS、实施分布式限流，并避免直接暴露 9876 端口。不要在生产环境使用 `SUBCONV_ALLOW_INSECURE_PUBLIC=true`，它会关闭整个管理边界。LuCI 继续使用路由器现有登录认证和 rpcd ACL 权限模型。

发布订阅 URL 是 bearer link。任何持有有效 `/s/{token}/mihomo.yaml` URL 的人都可以获取生成配置。链接泄露后请在 Web UI 中轮换。

日志和 API 会脱敏已知敏感字段，包括完整 token、上游 URL 凭据、密码、UUID、私钥、预共享密钥、`Authorization` 和 `Cookie`。安全模型和漏洞报告方式见 [SECURITY.md](../SECURITY.md)。

## 文档

- [Docker 部署](docker.md)
- [配置说明](configuration.md)
- [故障排查](troubleshooting.md)
- [发版检查清单](release-checklist.md)
- [OpenWrt 构建和打包](openwrt-build.md)
- [安全模型细节](security.md)
- [OpenWrt 包说明](03-openwrt-package.md)
- [LuCI 应用说明](10-luci-app.md)

## 开发

需要 Go 1.22 或更高版本。

```sh
go test ./...
go test -race ./...
go vet ./...
```

本地运行：

```sh
go run ./cmd/subconv-next serve \
  --host 127.0.0.1 \
  --port 9876 \
  --data-dir "$PWD/data" \
  --log-level info
```

构建本地二进制：

```sh
go build -o subconv-next ./cmd/subconv-next
```

## 许可证

SubConv Next 使用 [MIT License](../LICENSE) 发布。
