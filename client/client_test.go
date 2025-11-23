package main

import (
	"errors"
	"fmt"
	"github.com/kebukeYi/TrainFS/client/cli"
	"github.com/kebukeYi/TrainFS/common"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// parseSize 解析类似 "1GB", "512MB" 的字符串为字节数
func parseSize(sizeStr string) (int64, error) {
	re := regexp.MustCompile(`^(\d+)([KMGT]B)?$`)
	parts := re.FindStringSubmatch(strings.ToUpper(sizeStr))
	if parts == nil {
		return 0, errors.New("invalid size format")
	}

	num, _ := strconv.ParseInt(parts[1], 10, 64)
	unit := parts[2]

	switch unit {
	case "TB":
		return num * 1024 * 1024 * 1024 * 1024, nil
	case "GB":
		return num * 1024 * 1024 * 1024, nil
	case "MB":
		return num * 1024 * 1024, nil
	case "KB":
		return num * 1024, nil
	default:
		return num, nil // 无单位时视为字节
	}
}

// createFile 快速生成指定大小的文件;
func CreateEmptyFile(filename string, sizeStr string) error {
	size, err := parseSize(sizeStr)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 使用 1MB 缓冲区提高写入效率
	buffer := make([]byte, 1024*1024) // 1MB 块
	for written := int64(0); written < size; {
		chunkSize := size - written
		if chunkSize > int64(len(buffer)) {
			chunkSize = int64(len(buffer))
		}
		n, err := file.Write(buffer[:chunkSize])
		if err != nil {
			return err
		}
		written += int64(n)
	}
	return nil
}

func TestTruncateFile(t *testing.T) {
	localFilePath := "/usr/golanddata/trainfs/client/put/111.data" // 65MB / 400 = 3块
	err := cli.TruncateFile(localFilePath, 1024*1024)
	if err != nil {
		return
	}
}

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

func TestMkdirInNameNode(t *testing.T) {
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
	fmt.Println("==================ListDir=====================================")
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

func TestReName(t *testing.T) {
	client := cli.NewClient()
	oldPath := "/root/app"
	newPath := "/root/apps"
	//nameReply, err := client.ReName(oldPath, newPath)
	//if err != nil {
	//	fmt.Println(err)
	//} else {
	//	fmt.Println(nameReply.String())
	//}
	fmt.Println("===================================================")
	oldPath = "/root/aop"
	newPath = "/root/newAop"
	if err := client.Mkdir(oldPath); err != nil {
		fmt.Println(err)
		return
	}
	nameReply, err := client.ReName(oldPath, newPath)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(nameReply.String())
	}

	if dirMetaList, err := client.ListDir("/root"); err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(dirMetaList.String())
	}
}

func TestPutFile(t *testing.T) {
	client := cli.NewClient()
	remotePaths := make([]string, 0)

	// 23_2MB.jpg  810KB.png   200KB.png   1440KB.jpg
	// /usr/projects_gen_data/goprogendata/trainfsdata/test/client/get
	// /usr/projects_gen_data/goprogendata/trainfsdata/test/client/put

	// linux
	//localFilePath := "/usr/golanddata/trainfs/client/put/912KB.data" // 912KB / 400 = 3块

	// windows
	localFilePath := "F:\\ProjectsData\\golang\\TrainFS\\client\\put\\y.jpg"

	// nameNode`s remotePath format linux
	remotePath1 := "/root/app"
	//remotePath2 := "/root/local"
	//remotePath3 := "/root/mbn"
	//remotePaths = append(remotePaths, remotePath1, remotePath2, remotePath3)
	remotePaths = append(remotePaths, remotePath1)
	for _, path := range remotePaths {
		_, err := client.PutFile(localFilePath, path)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	fmt.Printf("=======================================================\n")
	// time.Sleep(time.Second * 2)
	for _, path := range remotePaths {
		dirMetaList, err := client.ListDir(path)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("listDir(%s):%s \n", path, dirMetaList.String())
	}
	fmt.Printf("=======================================================\n")
	// time.Sleep(time.Second * 1)
	// linux client get path
	//localPath2 := "/usr/golanddata/trainfs/client/get2"

	// windows client get path
	localPath2 := "F:\\ProjectsData\\golang\\TrainFS\\client\\get2"

	// nameNode`s remotePath format linux
	//remoteFilePath2 := "/root/app/810KB.png"
	//remoteFilePath2 := "/root/app/222.data"
	remoteFilePath2 := "/root/app/y.jpg"
	file, err := client.GetFile(localPath2, remoteFilePath2)
	if err != nil {
		fmt.Printf("getFile(%s) ,err:%s \n", remoteFilePath2, err)
		return
	}
	fmt.Println(file.Name())
	file.Close()
}

func TestGetFile(t *testing.T) {
	client := cli.NewClient()
	// linux client get path
	//localPath1 := "/usr/golanddata/trainfs/client/get1"

	// windows client get path
	localPath1 := "F:\\ProjectsData\\golang\\TrainFS\\client\\get1"

	// nameNode`s remotePath format linux
	//remoteFilePath1 := "/root/app/810KB.png"
	//remoteFilePath1 := "/root/app/y.jpg"
	remoteFilePath1 := "/root/app8282524010649360361/222.data"
	file, err := client.GetFile(localPath1, remoteFilePath1)
	if err != nil {
		fmt.Printf("getFile(%s) ,err:%s \n", remoteFilePath1, err)
		return
	}
	fmt.Println(file.Name())
	file.Close()
}

func TestDelete(t *testing.T) {
	client := cli.NewClient()
	remoteFilePaths := make([]string, 0)
	// nameNode`s remotePath format linux
	remoteFilePath1 := "/root/app/y.jpg"
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
	time.Sleep(time.Second * 2)
	dirMetaList, err := client.ListDir("/root/app")
	if err != nil {
		fmt.Printf("client.ListDir err:%v \n", err)
		return
	}
	fmt.Printf("dirMetaList: %v \n", dirMetaList.String())
	fmt.Printf("=======================================================\n")
	time.Sleep(time.Second * 2)

	// windows client get path
	//localPath4 := "/usr/golanddata/trainfs/client/get4"

	// linux client get path
	localPath4 := "F:\\ProjectsData\\golang\\TrainFS\\client\\get4"

	// nameNode`s remotePath format linux
	//remoteFilePath4 := "/root/app/810KB.png"
	remoteFilePath4 := "/root/app/y.jpg"
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
	remotePath := "/root"
	listDir, err := client.ListDir(remotePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf(listDir.String())
}
