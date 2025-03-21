package service

import (
	"fmt"
	proto "github.com/kebukeYi/TrainFS/profile"
	"github.com/shirou/gopsutil/v3/disk"
	"log"
	"testing"
)

const (
	BytesPerGB = 1024 * 1024 * 1024
	BytesPerMB = 1024 * 1024
)

func TestDiskSpace(t *testing.T) {
	path := "/usr"
	freeSpace, err := disk.Usage(path)
	if err != nil {
		log.Fatal(err)
	}
	freeSpace.Free = freeSpace.Free / BytesPerMB
	fmt.Printf("剩余磁盘空间: %v MB\n", freeSpace.Free)
	fmt.Printf("使用磁盘空间: %v \n", freeSpace.Used)
}

func TestStoreManger_PutChunkInfos(t *testing.T) {
	buf := []byte("hello world")
	chunkInfo := &proto.ChunkInfo{
		ChunkId:           1,
		ChunkSize:         int64(len(buf)),
		FilePathName:      "filePathName",
		FilePathChunkName: "filePathChunkName",
		DataNodeAddress:   &proto.DataNodeChain{DataNodeAddress: []string{"dataNode.Config.Host"}},
	}
	allChunkInfos := make(map[string]*proto.ChunkInfo)
	allChunkInfos["filePathChunkName"] = chunkInfo
	if infos2bytes, err := chunkInfos2bytes(allChunkInfos); err != nil {
		t.Error(err)
	} else {
		fmt.Printf("infos2bytes: len:%d \n", len(infos2bytes))
		storeManager := OpenStoreManager("")
		defer storeManager.Close()
		if err = storeManager.PutChunkInfos("infos2bytes", allChunkInfos); err != nil {
			t.Error(err)
		}
		if infos, err := storeManager.GetChunkInfos("infos2bytes"); err != nil {
			t.Error(err)
		} else {
			fmt.Printf("infos: %v \n", infos)
		}
	}

}
