package service

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path/filepath"
)

type DataNodeConfig struct {
	Config Config `yaml:"Config"`
}

type Config struct {
	Host              string `yaml:"Host"`
	NameNodeHost      string `yaml:"NameNodeHost"`
	DataNodeId        string `yaml:"DataNodeId"`
	HeartbeatInterval int    `yaml:"HeartbeatInterval"`
	HeartBeatRetry    int    `yaml:"HeartBeatRetry"`
	DataDir           string `yaml:"DataDir"`
	TaskDir           string
	MetaDir           string
	MetaFileName      string `yaml:"MetaFileName"`
}

func GetDataNodeConfig(configFile *string, hostPort *string, dataNodeId *string) *Config {
	var conf *DataNodeConfig
	file, err := os.OpenFile(*configFile, os.O_RDWR|os.O_CREATE, 0777)
	defer file.Close()
	if err != nil {
		log.Fatalf(" fail to read fileName: %s, err: %s ;\n", configFile, err)
	}
	v, err := file.Stat()
	size := v.Size()
	buf := make([]byte, size)
	_, err = file.ReadAt(buf, 0)
	if err != nil {
		log.Fatal("fail to read yaml ", err)
	}
	err = yaml.Unmarshal(buf, &conf)
	if err != nil {
		log.Fatal("fail to unmarshal yaml :", err)
	}
	if conf == nil {
		log.Fatal("fail to unmarshal yaml :", err)
	}

	if hostPort != nil {
		conf.Config.Host = *hostPort
	}
	if dataNodeId != nil {
		conf.Config.DataNodeId = *dataNodeId
	}

	dataDir := conf.Config.DataDir
	conf.Config.DataDir = filepath.Join(dataDir+conf.Config.DataNodeId, "data")
	conf.Config.TaskDir = filepath.Join(dataDir+conf.Config.DataNodeId, "task")
	conf.Config.MetaDir = filepath.Join(dataDir+conf.Config.DataNodeId, "meta")

	err = os.MkdirAll(conf.Config.DataDir, 0777)
	err = os.MkdirAll(conf.Config.TaskDir, 0777)
	err = os.MkdirAll(conf.Config.MetaDir, 0777)

	if err != nil {
		log.Fatal("fail to open nameNodeData dir  :", err)
	} else {
		fmt.Println("Decode dataNode_config success.")
	}

	return &conf.Config
}
