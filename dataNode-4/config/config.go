package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path"
	"runtime"
)

type DataNodeConfig struct {
	DataNode DataNode `yaml:"DataNode"`
}

type DataNode struct {
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

var conf *DataNodeConfig

func GetDataNodeConfig() *DataNode {
	inits()
	return &conf.DataNode
}

func inits() {
	fileName := "dataNode-4/config/dataNode_config.yml"
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0777)
	defer file.Close()
	if err != nil {
		log.Fatalf(" fail to read fileName: %s, err: %s ;\n", fileName, err)
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
	dataDir := conf.DataNode.DataDir
	switch runtime.GOOS {
	case "windows":
		conf.DataNode.DataDir = dataDir + conf.DataNode.DataNodeId + "//data"
		conf.DataNode.TaskDir = dataDir + conf.DataNode.DataNodeId + "//task"
		conf.DataNode.MetaDir = dataDir + conf.DataNode.DataNodeId + "//meta"
	case "linux":
		conf.DataNode.DataDir = path.Join(dataDir+conf.DataNode.DataNodeId, "data")
		conf.DataNode.TaskDir = path.Join(dataDir+conf.DataNode.DataNodeId, "task")
		conf.DataNode.MetaDir = path.Join(dataDir+conf.DataNode.DataNodeId, "meta")
	default:
		conf.DataNode.DataDir = path.Join(dataDir+conf.DataNode.DataNodeId, "data")
		conf.DataNode.TaskDir = path.Join(dataDir+conf.DataNode.DataNodeId, "task")
		conf.DataNode.MetaDir = path.Join(dataDir+conf.DataNode.DataNodeId, "meta")
	}

	err = os.MkdirAll(conf.DataNode.DataDir, 0777)
	err = os.MkdirAll(conf.DataNode.TaskDir, 0777)
	err = os.MkdirAll(conf.DataNode.MetaDir, 0777)

	if err != nil {
		log.Fatal("fail to open nameNodeData dir  :", err)
	} else {
		fmt.Println("Decode dataNode_config success.")
	}
}
