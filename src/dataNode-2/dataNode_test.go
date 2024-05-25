package main

import (
	"fmt"
	"testing"
	"trainfs/src/dataNode-2/service"
)

func TestGetConfig(t *testing.T) {
	dataNode := service.NewDataNode()
	fmt.Println(dataNode)
}
