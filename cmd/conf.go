//
// Copyright (c) 2020-present Codist <countstarlight@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// Written by Codist <countstarlight@gmail.com>, March 2020
//

package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"strings"
)

const (
	aiicyCoreModulePrefix = "aiicy-"
	aiicyFunctionPrefix   = "aiicy-function-"
)

var (
	confDocker bool
	confCmd    = &cobra.Command{
		Use:   "conf",
		Short: "Check required conf files and generate application.yml",
		Run: func(cmd *cobra.Command, args []string) {
			if confDocker {
				log.Info("Is docker")
			} else {
				genNativeConf()
			}
		},
	}
)

func init() {
	confCmd.PersistentFlags().Bool("native", true, "Check and generate for native")
	confCmd.PersistentFlags().BoolVar(&confDocker, "docker", false, "Check and generate for docker")
}

func genNativeConf() {
	f, err := os.Open(WorkPath)
	log.Info("workdir: " + WorkPath)
	if err != nil {
		log.Fatal(err)
	}
	files, err := f.Readdir(-1)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() && strings.HasPrefix(file.Name(), aiicyCoreModulePrefix) && !strings.HasPrefix(file.Name(), aiicyFunctionPrefix) {
			log.Info("find aiicy module: " + file.Name())
		}
	}
}
