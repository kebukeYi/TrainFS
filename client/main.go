package main

import (
	"fmt"
	"github.com/kebukeYi/TrainFS/client/cli"
)

func main() {
	client := cli.NewClient()
	localFilePathPut := "/usr/golanddata/trainfs/client/put/912KB.data"
	localPathGet := "/usr/golanddata/trainfs/client/get"

	//remotePath := "/user/app"
	remotePath := "/user/local"
	remoteFilePath, err := client.PutFile(localFilePathPut, remotePath)
	if err != nil {
		fmt.Printf("client.PutFile(%s, %s); err:%s", localFilePathPut, remotePath, err)
		return
	}
	fmt.Println("===============================================================")
	file, err2 := client.GetFile(localPathGet, remoteFilePath)
	if err2 != nil {
		fmt.Printf("client.GetFile(%s); err2:%v ;\n", remotePath, err2)
	} else {
		fmt.Printf(file.Name())
	}

	dir, err := client.ListDir("/")
	if err != nil {
		fmt.Printf("client.ListDir(/); err:%s", err)
	} else {
		fmt.Printf("client.ListDir(/) %v; err:%s", dir, err)
	}

	dir, err = client.ListDir("/user")
	if err != nil {
		fmt.Printf("client.ListDir(/user); err:%s", err)
	} else {
		fmt.Printf("client.ListDir(/user) %v; err:%s", dir, err)
	}

	err = client.DeleteFile(remoteFilePath)
	if err != nil {
		fmt.Printf("client.DeleteFile(%s); err:%s", remoteFilePath, err)
	}

	dir, err = client.ListDir("/")
	if err != nil {
		fmt.Printf("client.ListDir(/); err:%s", err)
	} else {
		fmt.Printf("client.ListDir(/) %v; err:%s", dir, err)
	}

	dir, err = client.ListDir("/user")
	if err != nil {
		fmt.Printf("client.ListDir(/user); err:%s", err)
	} else {
		fmt.Printf("client.ListDir(/user) %v; err:%s", dir, err)
	}

	_, err = client.ReName("/user", "/users")
	if err != nil {
		fmt.Printf("client.ReName(/user); err:%s", err)
	}

	dir, err = client.ListDir("/")
	if err != nil {
		fmt.Printf("client.ListDir(/user); err:%s", err)
	} else {
		fmt.Printf("client.ListDir(/user) %v; err:%s", dir, err)
	}
}
