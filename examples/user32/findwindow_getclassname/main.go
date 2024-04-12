package main

import (
	"os"

	"github.com/m4n5ter/lindows/pkg/yalog"
	"github.com/m4n5ter/lindows/winapi"
)

func main() {
	hwnd, err := winapi.FindWindow("", "计算器")
	if err != nil {
		yalog.Error("查找窗口失败", "err", err)
		os.Exit(1)
	}
	if hwnd == 0 {
		yalog.Error("计算器 窗口未找到")
	} else {
		yalog.Info("找到 计算器 窗口", "hwnd", hwnd)
	}

	className, err := winapi.GetClassName(hwnd)
	if err != nil {
		yalog.Error("获取窗口类名失败", "err", err)
		os.Exit(1)
	}
	yalog.Info("QQ 窗口类名", "className", className)
}
