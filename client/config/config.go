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
	//fileName := "E:\\Projects\\GoLang-Projects\\TrainFS\\src\\client\\config\\client_config.yml"
	// 全部以 working directory 为 项目主目录!!!
	//fileName := "./client_config.yml" // package 在同级目录
	fileName := "./config/client_config.yml" // package 在上层目录
	//fileName := "src/client/config/client_config.yml" // package 在上层目录
	file, err := os.OpenFile(fileName, os.O_RDWR, 0777)
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
		log.Fatal("fail to yaml unmarshal:", err)
	}
	fmt.Println("encode client_config success.")
}
