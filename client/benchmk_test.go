package main

import (
	"fmt"
	"github.com/kebukeYi/TrainFS/client/cli"
	"github.com/stretchr/testify/require"
	"math/rand"
	"strconv"
	"testing"
)

func BenchmarkPutFile(b *testing.B) {
	// go test -bench=BenchmarkNormalEntry -benchtime=3s -count=2 -failfast
	// go test -bench=BenchmarkNormalEntry -benchtime=100000x -count=5 -failfast
	client := cli.NewClient()
	clientPutPath := "F:\\ProjectsData\\golang\\TrainFS\\client\\put\\"
	remotePath := "/root/app"
	filePathName := clientPutPath + "222.data"
	//fileSize := 1 << 20 // 1MB      chunkSize: 400KB
	//fileSize := 15 << 20 // 15MB    chunkSize: 1M
	fileSize := 65 << 20 // 65MB      chunkSize: 35MB
	//fileSize := 885 << 20 // 885MB  chunkSize: 64MB
	err := cli.TruncateFile(filePathName, int64(fileSize))
	if err != nil {
		fmt.Printf("TruncateFile error: %v", err)
		return
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		randPath := rand.Int()
		newRemotePath := remotePath + strconv.Itoa(randPath)
		_, err = client.PutFile(filePathName, newRemotePath)
		require.Nil(b, err)
	}
}
