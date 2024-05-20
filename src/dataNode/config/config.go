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
	Port              int    `yaml:"Port"`
	DataNodeId        string `yaml:"DataNodeId"`
	TransFilePort     int    `yaml:"TransFilePort"`
	HeartbeatInterval int    `yaml:"HeartbeatInterval"`
	HeartBeatRetry    int    `yaml:"HeartBeatRetry"`
	DataDir           string `yaml:"DataDir"`
	MetaFileName      string `yaml:"MetaFileName"`
}

var conf *DataNodeConfig

func GetDataNodeConfig() *DataNodeConfig {
	return conf
}

func init() {
	fileName := "dataNode_config.yml"
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
	switch runtime.GOOS {
	case "windows":
		conf.DataNode.DataDir = conf.DataNode.DataDir + conf.DataNode.DataNodeId + "\\data"
	case "linux":
		conf.DataNode.DataDir = path.Join(conf.DataNode.DataDir+conf.DataNode.DataNodeId, "data")
	case "os":
	default:
		conf.DataNode.DataDir = path.Join(conf.DataNode.DataDir+conf.DataNode.DataNodeId, "data")
	}
	err = os.MkdirAll(conf.DataNode.DataDir, 0777)
	if err != nil {
		log.Fatal("fail to open nameNodeData dir  :", err)
	}
	fmt.Println("encode nameNode_config success")
}
