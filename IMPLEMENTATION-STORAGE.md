# SubConv Next Storage Page Implementation

实施日期：2026-07-11
实施范围：第七阶段 Storage 页面。未实现 Logs 或 About 页面，也未实现文件管理功能。

## 1. 修改文件

### 新增

- `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/view/subconv-next/storage.js`
  - Storage 页面和 Check Directory 交互。
- `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/subconv-next/storage.css`
  - 数据目录状态、诊断网格和响应式样式。

### 变更

- `openwrt/luci-app-subconv-next/root/usr/libexec/rpcd/luci.subconv`
  - 新增无参数 `storage_check`。
  - 收紧 `set_config.data_dir` 的后端路径校验。
- `openwrt/luci-app-subconv-next/htdocs/luci-static/resources/subconv-next/api.js`
  - 新增 `api.storage.check()`。
- `openwrt/luci-app-subconv-next/root/usr/share/rpcd/acl.d/luci-app-subconv-next.json`
  - read 增加 `storage_check`。
- `openwrt/luci-app-subconv-next/root/usr/share/luci/menu.d/luci-app-subconv-next.json`
  - 新增 `admin/services/subconv-next/storage`。
- `openwrt/subconv-next/Makefile`
- `openwrt/luci-app-subconv-next/Makefile`
- `scripts/package-openwrt-ipk-sdk.sh`
- `scripts/package-openwrt-luci-ipk-sdk.sh`
  - 安装并验证 Storage View 和 CSS。
  - OpenWrt package release 提升到 `1.0.7-17`。
- `IMPLEMENTATION-RPC.md`
  - 更新 `data_dir` 当前安全规则。

## 2. RPC 变化

### `storage_check`

调用：

```text
luci.subconv storage_check
```

该方法不接受路径或其他参数。路径只能从以下 UCI 配置读取：

```text
subconv-next.main.data_dir
```

返回示例：

```json
{
  "path": "/etc/subconv-next/data",
  "safe": true,
  "exists": true,
  "is_directory": true,
  "readable": true,
  "writable": true,
  "permission": "755",
  "owner": "root:root",
  "available_kb": 123456,
  "message": "storage checked"
}
```

检测来源：

- `test -e`：路径存在。
- `test -d`：是目录。
- `test -r`：服务用户可读。
- 在目标目录创建随机 PID 后缀的零字节临时文件并立即删除：服务用户可写。
- `stat`：权限和 owner。
- `df -Pk`：可用空间。

页面通过共享 API 调用：

```text
api.storage.check()
```

没有在 Storage View 中声明 RPC。

## 3. 安全限制

### 无任意路径参数

`storage_check` 的 rpcd 方法签名为空对象。前端和调用方不能提交 `/`、`/etc/passwd`、`/tmp` 或其他待检查路径。

### 允许目录

配置的 `data_dir` 只允许位于以下子目录：

```text
/etc/subconv-next/*
/mnt/*
/overlay/*
```

这保留默认目录和 `/mnt/sda1/subconv-data` 一类挂载盘目录，同时阻止普通系统路径。

### 拒绝规则

- 根目录 `/`。
- `/tmp` 及其他未允许前缀。
- `/etc/passwd` 等系统文件路径。
- 控制字符。
- 重复路径分隔符。
- `.`、`..` 路径段。
- 解析后逃逸到允许目录之外的符号链接。

rpcd 使用 `readlink -f` 检查已有路径或可解析父路径的真实目标。`/mnt/data` 即使文本前缀合法，如果符号链接指向 `/etc`，仍会返回 `safe: false`，`set_config` 也会拒绝保存。

### 无文件管理能力

本阶段没有新增：

- 路径输入框。
- 文件列表或浏览器。
- 上传、下载、删除、移动或复制。
- 任意 shell 命令参数。

写检测只创建零字节临时文件，并在同一次 RPC 中删除。

## 4. 页面说明

入口：

```text
/cgi-bin/luci/admin/services/subconv-next/storage
```

Data Directory 区域显示：

- 配置路径。
- Allowed Path。
- Exists。
- Is Directory。
- Readable。
- Writable。
- Permission。
- Owner。
- Available Space。

综合状态：

- Healthy：路径安全、存在、是目录、可读且可写。
- Attention：路径安全，但目录不存在、类型错误或读写检测失败。
- Error：RPC 不可用或配置路径不在允许范围。

Check Directory 按钮重新调用 `api.storage.check()` 并替换当前状态。页面不自动轮询，也不修改 `data_dir`。

所有动态内容通过 `E()`、`dom.content()` 或 `textContent` 渲染。没有使用 `innerHTML`、`rawhtml`、Vue、React 或 jQuery。

## 5. 测试结果

### RPC 与安全测试

使用 SDK `jshn`、BusyBox ash、状态化 UCI 桩和非 root 用户验证：

1. 正常目录：
   - `safe=true`
   - `exists=true`
   - `is_directory=true`
   - `readable=true`
   - `writable=true`
   - 返回权限和可用空间。
2. 不存在目录：
   - `safe=true`
   - `exists=false`
   - `writable=false`
3. 权限为 `0555` 且普通用户执行：
   - `exists=true`
   - `readable=true`
   - `writable=false`
   - `permission="555"`
4. 危险路径 `/tmp`：
   - `safe=false`
   - 不执行目录检查。
5. `/mnt/escape` 符号链接指向 `/etc`：
   - `safe=false`
   - 不执行目标目录写测试。
6. `set_config` 拒绝 `/`、`/etc/passwd`、`/tmp` 和逃逸符号链接。
7. `set_config` 接受 `/mnt/sda1/subconv-data`。
8. rpcd `list` 包含无参数 `storage_check`。

### 静态和回归测试

- API 和 Storage View 通过 `node --check`。
- rpcd 和打包脚本通过 shell 语法检查。
- ACL 和菜单 JSON 通过 `jq` 解析。
- Storage View 中没有 `rpc.declare`、路径输入、文件上传、文件删除、`innerHTML` 或 `rawhtml`。
- `git diff --check` 通过。
- `go test ./...` 通过。

### 构建结果

```text
dist/subconv-next_1.0.7-17_aarch64_generic.ipk
dist/luci-app-subconv-next_1.0.7-17_all.ipk
```

`make package` 和独立 LuCI 包构建均通过。两个包均包含：

```text
/usr/libexec/rpcd/luci.subconv
/www/luci-static/resources/subconv-next/api.js
/www/luci-static/resources/subconv-next/storage.css
/www/luci-static/resources/view/subconv-next/storage.js
```

rpcd 模式为 `0755`，LuCI 资源为 `0644`。

### 真实 OpenWrt 验收

开发机没有 `opkg`、`ubus`、rpcd、LuCI 或浏览器环境。设备端应执行：

```sh
opkg install /tmp/subconv-next_1.0.7-17_aarch64_generic.ipk
ubus call luci.subconv storage_check '{}'
```

浏览器访问：

```text
/cgi-bin/luci/admin/services/subconv-next/storage
```

验证默认 overlay 目录、真实 `/mnt` 挂载盘、只读挂载和目录不存在状态。

## 6. 已知问题

- 本阶段只检查现有目录，不创建目录、不迁移数据，也不修改配置。
- `available_kb` 来自当前目录所在文件系统的 `df` 结果，不代表应用可独占使用的空间。
- Readable/Writable 反映 rpcd/service 用户的实际权限；当前 OpenWrt 服务通常以 root 运行。
- 允许路径目前限定为 `/etc/subconv-next`、`/mnt` 和 `/overlay` 子目录。使用其他自定义挂载点的用户需要先调整挂载路径或后续明确扩展允许列表。
- 页面尚未在真实 OpenWrt 主题和移动浏览器中截图验收。
