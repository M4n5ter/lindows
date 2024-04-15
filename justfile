# set shell
set windows-shell := ["powershell.exe", "-c"]

# set `&&` or `;` for different OS
and := if os_family() == "windows" {";"} else {"&&"}

# load environment from `.env` file
set dotenv-load

#====================================== alias start ============================================#

alias a := add-hook
alias b := build
alias r := run
alias t := test
alias deps := dependencies

#======================================= alias end =============================================#

#===================================== targets start ===========================================#

# default target - `just` 默认目标
default: lint test

# PLEASE DO THIS FIRSET! - 务必先执行 `just add-hook`
[unix]
add-hook:
    @echo "just" > {{pre_commit}}
    @chmod +x {{pre_commit}}

# git hooks 只能识别 Unix LF
pre-commit-win := '#!/bin/sh
echo Checking...
just
'

[windows]
add-hook:
    @sh -c "echo '{{pre-commit-win}}' > pre-commit"
    @If (Test-Path {{pre_commit}}) { Remove-Item {{pre_commit}} }; mv pre-commit {{pre_commit}}

# go build
[unix]
build:
    @echo "Building..."
    @GIN_MODE=release go build -ldflags "-s -w" -o {{bin}} {{main_file}}
    @echo "Build done."

[windows]
build:
    @echo "Building..."
    @$env:GIN_MODE="release" {{and}} go build -ldflags "-s -w" -o {{bin}} {{main_file}}
    @echo "Build done."

# go run
run:
    @go run {{main_file}}

# go test
test:
    @go test -v {{join(".", "...")}}

# lint - 代码检查
lint: dep-golangci-lint
    @go mod tidy 
    @golangci-lint run

format: dep-gofumpt
    @gofumpt -extra -w {{root}}

# install dependencies - 安装依赖工具
dependencies: dep-golangci-lint dep-gofumpt

# a linter for Go - 一个 Go 语言的代码检查工具
dep-golangci-lint:
    @go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# a stricter gofmt - 一个更严格的 gofmt
dep-gofumpt:
    @go install mvdan.cc/gofumpt@latest

#===================================== targets end ===========================================#

#=================================== variables start =========================================#

# project name - 项目名称
project_name := "lindows"

# project root directory - 项目根目录
root := justfile_directory()

# binary path - go build 输出的二进制文件路径
bin := join(root, project_name)

# main.go path - main.go 文件路径
main_file := join(root, "main.go")

# pre-commit path - pre-commit 文件路径
pre_commit := join(root, ".git", "hooks", "pre-commit")

#=================================== variables end =========================================#