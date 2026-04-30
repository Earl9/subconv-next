include $(TOPDIR)/rules.mk

PKG_NAME:=subconv-next
PKG_VERSION:=0.1.0
PKG_RELEASE:=1
PKG_LICENSE:=MIT
PKG_MAINTAINER:=SubConv Next

PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)-$(PKG_VERSION)
PKG_BUILD_DEPENDS:=golang/host
PKG_BUILD_PARALLEL:=1
PKG_USE_MIPS16:=0

SUBCONV_NEXT_SRC ?= $(TOPDIR)/../subconv-next
SUBCONV_NEXT_VERSION ?= $(PKG_VERSION)-$(PKG_RELEASE)

include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

GO_TARGET_ARCH:=$(GO_ARCH)
ifeq ($(GO_TARGET_ARCH),)
  ifeq ($(ARCH),aarch64)
    GO_TARGET_ARCH:=arm64
  else ifeq ($(ARCH),x86_64)
    GO_TARGET_ARCH:=amd64
  else ifeq ($(ARCH),i386)
    GO_TARGET_ARCH:=386
  else
    GO_TARGET_ARCH:=$(ARCH)
  endif
endif

define Package/subconv-next
  SECTION:=net
  CATEGORY:=Network
  TITLE:=SubConv Next Mihomo subscription generator
  URL:=https://github.com/subconv-next/subconv-next
  DEPENDS:=
endef

define Package/subconv-next/description
  Self-hosted subscription converter and Mihomo YAML generator.
endef

define Build/Prepare
	test -d "$(SUBCONV_NEXT_SRC)/cmd/subconv-next" || (echo "SUBCONV_NEXT_SRC must point to the subconv-next repository" >&2; exit 1)
	mkdir -p $(PKG_BUILD_DIR)
	tar -C $(SUBCONV_NEXT_SRC) \
		--exclude .git \
		--exclude data \
		--exclude dist \
		--exclude .cache \
		--exclude subconv-next \
		-cf - . | tar -C $(PKG_BUILD_DIR) -xf -
endef

define Build/Compile
	mkdir -p $(PKG_BUILD_DIR)/.cache/go-build $(PKG_BUILD_DIR)/.cache/go-mod
	(cd $(PKG_BUILD_DIR); \
		CGO_ENABLED=0 \
		GOOS=linux \
		GOARCH=$(GO_TARGET_ARCH) \
		GOARM=$(GO_ARM) \
		GOMIPS=$(GO_MIPS) \
		GOMIPS64=$(GO_MIPS64) \
		GOCACHE=$(PKG_BUILD_DIR)/.cache/go-build \
		GOMODCACHE=$(PKG_BUILD_DIR)/.cache/go-mod \
		$(STAGING_DIR_HOSTPKG)/bin/go build -trimpath -ldflags="-s -w -X main.version=$(SUBCONV_NEXT_VERSION)" -o subconv-next ./cmd/subconv-next)
endef

define Package/subconv-next/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/subconv-next $(1)/usr/bin/subconv-next
	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) ./files/subconv-next.init $(1)/etc/init.d/subconv-next
	$(INSTALL_DIR) $(1)/etc/config
	$(INSTALL_CONF) ./files/subconv-next.config $(1)/etc/config/subconv-next
	$(INSTALL_DIR) $(1)/etc/subconv-next/data
endef

define Package/subconv-next/postinst
#!/bin/sh
[ -n "$${IPKG_INSTROOT}" ] || /etc/init.d/subconv-next enable
exit 0
endef

define Package/subconv-next/prerm
#!/bin/sh
[ -n "$${IPKG_INSTROOT}" ] || /etc/init.d/subconv-next disable
exit 0
endef

$(eval $(call BuildPackage,subconv-next))
