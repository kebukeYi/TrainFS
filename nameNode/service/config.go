package service

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path/filepath"
)

type NameNodeConfig struct {
	Config Config `yaml:"Config"`
}

type Config struct {
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

func GetDataNodeConfig(confPath *string) *NameNodeConfig {
	conf := &NameNodeConfig{}
	file, err := os.OpenFile(*confPath, os.O_RDWR, 0777)
	defer file.Close()
	if err != nil {
		log.Fatalf(" fail to read fileName: %s, err: %s ;\n", confPath, err)
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
	dataDir := conf.Config.DataDir

	conf.Config.DataDir = filepath.Join(dataDir+conf.Config.NameNodeId, "data")
	conf.Config.TaskDir = filepath.Join(dataDir+conf.Config.NameNodeId, "task")

	err = os.MkdirAll(conf.Config.DataDir, os.ModePerm)
	if err != nil {
		log.Fatal("fail to open nameNodeData dir  :", err)
	}
	err = os.MkdirAll(conf.Config.TaskDir, os.ModePerm)
	if err != nil {
		log.Fatal("fail to open nameNodeData dir  :", err)
	}

	model := conf.Config.Model
	if model == Single {
		return conf
	} else if model == Cluster {
		fmt.Printf("nameNode_config Cluster %v \n", conf.Config.PeerNameNodes)
	} else {
		log.Fatal("unknown the nameNode model: ", model)
	}
	fmt.Println("Decode nameNode_config success.")
	return conf
}
