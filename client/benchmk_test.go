package main

import (
	"fmt"
	"github.com/kebukeYi/TrainFS/client/cli"
	"github.com/stretchr/testify/require"
	"math/rand"
	"strconv"
	"testing"
)

var clientPutPath = "/usr/golanddata/trainfs/client/put"

func BenchmarkPutFile(b *testing.B) {
	// go test -bench=BenchmarkNormalEntry -benchtime=3s -count=2 -failfast
	// go test -bench=BenchmarkNormalEntry -benchtime=100000x -count=5 -failfast
	client := cli.NewClient()
	remotePath := "/root/app"
	filePathName := clientPutPath + "992KB.data" // 992KB 大小
	//truncateSize := 1 << 20 // 1MB      chunkSize: 400KB
	//truncateSize := 15 << 20 // 15MB    chunkSize: 1M
	truncateSize := 65 << 20 // 65MB      chunkSize: 35MB
	//truncateSize := 885 << 20 // 885MB  chunkSize: 64MB
	err := cli.TruncateFile(filePathName, int64(truncateSize))
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
