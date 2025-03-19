package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path"
	"runtime"
)

type NameNodeConfig struct {
	NameNode NameNode `yaml:"NameNode"`
}

type NameNode struct {
	Host                      string `yaml:"Host"`
	NameNodeId                string `yaml:"NameNodeId"`
	DataNodeHeartBeatInterval int    `yaml:"DataNodeHeartBeatInterval"`
	DataNodeHeartBeatTimeout  int    `yaml:"DataNodeHeartBeatTimeout"`
	TrashInterval             int    `yaml:"TrashInterval"`
	MetaFileName              string `yaml:"MetaFileName"`
	DataDir                   string `yaml:"DataDir"`
	TaskDir                   string
	Model                     string   `yaml:"Model"`
	MeIndex                   int      `yaml:"MeIndex"`
	PeerNameNodes             []string `yaml:"PeerNameNodes"`
}

const (
	Single  string = "Single"
	Cluster string = "Cluster"
)

var conf *NameNodeConfig

func GetDataNodeConfig() *NameNodeConfig {
	inits()
	return conf
}

func inits() {
	//fileName := "nameNode_config.yml"
	fileName := "nameNode/config/nameNode_config.yml"
	file, err := os.OpenFile(fileName, os.O_RDWR, 0777)
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
	dataDir := conf.NameNode.DataDir
	switch runtime.GOOS {
	case "windows":
		conf.NameNode.DataDir = dataDir + conf.NameNode.NameNodeId + "\\data"
		conf.NameNode.TaskDir = dataDir + conf.NameNode.NameNodeId + "\\task"
	case "linux":
		conf.NameNode.DataDir = path.Join(dataDir+conf.NameNode.NameNodeId, "data")
		conf.NameNode.TaskDir = path.Join(dataDir+conf.NameNode.NameNodeId, "task")
	default:
		conf.NameNode.DataDir = path.Join(dataDir+conf.NameNode.NameNodeId, "data")
		conf.NameNode.TaskDir = path.Join(dataDir+conf.NameNode.NameNodeId, "task")
	}

	err = os.MkdirAll(conf.NameNode.DataDir, 0777)
	if err != nil {
		log.Fatal("fail to open nameNodeData dir  :", err)
	}
	err = os.MkdirAll(conf.NameNode.TaskDir, 0777)
	if err != nil {
		log.Fatal("fail to open nameNodeData dir  :", err)
	}
	model := conf.NameNode.Model
	if model == Single {
		return
	} else if model == Cluster {
		fmt.Printf("nameNode_config Cluster %v \n", conf.NameNode.PeerNameNodes)
	} else {
		log.Fatal("unknown the nameNode model: ", model)
	}
	fmt.Println("Decode nameNode_config success.")
}
