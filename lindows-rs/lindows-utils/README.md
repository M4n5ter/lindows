# Lindows utils

## Description

这个项目是一个用于 Lindows 的工具集合。

## Pre-requirements

### vcpkg

```powershell
> git clone https://github.com/microsoft/vcpkg
> .\vcpkg\bootstrap-vcpkg.bat
> .\vcpkg\vcpkg update
> .\vcpkg\vcpkg install libyuv:x64-windows-static --triplet=x64-windows-static
> .\vcpkg\vcpkg install libvpx:x64-windows-static --triplet=x86-windows-static
> .\vcpkg\vcpkg install aom:x64-windows-static --triplet=x64-windows-static
```

### rust with msvc

```powershell
> rustup toolchain install nightly-msvc
> rustup default nightly-msvc
```