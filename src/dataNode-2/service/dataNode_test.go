package service

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/disk"
	"log"
	"testing"
)

const (
	BytesPerGB = 1024 * 1024 * 1024
	BytesPerMB = 1024 * 1024
)

func TestDiskSpace(t *testing.T) {
	//path := "F://" // 根目录为例，你可以替换为任何你想要检查的路径
	path := "F:\\TrainFS\\DataNode\\DataNode1\\data" // 根目录为例，你可以替换为任何你想要检查的路径
	freeSpace, err := disk.Usage(path)
	if err != nil {
		log.Fatal(err)
	}
	freeSpace.Free = freeSpace.Free / BytesPerMB
	fmt.Printf("剩余磁盘空间: %v MB\n", freeSpace.Free)
	fmt.Printf("使用磁盘空间: %v \n", freeSpace.Used)
}
