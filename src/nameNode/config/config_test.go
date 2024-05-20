package config

import (
	"fmt"
	"testing"
)

func TestGetNameNodeConfig(t *testing.T) {
	conf := GetDataNodeConfig()
	fmt.Println(conf)
}
