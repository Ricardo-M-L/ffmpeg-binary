//go:build !windows

package main

// runAsWindowsService 是 Windows 平台专用函数的存根
// 在非 Windows 平台上,这个函数永远不会被调用
func runAsWindowsService() {
	panic("runAsWindowsService should not be called on non-Windows platforms")
}
