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
	Host                      string   `yaml:"Host"`
	NameNodeId                string   `yaml:"NameNodeId"`
	DataNodeHeartBeatInterval int      `yaml:"DataNodeHeartBeatInterval"`
	DataNodeHeartBeatTimeout  int      `yaml:"DataNodeHeartBeatTimeout"`
	TrashInterval             int      `yaml:"TrashInterval"`
	MetaFileName              string   `yaml:"MetaFileName"`
	DataDir                   string   `yaml:"DataDir"`
	Model                     string   `yaml:"Model"`
	ModelDir                  string   `yaml:"ModelDir"`
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
	fileName := "src/nameNode/config/nameNode_config.yml"
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
		conf.NameNode.DataDir = conf.NameNode.DataDir + conf.NameNode.NameNodeId + "\\data"
	case "linux":
		conf.NameNode.DataDir = path.Join(conf.NameNode.DataDir+conf.NameNode.NameNodeId, "data")
	case "os":
	default:
		conf.NameNode.DataDir = path.Join(conf.NameNode.DataDir+conf.NameNode.NameNodeId, "data")
	}

	err = os.MkdirAll(conf.NameNode.DataDir, 0777)
	if err != nil {
		log.Fatal("fail to open nameNodeData dir  :", err)
	}

	model := conf.NameNode.Model
	if model == Single {
		return
	} else if model == Cluster {
		fmt.Printf("encode nameNode_config %v \n", conf.NameNode.PeerNameNodes[2])
	} else {
		log.Fatal("unknown the nameNode model : ", model)
	}
	fmt.Println("encode nameNode_config success")
}
