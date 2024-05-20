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
	NameNode NameNode `yaml:"NameNode"`
}

type NameNode struct {
	Host string `yaml:"Host"`
	Port int    `yaml:"Port"`
}

var conf *ClientConfig

func GetClientConfig() *ClientConfig {
	return conf
}

func init() {
	fileName := "client_config.yml"
	file, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatalf(" fail to read fileName: %s, err: %s ;\n", fileName, err)
	}
	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		log.Fatal("fail to yaml unmarshal:", err)
	}
	fmt.Println("encode client_config success")
}
