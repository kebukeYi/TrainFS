package main

import (
	"fmt"
	"testing"
	"trainfs/src/client/cli"
)

func TestGetFileOfChunkName(t *testing.T) {
	fileName := "/user/app/example.txt"
	chunkNum := 3

	expectedResult := []string{"/user/app/example.txt_chunk_0", "/user/app/example.txt_chunk_1", "/user/app/example.txt_chunk_2"}

	result := cli.GetFileOfChunkName(fileName, chunkNum)

	if len(result) != len(expectedResult) {
		t.Errorf("Expected length of result to be %d, but got %d", len(expectedResult), len(result))
	}

	for i, value := range result {
		if value != expectedResult[i] {
			t.Errorf("Expected value at index %d to be %s, but got %s", i, expectedResult[i], value)
		}
	}
}

func TestPutFile(t *testing.T) {
	client := cli.NewClient()
	fmt.Println("client:", client)
	//localPath := "F:\\yyyyy.jpg"
	//remotePath := "/user/apps/yyyyy.data"
	//client.PutFile(localPath, remotePath)
}
