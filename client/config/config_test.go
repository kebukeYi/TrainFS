package config

import (
	"fmt"
	"testing"
)

func TestGetClientConfig(t *testing.T) {
	clientConfig := GetClientConfig()
	fmt.Printf("clientConfig: %v\n", clientConfig)
}
