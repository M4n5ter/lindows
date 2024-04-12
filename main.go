//go:build windows

package main

import (
	"os"

	"github.com/m4n5ter/lindows/pkg/yalog"
	"github.com/m4n5ter/lindows/winapi"
)

func main() {
	// 注意：请使用您想要查找窗口的确切名称和/或类名
	hwnd, err := winapi.FindWindow("", "QQ")
	if err != nil {
		yalog.Error("查找窗口失败", "err", err)
		os.Exit(1)
	}
	if hwnd == 0 {
		yalog.Error("QQ 窗口未找到")
	} else {
		yalog.Info("找到 QQ 窗口", "hwnd", hwnd)
	}

	className, err := winapi.GetClassName(hwnd)
	if err != nil {
		yalog.Error("获取窗口类名失败", "err", err)
		os.Exit(1)
	}
	yalog.Info("QQ 窗口类名", "className", className)

	winapi.EnumWindows(func(hwnd winapi.HWND, lParam uintptr) bool {
		title, err := winapi.GetWindowText(hwnd)
		if err != nil {
			yalog.Error("获取窗口标题失败", "err", err)
			return true
		}
		if title == "计算器" {
			yalog.Info("找到计算器窗口", "hwnd", hwnd)
			return false

		}
		return true
	}, 0)
}
