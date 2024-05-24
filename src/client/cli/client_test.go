package cli

import (
	"testing"
)

func TestGetFileOfChunkName(t *testing.T) {
	fileName := "/user/app/example.txt"
	chunkNum := 3

	expectedResult := []string{"/user/app/example.txt_chunk_0", "/user/app/example.txt_chunk_1", "/user/app/example.txt_chunk_2"}

	result := GetFileOfChunkName(fileName, chunkNum)

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
	client := NewClient()
	localPath := "F:\\testGPU.py"
	remotePath := "/user/testGPU.py"
	client.PutFile(localPath, remotePath)
}
