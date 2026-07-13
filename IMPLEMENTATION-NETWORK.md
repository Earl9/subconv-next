# SubConv Next Network Page Implementation

实施日期：2026-07-11
实施范围：第六阶段 Network 页面。未实现 Storage、Logs 或 About 页面。

## 1. 修改文件

### 新增

- `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/view/subconv-next/network.js`
  - Network 页面入口和保存流程。
- `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/subconv-next/network-validation.js`
  - 无 RPC、无 DOM 依赖的网络字段校验模块。
- `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/subconv-next/network.css`
  - Network 表单、错误状态、重启提示和移动布局。

### 变更

- `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/subconv-next/api.js`
  - 整理为 `service`、`config`、`system` 命名空间。
- `dashboard.js` 和 `service.js`
  - 同步使用新的共享 API 命名空间。
- `openwrt/luci-app-subconv-next/root/usr/share/luci/menu.d/luci-app-subconv-next.json`
  - 新增 `admin/services/subconv-next/network`。
- `openwrt/subconv-next/Makefile`
- `openwrt/luci-app-subconv-next/Makefile`
- `scripts/package-openwrt-ipk-sdk.sh`
- `scripts/package-openwrt-luci-ipk-sdk.sh`
  - 安装并验证 Network View、校验模块和 CSS。
  - OpenWrt package release 提升到 `1.0.7-16`。
- `IMPLEMENTATION-SERVICE.md`
  - 更新共享 API 的当前命名空间说明。

## 2. 配置字段

Network 页面只读取和写入第三阶段已有配置接口：

```text
api.config.get()
api.config.save(values)
```

没有新增 rpcd 方法，也没有在页面中直接调用 UCI。

字段映射：

| 页面字段 | RPC 字段 | UCI 字段 |
| --- | --- | --- |
| Listen Address | `listen` | `subconv-next.main.host` |
| Port | `port` | `subconv-next.main.port` |
| Public Base URL | `public_base_url` | `subconv-next.main.public_base_url` |

共享 API 当前结构：

```text
api.service.status/start/stop/restart
api.config.get/save
api.system.getAutostart/setAutostart
api.settle
```

所有 `rpc.declare()` 仍只存在于 `subconv-next/api.js`。

## 3. 校验规则

### Listen Address

支持：

- `0.0.0.0`、`127.0.0.1` 等合法 IPv4。
- 合法 IPv6，例如 `::1`、`2001:db8::1`。
- 合法主机名，例如 `router.local`。

拒绝：

- 空值。
- IPv4 段超过 `255`。
- 残缺 IPv4。
- 含空格的地址。
- 空标签、非法首尾连字符或过长标签的主机名。

### Port

- 必须为十进制整数。
- 范围 `1-65535`。
- 拒绝空值、`0`、`65536`、负数、小数和非数字。

### Public Base URL

- 允许空值。
- 非空时必须是绝对 HTTP 或 HTTPS URL。
- 拒绝相对地址、FTP、JavaScript URL 和包含非法空格的 URL。

前端校验失败时：

- 不调用 `api.config.save()`。
- 在对应字段下显示错误。
- 设置 `aria-invalid="true"`。
- 显示 `Save failed` 通知。

rpcd 的 `set_config` 仍执行后端二次校验，前端校验不能绕过后端安全边界。

## 4. 保存流程

```text
用户提交
  -> 前端校验
  -> api.config.save()
  -> luci.subconv set_config
  -> uci set
  -> uci commit subconv-next
  -> Configuration saved
```

保存期间禁用三个输入框和 Save 按钮。失败时恢复控件并显示 RPC 错误。

保存不会自动重启服务。

- 端口相对页面加载值发生变化时显示：`Port changed. Restart service required.`
- 监听地址相对页面加载值发生变化时显示对应的重启提示。
- Public Base URL 修改不要求重启。

重启提示在当前页面内持续存在；将值改回页面加载时的运行值并保存后提示消失。

## 5. 测试结果

### 前端校验

`network-validation.js` 作为独立模块执行测试，验证：

- 合法 IPv4、IPv6和主机名通过。
- `999.1.1.1`、残缺 IPv4、空格和非法主机名被拒绝。
- `1`、`9876`、`65535` 端口通过。
- 空值、`0`、`65536`、负数、小数和文本端口被拒绝。
- 空 URL、HTTP 和 HTTPS URL 通过。
- 相对地址、FTP、JavaScript URL 和非法空格 URL 被拒绝。

### RPC 保存链路

使用状态化 UCI 桩和真实 rpcd 脚本验证：

1. 调用 `set_config` 保存：

   ```json
   {
     "listen": "10.0.0.5",
     "port": 19876,
     "public_base_url": "https://subconv.example.test"
   }
   ```

2. 随后调用 `get_config` 返回相同的新值。
3. 端口 `70000` 被后端拒绝。
4. 非法保存不会覆盖之前的端口。
5. 合法保存只执行一次 `uci commit subconv-next`。

### 静态和回归测试

- API、Dashboard、Service、Network 和校验模块通过 `node --check`。
- View 中没有 `rpc.declare`、直接 UCI、`innerHTML` 或 `rawhtml`。
- 打包脚本通过 shell 语法检查。
- 菜单 JSON 通过 `jq` 解析。
- `git diff --check` 通过。
- `go test ./...` 通过。

### 构建结果

```text
dist/subconv-next_1.0.7-16_aarch64_generic.ipk
dist/luci-app-subconv-next_1.0.7-16_all.ipk
```

`make package` 和独立 LuCI 包构建均通过。两个包均包含：

```text
/www/luci-static/resources/subconv-next/api.js
/www/luci-static/resources/subconv-next/network-validation.js
/www/luci-static/resources/subconv-next/network.css
/www/luci-static/resources/view/subconv-next/network.js
```

以上 LuCI 文件模式为 `0644`。

### 真实 OpenWrt 验收

开发机没有 `opkg`、`ubus`、rpcd、LuCI 或浏览器环境，因此设备端仍需执行：

```sh
opkg install /tmp/subconv-next_1.0.7-16_aarch64_generic.ipk
ubus call luci.subconv get_config '{}'
```

浏览器访问：

```text
/cgi-bin/luci/admin/services/subconv-next/network
```

验证非法字段无法提交、合法保存成功、重启提示和移动端布局。

## 6. 已知问题

- 本阶段按要求不实现端口占用检测，因此合法端口格式不代表端口当前可绑定。
- 保存后不自动重启。页面只提示用户需要重启服务，不会改变当前运行进程。
- 重启提示以页面首次加载的配置作为当前运行基线。如果服务在其他会话中重启或配置被外部修改，应重新加载页面获取新基线。
- 页面尚未在真实 OpenWrt 主题、平板和手机浏览器中截图验收。
