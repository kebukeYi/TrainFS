package config

import (
	"fmt"
	"testing"
)

func TestGetNameNodeConfig(t *testing.T) {
	// fileName := "nameNode_config.yml"
	conf := GetDataNodeConfig()
	fmt.Println(conf)
}
