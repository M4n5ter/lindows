//go:build windows

package main

import (
	"fmt"
	"syscall"

	"github.com/m4n5ter/lindows/winapi"
)

func main() {
	// 注意：请使用您想要查找窗口的确切名称和/或类名
	hwnd, err := winapi.FindWindow("", "QQ")
	if err != nil {
		fmt.Println(err)
		return
	}
	if hwnd == 0 {
		fmt.Println("QQ 窗口未找到")
	} else {
		fmt.Printf("找到 QQ 窗口，句柄为: %d\n", hwnd)
	}

	className, err := winapi.GetClassName(hwnd)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("QQ 窗口类名: %s\n", className)

	winapi.EnumWindows(func(hwnd syscall.Handle, lParam uintptr) bool {
		title, err := winapi.GetWindowText(hwnd)
		if err != nil {
			fmt.Println(err)
			return true
		}
		if title == "计算器" {
			fmt.Printf("找到计算器窗口，句柄为: %d\n", hwnd)
			return false

		}
		return true
	}, 0)
}
