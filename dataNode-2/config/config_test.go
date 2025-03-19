package config

import (
	"fmt"
	"testing"
)

func TestGetDataNodeConfig(t *testing.T) {
	conf := GetDataNodeConfig()
	fmt.Println(conf)
}
