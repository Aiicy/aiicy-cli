//
// Copyright (c) 2020-present Codist <countstarlight@gmail.com>. All rights reserved.
// Use of this source code is governed by Apache License 2.0 that can
// be found in the LICENSE file.
// Written by Codist <countstarlight@gmail.com>, March 2020
//

package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/aiicy/aiicy-go/logger"
	"github.com/spf13/cobra"
)

var (
	debugMode bool
	IsWindows bool
	WorkPath  string
	log       = logger.S
	rootCmd   = &cobra.Command{
		Use:   "aiicy-cli",
		Short: "The command tool for Aiicy",
	}
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of aiicy-cli",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Aiicy Command Tool v0.0.1 -- HEAD")
		},
	}
)

func init() {
	//  global
	IsWindows = runtime.GOOS == "windows"
	var err error
	if WorkPath, err = getWorkPath(); err != nil {
		log.Fatal("Failed to get app path: %s", err.Error())
	}
	// command
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "debug mode")
	rootCmd.AddCommand(confCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if !debugMode {
			log = logger.New(logger.LogConfig{Level: "info"})
		}
	}
}

func Execute() error {
	return rootCmd.Execute()
}

func getWorkPath() (string, error) {
	path, err := os.Getwd()
	return path, err
}
