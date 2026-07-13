# SubConv Next Logs Page Implementation

实施日期：2026-07-11
实施范围：第八阶段 Logs 页面。未实现 About 页面，也未实现全量日志、`tail -f`、WebSocket 或任意文件读取。

## 1. 修改文件

### 新增

- `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/view/subconv-next/logs.js`
  - Logs 页面、日志源选择、手动刷新、自动刷新、清空和下载交互。
- `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/subconv-next/logs.css`
  - 工具栏、开关、固定高度日志面板和移动端样式。

### 变更

- `openwrt/luci-app-subconv-next/root/usr/libexec/rpcd/luci.subconv`
  - 新增 `logs_read`、`logs_download` 和 `logs_clear`。
- `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/subconv-next/api.js`
  - 新增 `api.logs.read()`、`api.logs.clear()` 和 `api.logs.download()`。
- `openwrt/luci-app-subconv-next/root/usr/share/rpcd/acl.d/luci-app-subconv-next.json`
  - read 增加 `logs_read`、`logs_download`；write 增加 `logs_clear`。
- `openwrt/luci-app-subconv-next/root/usr/share/luci/menu.d/luci-app-subconv-next.json`
  - 新增 `admin/services/subconv-next/logs`。
- `openwrt/subconv-next/Makefile`
- `openwrt/luci-app-subconv-next/Makefile`
- `scripts/package-openwrt-ipk-sdk.sh`
- `scripts/package-openwrt-luci-ipk-sdk.sh`
  - 安装并验证 Logs View 和 CSS。
  - OpenWrt package release 后续已由紧凑界面与备份恢复阶段提升到 `1.0.7-22`。

## 2. RPC 接口

### `logs_read`

参数：

```json
{
  "source": "service",
  "lines": 100
}
```

`source` 只能是：

```text
service
application
```

返回示例：

```json
{
  "success": true,
  "source": "service",
  "lines": ["service started", "listening on 0.0.0.0:9876"],
  "returned_lines": 2,
  "truncated": false
}
```

默认返回 100 行，允许范围为 1 到 1000 行，单次响应原始内容上限为 256 KiB。服务日志使用固定 `/sbin/logread`，并设置 3 秒执行超时。

### `logs_download`

参数：

```json
{
  "source": "application"
}
```

下载数据仍通过 JSON RPC 返回受限、脱敏的行数组。最多读取 1000 行和 512 KiB，浏览器端使用 Blob 生成时间戳文件名，不开放服务器任意文件下载路径。

### `logs_clear`

无参数。只处理配置 `data_dir` 下的固定文件：

```text
logs/app.log
logs/app.log.1
logs/app.log.2
logs/app.log.3
```

当前日志文件被截断，受控轮转文件被删除。任何目标为符号链接时操作会被拒绝。该接口不会清理 OpenWrt `logd` 系统日志。

页面只通过共享 API 调用：

```text
api.logs.read(source, lines)
api.logs.clear()
api.logs.download(source)
```

Logs View 中没有直接声明 `rpc.declare()`。

## 3. 安全限制

- 日志源为固定枚举，不接受文件路径或 shell 片段。
- `lines` 必须是整数，范围为 1 到 1000。
- 不支持全文件读取、`tail -f`、长连接或 WebSocket。
- 服务日志读取最多执行 3 秒。
- 页面读取上限为 256 KiB，下载上限为 512 KiB。
- 单行最多输出 8192 字节，超出时标记结果为 truncated。
- 应用日志目录只能来自经过 Storage 阶段安全校验的 UCI `data_dir`。
- 应用日志目录和日志文件拒绝符号链接，防止路径逃逸。
- token、password、cookie、authorization、private key、代理 URI、URI 凭据、UUID 和发布 `/s/` 路径所在行会整体替换为 `[redacted sensitive log line]`。
- 临时文件使用固定 `/tmp/subconv-next-logs-*.$$` 名称，并在成功和失败路径清理。
- ACL 只开放三个明确方法，没有开放 UCI、任意 shell 或文件访问权限。

## 4. 页面说明

入口：

```text
/cgi-bin/luci/admin/services/subconv-next/logs
```

页面提供：

- Service/Application 日志源选择。
- 100、200、500、1000 行选择。
- Refresh 手动刷新。
- 默认关闭、30 秒周期的 Auto Refresh。
- 页面隐藏时暂停自动请求，离开页面时销毁计时器。
- Service 模式下 Clear 只清空浏览器显示，不修改系统日志。
- Application 模式下 Clear 需要确认并调用 `logs_clear`。
- Download 生成本地文本文件；截断结果在文件首行加入提示。

日志区域使用等宽字体、固定高度、滚动和自动换行。动态日志通过 `textContent` 渲染，没有使用 `innerHTML`、`rawhtml`、Vue、React 或 jQuery。

## 5. 测试结果

### RPC 与安全测试

使用 SDK `jshn`、隔离 `IPKG_INSTROOT`、模拟 `logread` 和 UCI 验证：

1. 正常服务日志和空应用日志返回正确。
2. `app.log.3` 到 `app.log` 按时间顺序合并，再截取最后 N 行。
3. `lines=1001`、非整数行数和非法 source 被拒绝。
4. `invalid;rm -rf /` 仅返回 `invalid log source`，不会执行命令。
5. 敏感日志行被整体脱敏。
6. 读取超过 256 KiB、下载超过 512 KiB 时设置 `truncated=true`。
7. Application Clear 截断当前日志并删除固定轮转文件。
8. 符号链接日志文件被拒绝。
9. RPC 结束后没有残留 `/tmp/subconv-next-logs-*` 临时文件。
10. 阻塞的服务日志读取在 3 秒后失败返回。

### 静态和回归测试

- rpcd 通过 `dash -n`。
- API 和 Logs View 通过 `node --check`。
- ACL 和菜单 JSON 通过 `jq` 解析。
- `git diff --check` 通过。
- `go test ./...` 通过。

### 构建结果

```text
dist/subconv-next_1.0.7-22_aarch64_generic.ipk
dist/luci-app-subconv-next_1.0.7-22_all.ipk
```

`make package` 和独立 LuCI 包构建均通过。两个包均包含：

```text
/usr/libexec/rpcd/luci.subconv
/www/luci-static/resources/subconv-next/api.js
/www/luci-static/resources/subconv-next/logs.css
/www/luci-static/resources/view/subconv-next/logs.js
```

解包后关键文件与源码逐字节一致。rpcd 模式为 `0755`，LuCI、菜单和 ACL 资源为 `0644`。

### 真实 OpenWrt 验收

开发机没有 `opkg`、`ubus`、rpcd、LuCI 或浏览器环境，因此无法完成设备端调用和页面截图。设备端应执行：

```sh
opkg install /tmp/subconv-next_1.0.7-22_aarch64_generic.ipk
opkg install /tmp/luci-app-subconv-next_1.0.7-22_all.ipk
ubus call luci.subconv logs_read '{"source":"service","lines":100}'
```

浏览器访问：

```text
/cgi-bin/luci/admin/services/subconv-next/logs
```

验证手动刷新、30 秒自动刷新、应用日志确认清空和文本下载。

## 6. 已知问题

- Service 日志依赖设备上的 BusyBox `timeout` 和 `/sbin/logread` 参数兼容性，需要真实 OpenWrt 验收。
- Application 日志当前只识别固定 `app.log` 及 `.1` 到 `.3`，不会扫描其他文件。
- 下载由浏览器基于 RPC 数据生成，不是流式下载；因此仍受 1000 行和 512 KiB 限制。
- 敏感信息策略以整行隐藏为主，安全优先但可能遮蔽包含 UUID 或 `/s/` 路径的普通诊断行。
- RPC 尚未全面迁移到统一 `code` 错误模型；本阶段保持现有 `{ success, message }` 契约，避免扩大修改范围。
- 页面尚未在真实 OpenWrt 主题和移动浏览器中截图验收。
