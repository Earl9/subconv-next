# SubConv Next LuCI RPC Infrastructure Implementation

实施日期：2026-07-11
实施范围：第三阶段后端基础设施。未实现正式 Dashboard、Service、Network、Storage、Logs 或 About 页面。

## 1. 新增和变更文件

### 新增

- `openwrt/luci-app-subconv-next/root/usr/libexec/rpcd/luci.subconv`
  - 标准 rpcd executable plugin。
  - 提供固定的服务控制、状态读取和 UCI 配置接口。
- `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/view/subconv-next/dashboard.js`
  - 最小 LuCI JavaScript 入口，仅验证 `luci.subconv.status` 是否可调用。

### 变更

- `openwrt/luci-app-subconv-next/root/usr/share/rpcd/acl.d/luci-app-subconv-next.json`
  - 移除直接 UCI、`service.list` 和 `luci.setInitAction` 权限。
  - 只授权 `luci.subconv` 的指定方法。
- `openwrt/luci-app-subconv-next/root/usr/share/luci/menu.d/luci-app-subconv-next.json`
  - `admin/services/subconv-next` 改为 `firstchild` 父节点。
  - 新增 `admin/services/subconv-next/dashboard` 最小入口。
- `openwrt/luci-app-subconv-next/Makefile`
  - 明确依赖 `rpcd`。
- `openwrt/subconv-next/Makefile`
  - all-in-one 包安装 rpcd、ACL、菜单和最小 JS View。
- `scripts/package-openwrt-ipk-sdk.sh`
  - 打包并验证 rpcd 和 `dashboard.js` 的路径与权限。
- `scripts/package-openwrt-luci-ipk-sdk.sh`
  - 独立 LuCI 包打包并验证 rpcd 和 `dashboard.js`。

### 删除

- `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/view/subconv-next/index.js`
  - 删除直接调用通用 LuCI RPC、直接读写 UCI、使用 `rawhtml` 的旧单文件页面。

## 2. RPC 接口

RPC 对象：

```text
luci.subconv
```

### `status`

无参数。返回：

```json
{
  "running": true,
  "pid": 1234,
  "version": "1.0.7-13",
  "uptime": "1 day 02:03:04",
  "uptime_seconds": 93784
}
```

数据来源：

- PID 优先从 `ubus call service list` 的 procd 实例读取。
- procd 数据不可用时回退到 `pidof subconv-next`。
- 版本来自 `/usr/bin/subconv-next version`。
- uptime 根据 `/proc/<pid>/stat`、`/proc/uptime` 和系统时钟频率计算。
- 服务未运行时返回 `running: false`、`pid: 0` 和零 uptime。

### `start`、`stop`、`restart`

无参数。只允许执行固定命令：

```text
/etc/init.d/subconv-next start
/etc/init.d/subconv-next stop
/etc/init.d/subconv-next restart
```

成功返回：

```json
{
  "success": true,
  "message": "started"
}
```

失败返回 `success: false` 和固定错误信息。接口不接受服务名、动作名或 shell 命令参数。

### `get_config`

无参数。通过 UCI CLI 读取 `subconv-next.main`：

```json
{
  "enabled": true,
  "listen": "0.0.0.0",
  "port": 9876,
  "data_dir": "/etc/subconv-next/data",
  "log_level": "info",
  "public_base_url": ""
}
```

RPC 字段 `listen` 映射现有 UCI 字段 `host`，保持现有配置兼容。

### `set_config`

接受上述字段的部分更新，例如：

```json
{
  "port": 9877
}
```

后端先校验所有传入字段，再执行 `uci set`，最后只执行一次：

```text
uci commit subconv-next
```

校验包括：

- `enabled` 必须为布尔值。
- `port` 必须为 `1-65535` 的整数。
- `listen` 必须为非空且只包含地址或主机名安全字符。
- `data_dir` 必须位于 `/etc/subconv-next`、`/mnt` 或 `/overlay` 的子目录中，且不能包含控制字符、重复分隔符、`.`、`..` 或解析后逃逸到其他目录的符号链接。
- `log_level` 只能是 `debug`、`info`、`warn` 或 `error`。
- `public_base_url` 为空或 HTTP(S) URL。

接口不直接修改 `/etc/config/subconv-next`。

## 3. ACL

ACL 名称：`luci-app-subconv-next`。

只读方法：

```text
luci.subconv.status
luci.subconv.get_config
```

写方法：

```text
luci.subconv.start
luci.subconv.stop
luci.subconv.restart
luci.subconv.set_config
```

未开放：

- 任意 shell 执行。
- 通用 `luci.setInitAction`。
- `service.list`。
- LuCI 直接 UCI 读写权限。
- 任意文件读写接口。

## 4. 测试结果

### 已在开发机完成

- `dash -n`：rpcd 和两套 IPK 打包脚本通过。
- `node --check`：最小 `dashboard.js` 通过。
- `jq`：ACL 和菜单 JSON 解析通过。
- 使用 SDK `jshn` 和 BusyBox ash 模拟 rpcd 协议：
  - `list` 精确包含 `status`、`start`、`stop`、`restart`、`get_config`、`set_config`。
  - `status` 正确读取模拟 procd PID、程序版本和 `/proc` uptime。
  - `get_config` 正确读取并映射 UCI 字段。
  - `set_config` 正确执行部分更新并只提交一次。
  - 非法端口不会执行 `uci set` 或 `uci commit`。
  - 三个服务动作只调用固定 init 脚本。
- all-in-one IPK 构建通过：
  - `/usr/libexec/rpcd/luci.subconv` 为 `0755`。
  - ACL、菜单、`dashboard.js` 为 `0644`。
- 独立 LuCI IPK 构建通过，并验证相同路径与权限。
- `git diff --check` 通过。

### 需要在真实 OpenWrt 上验收

开发机没有运行 OpenWrt 的 `ubus`、rpcd、UCI 和 LuCI 环境，因此以下运行时检查不能在开发机完成：

```sh
ubus -v list luci.subconv
ubus call luci.subconv status '{}'
ubus call luci.subconv get_config '{}'
```

安装后还应验证：

1. 普通 LuCI 会话可以调用 ACL 中授权的方法。
2. 同一会话不能调用未授权的通用 shell、init 或 UCI 接口。
3. `/cgi-bin/luci/admin/services/subconv-next` 跳转到 Dashboard 子入口。
4. 页面显示 `Hello SubConv Next` 和 `RPC connected`。

## 5. 后续 Dashboard 计划

第四阶段再基于当前 RPC 基础实现正式页面：

1. Dashboard：服务状态、版本、监听地址和关键诊断摘要。
2. Service：启动、停止、重启和启用状态管理。
3. Network：监听地址、端口和 Public Base URL。
4. Storage：数据目录、空间和权限诊断。
5. Logs：受限、脱敏、分页的服务日志接口与查看器。
6. About：版本、构建信息和更新状态。

后续页面应继续通过专用 `luci.subconv` RPC 扩展，不恢复前端直接 shell、通用 init RPC 或直接配置文件写入。
