package main

import (
	"github.com/jinzhu/configor"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

const ConfigRoot = "./configs/"

var (
	configLock  = new(sync.RWMutex)
	configFiles = make([]string, 0)
)

type ConfigType struct {
	Server Server
	Nodes
}

type Server struct {
	Host  string
	Topic string
}

type Nodes []Node

type Node struct {
	Host  string
	Topic string
}

// GetConfig - get application config
func GetConfig() (Config ConfigType, err error) {

	configLock.RLock()
	defer configLock.RUnlock()
	if len(configFiles) == 0 {
		for _, file := range getConfigFiles() {
			if isSetConfig(file) {
				configFiles = append(configFiles, ConfigRoot+file.Name())
			}
		}
	}
	err = configor.Load(&Config, configFiles...)

	return
}

// ReloadConfig - reload application config
func ReloadConfig() (cfg ConfigType, err error) {
	cfg, err = GetConfig()
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

// Read DIR for merge configs
func getConfigFiles() (files []os.FileInfo) {
	files, err := ioutil.ReadDir(ConfigRoot)
	if err != nil {
		log.Fatal(err)
	}

	return
}

// Check permitted config types
func isSetConfig(file os.FileInfo) bool {
	switch {
	case strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml"):
		return true
	case strings.HasSuffix(file.Name(), ".toml") || strings.HasSuffix(file.Name(), ".json"):
		return true
	default:
		return false
	}
}
