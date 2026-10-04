package main

import (
	"zoj/cmd"
	"zoj/internal/config"
	"zoj/internal/infra/logger"

	_ "zoj/pkg/remoteoj/all" // 触发各远程 OJ 实现自注册
)

func main() {
	config.Init()
	logger.Init(logger.ModeDev)
	defer logger.Sync()

	// Redis / MySQL 等资源由各子命令(http / judge)在自己的组合根里创建并注入
	cmd.RootCmd.Execute()
}
