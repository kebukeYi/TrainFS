package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

type ClientConfig struct {
	Client Client `yaml:"Client"`
}

type Client struct {
	ClientId int      `yaml:"ClientId"`
	NameNode NameNode `yaml:"NameNode"`
}

type NameNode struct {
	Host            string `yaml:"Host"`
	ChunkSize       int64  `yaml:"ChunkSize"`
	ChunkReplicaNum int32  `yaml:"ChunkReplicaNum"`
	ConfirmFileAck  bool   `yaml:"ConfirmFileAck"`
}

var conf *ClientConfig

func GetClientConfig() *ClientConfig {
	inits()
	return conf
}

func inits() {
	fileName := "src/client/config/client_config.yml"
	//fileName := "client_config.yml"
	//fileName := "E:\\Projects\\GoLang-Projects\\TrainFS\\src\\client\\config\\client_config.yml"
	file, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatalf(" fail to read fileName: %s, err: %s ;\n", fileName, err)
	}
	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		log.Fatal("fail to yaml unmarshal:", err)
	}
	fmt.Println("encode client_config success.")
}
