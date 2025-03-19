package main

import (
	"fmt"
	"github.com/kebukeYi/TrainFS/client/cli"
	"github.com/kebukeYi/TrainFS/common"
	"testing"
	"time"
)

func TestGetFileOfChunkName(t *testing.T) {
	fileName := "/user/app/example.txt"
	chunkNum := 3

	expectedResult := []string{"/user/app/example.txt_chunk_0",
		"/user/app/example.txt_chunk_1",
		"/user/app/example.txt_chunk_2"}

	result := common.GetFileChunkNameOfNum(fileName, chunkNum)

	if len(result) != len(expectedResult) {
		t.Errorf("Expected length of result to be %d, but got %d", len(expectedResult), len(result))
	}

	for i, value := range result {
		if value != expectedResult[i] {
			t.Errorf("Expected value at index %d to be %s, but got %s", i, expectedResult[i], value)
		}
	}
}

func TestMkdir(t *testing.T) {
	client := cli.NewClient()
	remotePath1 := "/root/app/a1"
	remotePath2 := "/root/app/a1/b2"
	remotePath3 := "/root/app/a1/b2/c3"
	remotePath4 := "/root/app/d4"
	remotePaths := make([]string, 0)
	remotePaths = append(remotePaths, remotePath1, remotePath2, remotePath3, remotePath4)
	for _, path := range remotePaths {
		err := client.Mkdir(path)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	fmt.Printf("==================ListDir=====================================")
	time.Sleep(time.Second * 2)
	remotePath11 := "/root/app"
	remotePath5 := "/root/app/a1"
	remotePath6 := "/root/app/a1/b2"
	remotePaths = remotePaths[:0]
	remotePaths = append(remotePaths, remotePath11, remotePath5, remotePath6)
	for _, path := range remotePaths {
		reply, err := client.ListDir(path)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(reply.String())
	}
}

func TestPutFile(t *testing.T) {
	client := cli.NewClient()
	remotePaths := make([]string, 0)

	// 23_2MB.jpg  810KB.png   200KB.png   1440KB.jpg
	// /usr/projects_gen_data/goprogendata/trainfsdata/test/client/get
	// /usr/projects_gen_data/goprogendata/trainfsdata/test/client/put
	localFilePath := "/usr/projects_gen_data/goprogendata/trainfsdata/test/client/put/810KB.png" // 912KB / 400 = 3块

	remotePath1 := "/root/app"
	//remotePath2 := "/root/local"
	//remotePath3 := "/root/mbn"
	//remotePaths = append(remotePaths, remotePath1, remotePath2, remotePath3)
	remotePaths = append(remotePaths, remotePath1)
	for _, path := range remotePaths {
		client.PutFile(localFilePath, path)
	}
	fmt.Printf("=======================================================\n")
	time.Sleep(time.Second * 3)
	for _, path := range remotePaths {
		dirMetaList, err := client.ListDir(path)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(dirMetaList.String())
	}
	fmt.Printf("=======================================================\n")
	time.Sleep(time.Second * 3)
	localPath2 := "/usr/projects_gen_data/goprogendata/trainfsdata/test/client/get2"
	remoteFilePath2 := "/root/app/810KB.png"
	file, err := client.GetFile(localPath2, remoteFilePath2)
	if err != nil {
		fmt.Printf("getFile(%s) ,err:%s \n", remoteFilePath2, err)
	}
	fmt.Println(file.Name())
	file.Close()
}

func TestGetFile(t *testing.T) {
	client := cli.NewClient()
	localPath1 := "/usr/projects_gen_data/goprogendata/trainfsdata/test/client/get1"
	localPath2 := "/usr/projects_gen_data/goprogendata/trainfsdata/test/client/get2"
	localPath3 := "/usr/projects_gen_data/goprogendata/trainfsdata/test/client/get3"
	localPaths := make([]string, 0)
	localPaths = append(localPaths, localPath1, localPath2, localPath3)
	remoteFilePath1 := "/root/app/810KB.png"
	//remoteFilePath2 := "/root/local/810KB.png"
	//remoteFilePath3 := "/root/mbn/810KB.png"
	remoteFilePaths := make([]string, 0)
	//remoteFilePaths = append(remoteFilePaths, remoteFilePath1, remoteFilePath2, remoteFilePath3)
	remoteFilePaths = append(remoteFilePaths, remoteFilePath1)
	for i, path := range remoteFilePaths {
		file, err := client.GetFile(localPaths[i], path)
		if err != nil {
			fmt.Printf("getFile(%s) ,err:%s \n", path, err)
			continue
		}
		fmt.Println(file.Name())
		file.Close()
	}
}

func TestDelete(t *testing.T) {
	client := cli.NewClient()
	remoteFilePaths := make([]string, 0)
	remoteFilePath1 := "/root/app/810KB.png"
	//remotePath2 := "/root/local/y.data"
	//remotePath3 := "/root/mbn"        // 测试删除目录
	//remotePath5 := "/not_root"        // 测试删除不存在的目录
	//remotePaths = append(remotePaths, remotePath1, remotePath2, remotePath3, remotePath5)
	remoteFilePaths = append(remoteFilePaths, remoteFilePath1)
	for _, path := range remoteFilePaths {
		err := client.DeleteFile(path)
		if err != nil {
			fmt.Printf("client.DeleteFile err:%v \n", err)
		}
	}
	fmt.Printf("=======================================================\n")
	time.Sleep(time.Second * 3)
	dirMetaList, err := client.ListDir("/root/app")
	if err != nil {
		fmt.Printf("client.ListDir err:%v \n", err)
		return
	}
	fmt.Printf("dirMetaList: %v \n", dirMetaList.String())
	fmt.Printf("=======================================================\n")
	time.Sleep(time.Second * 2)
	localPath4 := "/usr/projects_gen_data/goprogendata/trainfsdata/test/client/get4"
	remoteFilePath4 := "/root/app/810KB.png"
	getFile, err := client.GetFile(localPath4, remoteFilePath4)
	if err != nil {
		fmt.Printf("client.GetFile(%s) err:%v \n", remoteFilePath4, err)
		return
	}
	fmt.Println(getFile.Name())
	getFile.Close()
}

func TestList(t *testing.T) {
	client := cli.NewClient()
	remotePath := "/"
	listDir, err := client.ListDir(remotePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf(listDir.String())
}

func TestReName(t *testing.T) {
	client := cli.NewClient()
	oldPath := "/user"
	newPath := "/users"
	listDir, err := client.ReName(oldPath, newPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf(listDir.String())
}
