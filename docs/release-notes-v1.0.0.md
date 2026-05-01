# SubConv Next v1.0.0 Release Notes

SubConv Next v1.0.0 provides the Go daemon with embedded Web UI, Docker images, Linux binaries, and OpenWrt/Kwrt all-in-one IPK packages.

## Release Assets

Verified:

- `subconv-next-linux-amd64`
- `subconv-next-linux-arm64`
- `subconv-next_1.0.0-4_aarch64_generic.ipk` verified on Kwrt 25.12.2 `rockchip/armv8`

Experimental:

- `subconv-next_1.0.0-4_x86_64.ipk`
- `subconv-next_1.0.0-4_arm_cortex-a7_neon-vfpv4.ipk`
- `subconv-next_1.0.0-4_arm_cortex-a9_vfpv3-d16.ipk`
- `subconv-next_1.0.0-4_mips_24kc.ipk`
- `subconv-next_1.0.0-4_mipsel_24kc.ipk`

Checksums:

- `checksums.txt` contains all final uploaded assets.

## OpenWrt / Kwrt Notes

The IPK packages include the backend binary, init script, UCI config, data directory, LuCI menu, rpcd ACL, and LuCI management page.

The multi-architecture IPKs are static Go cross-compiled binaries packaged via `ipkg-build`. They were not built separately with each target's official OpenWrt SDK.

Use `opkg print-architecture` on the router and choose the matching IPK. Do not install an experimental package unless the package architecture matches the device.

## Cross-Compile Parameters

```text
x86_64: GOOS=linux GOARCH=amd64 CGO_ENABLED=0
aarch64_generic: GOOS=linux GOARCH=arm64 CGO_ENABLED=0
arm_cortex-a7_neon-vfpv4: GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0
arm_cortex-a9_vfpv3-d16: GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0
mips_24kc: GOOS=linux GOARCH=mips GOMIPS=softfloat CGO_ENABLED=0
mipsel_24kc: GOOS=linux GOARCH=mipsle GOMIPS=softfloat CGO_ENABLED=0
```
