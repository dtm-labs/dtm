/*
 * Copyright (c) 2013 - 2020. 青木文化传播有限公司 版权所有.
 * DO NOT ALTER OR REMOVE COPYRIGHT NOTICES OR THIS FILE HEADER.
 *
 * File:    cmd.go
 * Created: 2020/7/23 16:51:5
 * Authors: MS geek.snail@qq.com
 */

package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	BinBuildVersion = ""
	BinBuildCommit  = ""
	BinBuildDate    = ""
	BinAppName      = ""
)

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "version",
	Long:  `version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("AppName   : %s\n", BinAppName)
		fmt.Printf("Version   : %s\n", BinBuildVersion)
		fmt.Printf("Commit    : %s\n", BinBuildCommit)
		fmt.Printf("BuildDate : %s\n", BinBuildDate)
	},
}
