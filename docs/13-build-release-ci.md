# 13. 构建、发布与 CI

## Makefile

仓库根目录创建：

```makefile
APP=subconv-next

.PHONY: test build clean run

test:
	go test ./...

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/$(APP) ./cmd/subconv-next

run:
	go run ./cmd/subconv-next serve --config ./testdata/config/basic.json

clean:
	rm -rf bin dist
```

## 本地构建

```sh
make test
make build
./bin/subconv-next version
```

## 多架构交叉编译

先提供普通 Linux 交叉编译产物：

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/subconv-next_linux_amd64 ./cmd/subconv-next
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/subconv-next_linux_arm64 ./cmd/subconv-next
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o dist/subconv-next_linux_armv7 ./cmd/subconv-next
CGO_ENABLED=0 GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -o dist/subconv-next_linux_mipsle_softfloat ./cmd/subconv-next
```

## OpenWrt SDK 构建策略

V1 允许两条路线：

### 路线 A：SDK 内构建 Go

如果 SDK 和 feeds 支持 Go helper，则在 OpenWrt Makefile 中直接编译。

### 路线 B：预构建二进制打包

如果 SDK 内 Go 构建不稳定，先用 GitHub Actions 交叉编译 binary，再让 OpenWrt package 安装对应架构 binary。

Codex 优先保证路线 B 可用。

## GitHub Actions

创建 `.github/workflows/ci.yml`：

- checkout
- setup-go
- go test
- build linux amd64/arm64/armv7/mipsle
- upload artifacts

### Auto Release

`.github/workflows/auto-release.yml` 会在 `main` 每次更新后自动发布 GitHub Release：

- 自动生成版本号：`1.0.<run number>`，Git tag 为 `v1.0.<run number>`。
- 用该版本号构建 Linux 多架构二进制。
- 使用仓库内 portable `ipkg-build` 打包 `aarch64_generic` all-in-one OpenWrt IPK，不需要配置 OpenWrt SDK URL。
- 推送 Docker 镜像到 GHCR：`latest` 和当前版本 tag。
- 创建 Git tag 和 GitHub Release，并上传二进制、OpenWrt IPK 与 `checksums.txt`。

提交信息包含 `[skip release]` 时，只保留普通 CI，不自动发版。

## Release Artifacts

V1 release 包含：

```text
subconv-next-linux-amd64
subconv-next-linux-arm64
subconv-next-linux-armv7
subconv-next-linux-mips-softfloat
subconv-next-linux-mipsle-softfloat
subconv-next_<version>-1_aarch64_generic.ipk
checksums.txt
```

## 版本命名

```text
v0.1.0
```

CLI 输出：

```sh
subconv-next version
```

结果：

```text
subconv-next version 0.1.0 commit <sha> built <date>
```
