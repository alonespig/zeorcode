package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "zeorcode",
	Short: "ZeorCode 在线判题系统后端",
}

func init() {
	RootCmd.AddCommand(httpCmd)
	RootCmd.AddCommand(judgeCmd)
}
