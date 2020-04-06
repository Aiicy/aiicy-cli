//
// Copyright (c) 2020-present Codist <countstarlight@gmail.com>. All rights reserved.
// Use of this source code is governed by Apache License 2.0 that can
// be found in the LICENSE file.
// Written by Codist <countstarlight@gmail.com>, March 2020
//

package cmd

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aiicy/aiicy-cli/utils"
	"github.com/aiicy/aiicy-go"
	"github.com/spf13/cobra"
)

const (
	aiicyCoreModulePrefix = "aiicy-"
	aiicyFunctionPrefix   = "aiicy-function-"
	aiicyConfFile         = "application.yml"
	aiicyModuleConfFile   = "service.yml"
	aiicyModuleConfSuffix = "-conf"
	aiicyMasterConfFile   = "conf.yml"
)

type ModuleConfig struct {
	Name    string
	BinDir  string
	ConfDir string
}

var (
	// global
	// conf source
	aiicyMasterConfPath = filepath.Join("master", "engine")
	// conf target
	confPath = filepath.Join("var", "db", "aiicy")
	// bin path, native mod only
	LibPath         = filepath.Join("lib", "aiicy")
	coreConfPath    = filepath.Join("etc", "aiicy")
	aiicyModuleList []ModuleConfig

	// native
	// conf source
	aiicyMasterNativeConfFile = filepath.Join(aiicyMasterConfPath, "native", aiicyMasterConfFile)
	// conf target
	nativeConfPath       = filepath.Join("conf", "native", confPath)
	nativeConfFile       string
	nativeMasterConfPath = filepath.Join("conf", "native", coreConfPath)
	nativeMasterConfFile = filepath.Join(nativeMasterConfPath, aiicyMasterConfFile)
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
	// TODO: migrate
	f, err := os.Open(WorkPath)
	log.Info("workdir: " + WorkPath)
	if err != nil {
		log.Fatal(err)
	}
	nativeConfFile = path.Join(WorkPath, nativeConfPath, aiicyConfFile)
	files, err := f.Readdir(-1)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		// skip aiicy-function module and aiicy-ui
		// TODO: aiicy-ui
		if file.IsDir() && strings.HasPrefix(file.Name(), aiicyCoreModulePrefix) && !strings.HasPrefix(file.Name(), aiicyFunctionPrefix) && !strings.HasPrefix(file.Name(), "aiicy-ui") {
			log.Info("find aiicy module: " + file.Name())
			aiicyModuleList = append(aiicyModuleList, ModuleConfig{
				Name:    file.Name(),
				ConfDir: file.Name() + aiicyModuleConfSuffix,
			})
		}
	}
	// 1.check conf file exist or not
	// create conf file path if not exist
	if !utils.DirExists(nativeConfPath) {
		err = os.MkdirAll(nativeConfPath, os.ModePerm)
		if err != nil {
			log.Fatalf("create dir %s failed: %s", nativeConfPath, err.Error())
		}
	}
	if utils.FileExists(nativeConfFile) {
		log.Infof("find exists conf file: %s", nativeConfFile)
		log.Infof("backup to %s", nativeConfFile+".bak")
		err = os.Rename(nativeConfFile, nativeConfFile+".bak")
		if err != nil {
			log.Fatalf("rename to %s failed: %s", nativeConfFile+".bak", err.Error())
		}
	}
	// 2.check and copy module conf file
	// check
	for _, m := range aiicyModuleList {
		if !utils.FileExists(path.Join(WorkPath, m.Name, aiicyModuleConfFile)) {
			log.Fatalf("module %s conf file '%s' not found", m.Name, aiicyModuleConfFile)
		}
	}
	// check master
	if !utils.FileExists(path.Join(WorkPath, aiicyMasterNativeConfFile)) {
		log.Fatalf("master conf file '%s' not found", aiicyModuleConfFile)
	}
	// copy
	for _, m := range aiicyModuleList {
		moduleConfPath := filepath.Join(WorkPath, nativeConfPath, m.ConfDir)
		if !utils.DirExists(moduleConfPath) {
			err = os.MkdirAll(moduleConfPath, os.ModePerm)
			if err != nil {
				log.Fatalf("create dir %s failed: %s", moduleConfPath, err.Error())
			}
		}
		// copy file
		err = utils.CopyFile(path.Join(WorkPath, m.Name, aiicyModuleConfFile), path.Join(moduleConfPath, aiicyModuleConfFile))
		if err != nil {
			log.Fatalf("copy module %s conf file '%s' failed: %s", m, err.Error())
		}
	}
	// copy master conf file
	if !utils.DirExists(filepath.Join(WorkPath, nativeMasterConfPath)) {
		err = os.MkdirAll(nativeMasterConfPath, os.ModePerm)
		if err != nil {
			log.Fatalf("create dir %s failed: %s", nativeMasterConfPath, err.Error())
		}
	}
	err = utils.CopyFile(path.Join(WorkPath, aiicyMasterNativeConfFile), path.Join(WorkPath, nativeMasterConfFile))
	if err != nil {
		log.Fatalf("copy master conf file '%s' failed: %s", aiicyMasterNativeConfFile, err.Error())
	}

	// 3.generate aiicy conf file
	cfg := aiicy.ComposeAppConfig{
		// TODO: check and set aiicy version
		Version:    "3",
		AppVersion: "v2",
	}
	aiicyModuleServices := make(map[string]aiicy.ComposeService)

	for _, m := range aiicyModuleList {
		aiicyModuleServices[m.Name] = aiicy.ComposeService{
			Image:   m.Name,
			Replica: 1,
			Volumes: []aiicy.ServiceVolume{
				// mount conf file
				{
					Source:   filepath.Join(confPath, m.ConfDir),
					Target:   "etc/aiicy",
					ReadOnly: true,
				},
				// mount bin dir
				{
					Source:   filepath.Join(confPath, m.Name),
					Target:   filepath.Join(LibPath, m.Name),
					ReadOnly: true,
				},
			},
		}
	}
	cfg.Services = aiicyModuleServices
	err = aiicy.CreateComposeAppConfigCompatible(cfg, nativeConfFile)
	if err != nil {
		log.Fatalf("write conf file %s failed: %s", nativeConfFile, err.Error())
	}
}
