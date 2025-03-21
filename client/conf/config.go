package conf

import (
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
	//fileName := "client/conf/client_config.yml" // package 在同级目录
	fileName := "./conf/client_config.yml" // package 在同级目录
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
		log.Fatal("fail to yaml unmarshal:", err)
	}
}
