# 03. OpenWrt 包设计

## 包拆分

V1 拆成两个 OpenWrt 包：

```text
subconv-next
luci-app-subconv-next
```

### `subconv-next`

包含：

- `/usr/bin/subconv-next`
- `/etc/config/subconv_next`
- `/etc/init.d/subconv-next`
- `/usr/share/subconv-next/templates/*.yaml`

### `luci-app-subconv-next`

包含：

- LuCI 菜单文件
- LuCI JS views
- rpcd ACL
- i18n 可后续补

## `subconv-next` Makefile 草案

Codex 应创建：

```makefile
include $(TOPDIR)/rules.mk

PKG_NAME:=subconv-next
PKG_VERSION:=0.1.0
PKG_RELEASE:=1

PKG_MAINTAINER:=SubConv Next Developers
PKG_LICENSE:=MIT

include $(INCLUDE_DIR)/package.mk

define Package/subconv-next
  SECTION:=net
  CATEGORY:=Network
  TITLE:=Modern subscription converter for Mihomo
  DEPENDS:=+ca-bundle
endef

define Package/subconv-next/description
  OpenWrt-first subscription converter and Mihomo YAML generator.
endef

define Build/Prepare
	mkdir -p $(PKG_BUILD_DIR)
	$(CP) ../../../../* $(PKG_BUILD_DIR)/
endef

define Build/Compile
	cd $(PKG_BUILD_DIR) && \
		CGO_ENABLED=0 GOOS=linux GOARCH=$(GO_ARCH) go build \
		-trimpath -ldflags="-s -w" \
		-o subconv-next ./cmd/subconv-next
endef

define Package/subconv-next/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/subconv-next $(1)/usr/bin/subconv-next

	$(INSTALL_DIR) $(1)/etc/config
	$(INSTALL_CONF) ./files/etc/config/subconv_next $(1)/etc/config/subconv_next

	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) ./files/etc/init.d/subconv-next $(1)/etc/init.d/subconv-next

	$(INSTALL_DIR) $(1)/usr/share/subconv-next/templates
	$(INSTALL_DATA) ./files/usr/share/subconv-next/templates/*.yaml $(1)/usr/share/subconv-next/templates/
endef

$(eval $(call BuildPackage,subconv-next))
```

Codex 注意：`GO_ARCH` 在实际 OpenWrt Go package 支持里可能需要映射，若 SDK 中无现成 Go helper，则先提供外部交叉编译 Make target，并让 OpenWrt 包安装预构建二进制。不要卡死在完美 OpenWrt Go 集成上。

## init script 草案

路径：

```text
package/openwrt/subconv-next/files/etc/init.d/subconv-next
```

内容：

```sh
#!/bin/sh /etc/rc.common

START=95
STOP=10
USE_PROCD=1

PROG=/usr/bin/subconv-next
CONFIG=/etc/config/subconv_next

start_service() {
	procd_open_instance
	procd_set_param command "$PROG" serve --config "$CONFIG"
	procd_set_param respawn 3600 5 5
	procd_set_param stdout 1
	procd_set_param stderr 1
	procd_close_instance
}

reload_service() {
	procd_send_signal subconv-next
}

service_triggers() {
	procd_add_reload_trigger "subconv_next"
}
```

## 默认 UCI 配置

路径：

```text
package/openwrt/subconv-next/files/etc/config/subconv_next
```

内容：

```uci
config service 'main'
	option enabled '1'
	option listen_addr '127.0.0.1'
	option listen_port '9876'
	option log_level 'info'
	option template 'standard'
	option output_path '/var/run/subconv-next/mihomo.yaml'
	option cache_dir '/var/run/subconv-next/cache'
	option state_path '/etc/subconv-next/state.json'
	option refresh_interval '3600'
	option max_subscription_bytes '5242880'
	option fetch_timeout_seconds '15'
	option allow_lan '0'

config subscription
	option name 'example'
	option enabled '0'
	option url ''
	option user_agent 'SubConvNext/0.1 OpenWrt'
```

## LuCI 包 Makefile 草案

```makefile
include $(TOPDIR)/rules.mk

PKG_NAME:=luci-app-subconv-next
PKG_VERSION:=0.1.0
PKG_RELEASE:=1

LUCI_TITLE:=LuCI support for SubConv Next
LUCI_DEPENDS:=+subconv-next
LUCI_PKGARCH:=all

include $(TOPDIR)/feeds/luci/luci.mk

$(eval $(call BuildPackage,luci-app-subconv-next))
```

## 安装后命令

```sh
opkg install subconv-next_*.ipk luci-app-subconv-next_*.ipk
/etc/init.d/subconv-next enable
/etc/init.d/subconv-next start
logread -f | grep subconv
```

## Codex 验收

- `/etc/init.d/subconv-next start` 成功。
- `curl http://127.0.0.1:9876/healthz` 返回 JSON。
- 修改 `/etc/config/subconv_next` 后执行 `/etc/init.d/subconv-next reload` 生效。
- 卸载包时不删除用户配置，除非用户 purge。
