# 03. OpenWrt 包设计

## 当前方案

发布时按包管理器提供对应格式：Kwrt/opkg 使用 IPK，启用 APK 的官方 OpenWrt 25.12 使用 APK。
两种格式的默认主包都同时包含后台服务和 LuCI 管理页面，普通用户只需安装 `subconv-next` 一个包。

IPK 文件名示例：

```text
subconv-next_1.0.0-3_aarch64_generic.ipk
```

该包同时包含：

- `/usr/bin/subconv-next`
- `/etc/init.d/subconv-next`
- `/etc/config/subconv-next`
- `/etc/subconv-next/data`
- `/usr/share/luci/menu.d/luci-app-subconv-next.json`
- `/usr/share/rpcd/acl.d/luci-app-subconv-next.json`
- `/www/luci-static/resources/view/subconv-next/overview.js`

安装后 `postinst` 会自动 enable 服务，并在 UCI `enabled=1` 或未设置时自动 start。脚本会清理 LuCI 缓存，但不会重启 `rpcd` 或 `uhttpd`，避免安装后打断当前 LuCI 登录会话。

卸载前 `prerm` 会 stop/disable 服务，但不会删除 `/etc/subconv-next/data`。

## Web UI 登录与访问令牌

OpenWrt 包默认免令牌登录，可直接访问 `http://路由器地址:9876/`。
UCI 的 `subconv-next.main.allow_insecure_public` 默认为 `1`；
旧配置未包含该选项时也默认免登录，首次启动不会生成令牌。
启动脚本通过 `SUBCONV_ALLOW_INSECURE_PUBLIC=true` 将该设置传给服务。
LuCI 本身仍使用路由器的登录认证。免登录时，所有能连接服务端口的设备均可使用管理功能。

如需启用令牌验证，在路由器上执行：

```sh
uci set subconv-next.main.allow_insecure_public='0'
uci commit subconv-next
/etc/init.d/subconv-next restart
uci -q get subconv-next.main.access_token
```

启用验证后，若令牌为空，启动脚本会生成 64 字符随机令牌并保存到
`subconv-next.main.access_token`。已有令牌会复用，配置文件权限为 `0600`。
包升级保留显式设置的 `allow_insecure_public=0`；设为 `1` 并重启即可恢复免登录。

## 依赖

主包依赖：

```text
ca-bundle, luci-base, rpcd, uci
```

OpenWrt SDK Makefile 中对应：

```makefile
DEPENDS:=+ca-bundle +luci-base +rpcd +uci
```

## 打包方式

Kwrt/opkg 设备默认使用 SDK 自带 `scripts/ipkg-build`：

```sh
SDK_DIR=/root/openwrt-sdk/openwrt-sdk-25.12.2-rockchip-armv8_gcc-14.3.0_musl.Linux-x86_64 \
VERSION=1.0.0 \
RELEASE=3 \
ARCH=aarch64_generic \
./scripts/package-openwrt-ipk-sdk.sh
```

输出：

```text
dist/subconv-next_1.0.0-3_aarch64_generic.ipk
```

不要把 APK 改名为 IPK，也不要恢复旧的手工 `ar` 打包路径；Kwrt/opkg 已验证手工 `ar` 包会被判定为 malformed。

官方 OpenWrt APK 固件使用匹配的 25.12 SDK 打包：

```sh
SDK_DIR=/path/to/openwrt-sdk-25.12.2-rockchip-armv8_gcc-14.3.0_musl.Linux-x86_64 \
VERSION=1.0.7 \
RELEASE=1 \
ARCH=aarch64_generic \
./scripts/package-openwrt-apk-sdk.sh dist/openwrt-arm64/subconv-next
```

生成 `dist/subconv-next-<版本>-r<发行号>.apk`，包含后台和 LuCI 页面。该 APK 只能安装到使用 apk 的固件：

```sh
apk add --allow-untrusted /tmp/subconv-next-<版本>-r<发行号>.apk
```

不要在使用 opkg 的 Kwrt 设备上安装 APK。

LuCI 源文件保留在 `openwrt/luci-app-subconv-next/`，但默认发布的 IPK 和 APK 都是包含后台与 LuCI 页面的单一 `subconv-next` 完整包。

## 验收

- `opkg install /tmp/subconv-next_1.0.0-3_aarch64_generic.ipk` 成功。
- 安装后 `/etc/init.d/subconv-next status` 显示服务已启动。
- `curl http://127.0.0.1:9876/healthz` 返回 OK。
- LuCI 出现 `Services / SubConv Next`。
- LuCI 页面能修改端口并重启服务。
- 卸载包不会删除 `/etc/subconv-next/data`。
