package main

import (
	"github.com/m4n5ter/lindows/pkg/yalog"
	"github.com/m4n5ter/lindows/winapi"
)

func main() {
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
