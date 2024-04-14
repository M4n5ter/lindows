# 如何运行

## RUST

Windows 请在 Git Bash 中运行以下命令

```bash
export RUSTUP_DIST_SERVER="https://rsproxy.cn"
export RUSTUP_UPDATE_ROOT="https://rsproxy.cn/rustup"
curl --proto '=https' --tlsv1.2 -sSf https://rsproxy.cn/rustup-init.sh | sh
```

修改 `~/.cargo/config` 文件，添加如下内容

```toml
[source.crates-io]
replace-with = 'rsproxy-sparse'
[source.rsproxy]
registry = "https://rsproxy.cn/crates.io-index"
[source.rsproxy-sparse]
registry = "sparse+https://rsproxy.cn/index/"
[registries.rsproxy]
index = "https://rsproxy.cn/crates.io-index"
[net]
git-fetch-with-cli = true
```

安装 `tauri-cli` 和 `trunk`

```bash
cargo install cargo-binstall
cargo binstall tauri-cli
cargo binstall trunk
```

## NPM

从 [Node.js 官网](https://nodejs.org/en/download) 下载并安装 Node.js

设置镜像源

```bash
npm config set registry https://registry.npmmirror.com
```

安装依赖

```bash
npm install
```

## 运行

```bash
cargo tauri dev
```
