package config

import (
	"fmt"
	"testing"
)

func TestGetClientConfig(t *testing.T) {
	clientConfig := GetClientConfig()
	fmt.Println(clientConfig)
}
