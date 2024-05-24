package main

import (
	"fmt"
	"time"
	"trainfs/src/client/cli"
)

func main() {
	client := cli.NewClient()
	// fmt.Println(client)
	//localPath := "F:\\testGPU.py"
	localPathGet := "F:\\getTestGPU.py"
	remotePath := "/user/testGPU.py"
	//client.PutFile(localPath, remotePath)
	fmt.Println("===============================================================")

	time.Sleep(5 * time.Second)
	file, err2 := client.GetFile(localPathGet, remotePath)
	if err2 != nil {
		fmt.Printf("client.GetFile(%s); err2:%v ;\n", remotePath, err2)
	} else {
		fmt.Printf(file.Name())
	}

	//dir, err := client.ListDir("/")
	//if err != nil {
	//	fmt.Printf("client.ListDir(/); err:%s", err)
	//} else {
	//	fmt.Printf("client.ListDir(/) %v; err:%s", dir, err)
	//}
	//
	//dir, err = client.ListDir("/user")
	//if err != nil {
	//	fmt.Printf("client.ListDir(/user); err:%s", err)
	//} else {
	//	fmt.Printf("client.ListDir(/user) %v; err:%s", dir, err)
	//}
	//
	//err = client.DeleteFile(remotePath)
	//if err != nil {
	//	fmt.Printf("client.DeleteFile(); err:%s", err)
	//}
	//
	//dir, err = client.ListDir("/")
	//if err != nil {
	//	fmt.Printf("client.ListDir(/); err:%s", err)
	//} else {
	//	fmt.Printf("client.ListDir(/) %v; err:%s", dir, err)
	//}
	//
	//dir, err = client.ListDir("/user")
	//if err != nil {
	//	fmt.Printf("client.ListDir(/user); err:%s", err)
	//} else {
	//	fmt.Printf("client.ListDir(/user) %v; err:%s", dir, err)
	//}
	//
	//_, err = client.ReName("/user", "/users")
	//if err != nil {
	//	fmt.Printf("client.ReName(/user); err:%s", err)
	//}
	//
	//dir, err = client.ListDir("/")
	//if err != nil {
	//	fmt.Printf("client.ListDir(/user); err:%s", err)
	//} else {
	//	fmt.Printf("client.ListDir(/user) %v; err:%s", dir, err)
	//}
}
