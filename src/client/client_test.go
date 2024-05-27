package main

import (
	"fmt"
	"testing"
	"time"
	"trainfs/src/client/cli"
	"trainfs/src/common"
)

func TestGetFileOfChunkName(t *testing.T) {
	fileName := "/user/app/example.txt"
	chunkNum := 3

	expectedResult := []string{"/user/app/example.txt_chunk_0", "/user/app/example.txt_chunk_1", "/user/app/example.txt_chunk_2"}

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
	remotePath2 := "/root/local/b2"
	remotePath3 := "/root/mbn/c3"
	remotePaths := make([]string, 0)
	remotePaths = append(remotePaths, remotePath1, remotePath2, remotePath3)
	for _, path := range remotePaths {
		err := client.Mkdir(path)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	fmt.Printf("=======================================================")
	time.Sleep(time.Second * 2)
	remotePath4 := "/root/app"
	remotePath5 := "/root"
	remotePath6 := "/"
	remotePaths = make([]string, 0)
	remotePaths = append(remotePaths, remotePath4, remotePath5, remotePath6)
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

	localPath := "F:\\y.jpg" // 912KB / 400 = 3块
	remotePath1 := "/root/app/y.data"
	remotePath2 := "/root/local/y.data"
	remotePath3 := "/root/mbn/y.data"
	remotePaths = append(remotePaths, remotePath1, remotePath2, remotePath3)
	for _, path := range remotePaths {
		client.PutFile(localPath, path)
	}
	fmt.Printf("=======================================================\n")
	time.Sleep(time.Second * 2)
	remotePath11 := "/root/app"
	remotePath22 := "/root/local"
	remotePath33 := "/root/mbn"
	remotePaths = make([]string, 0)
	remotePaths = append(remotePaths, remotePath11, remotePath22, remotePath33)
	for _, path := range remotePaths {
		dirMetaList, err := client.ListDir(path)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(dirMetaList.String())
	}
}

func TestGetFile(t *testing.T) {
	client := cli.NewClient()
	localPath1 := "F:\\getRootApp_y.jpg"
	localPath2 := "F:\\getRootLocal_y.jpg"
	localPath3 := "F:\\getRootMbn_y.jpg"
	localPaths := make([]string, 0)
	localPaths = append(localPaths, localPath1, localPath2, localPath3)
	remotePath1 := "/root/app/y.data"
	remotePath2 := "/root/local/y.data"
	remotePath3 := "/root/mbn/y.data"
	remotePaths := make([]string, 0)
	remotePaths = append(remotePaths, remotePath1, remotePath2, remotePath3)
	for i, path := range remotePaths {
		file, err := client.GetFile(localPaths[i], path)
		if err != nil {
			fmt.Printf("getFile(%s) ,err:%s \n", path, err)
			continue
		}
		fmt.Println(file.Name())
	}
}

func TestDelete(t *testing.T) {
	client := cli.NewClient()
	remotePaths := make([]string, 0)
	//localPath := "F:\\y.jpg" // 912KB / 400 = 3块
	//remotePath0 := "/root/app/y.data"
	//client.PutFile(localPath, remotePath0)
	fmt.Printf("=======================================================\n")
	time.Sleep(time.Second * 3)

	//remotePath0 := "/root/app"
	remotePath1 := "/root/app/y.data"
	//remotePath2 := "/root/local/y.data"
	//remotePath3 := "/root/mbn"        // 测试删除目录
	//remotePath5 := "/not_root"        // 测试删除不存在的目录
	//remotePaths = append(remotePaths, remotePath1, remotePath2, remotePath3, remotePath5)
	remotePaths = append(remotePaths, remotePath1)
	for _, path := range remotePaths {
		err := client.DeleteFile(path)
		if err != nil {
			fmt.Printf("client.DeleteFile err:%v \n", err)
		}
	}
	fmt.Printf("=======================================================\n")
	time.Sleep(time.Second * 2)
	dirMetaList, err := client.ListDir("/root/app")
	if err != nil {
		fmt.Printf("client.ListDir err:%v \n", err)
		return
	}
	fmt.Printf("dirMetaList: %v \n", dirMetaList.String())
	fmt.Printf("=======================================================\n")
	time.Sleep(time.Second * 2)
	localPath4 := "F:\\getDelRootApp_y.jpg"
	remotePath4 := "/root/app/y.data"
	getFile, err := client.GetFile(localPath4, remotePath4)
	if err != nil {
		fmt.Printf("client.GetFile err:%v \n", err)
		return
	}
	fmt.Println(getFile.Name())
}

func TestList(t *testing.T) {
	client := cli.NewClient()
	//remotePath := "/user"
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
