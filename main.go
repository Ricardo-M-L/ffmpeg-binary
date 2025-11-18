package main

import "goalfy-mediaconverter/internal/platform"

func main() {
	// 根据平台启动应用程序
	// Windows: 作为服务运行
	// 其他平台: 控制台模式运行
	platform.Start()
}
