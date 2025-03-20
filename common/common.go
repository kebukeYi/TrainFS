package common

import (
	"errors"
	"fmt"
	"github.com/kebukeYi/TrainDB/common"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var ErrFileCorruption = errors.New("client: file Corruption")
var ErrInputEmpty = errors.New("input not empty")
var ErrHeartBeatNotExist = errors.New("hearBeat not exist, first need to reRegister")

var ErrCanNotChangeRootDir = errors.New("NameNode: can not change root dir")
var ErrCanNotDelNotEmptyDir = errors.New("NameNode: A non-empty folder name change is not allowed")
var ErrPathFormatError = errors.New("NameNode: path format error")
var ErrFileFormatError = errors.New("NameNode: file format error")
var ErrFileNotFound = errors.New("NameNode: file not found")
var ErrNotDir = errors.New("NameNode: the files is not directory")
var ErrFileAlreadyExists = errors.New("NameNode: the file already exists")
var ErrDirAlreadyExists = errors.New("NameNode: the dir already exists")
var ErrCommitChunkType = errors.New("NameNode: CommitChunk not found")
var ErrChunkReplicaNotFound = errors.New("NameNode: chunk replica not found")

var ErrEnoughReplicaDataNodeServer = errors.New("DataNode: not enough replica num dataNodeServer")
var ErrEnoughUpDataNodeServer = errors.New("DataNode: not enough up num dataNodeServer")

var ErrNotSupported = errors.New("NameNode: not supported now ")

// GetFileChunkNameOfNum fileName: /root/app/test.txt; chunkNum:3;
// return: [ /root/app/test.txt_chunk_0, /root/app/test.txt_chunk_1, /root/app/test.txt_chunk_2]
func GetFileChunkNameOfNum(fileName string, chunkNum int) []string {
	result := make([]string, 0)
	for i := 0; i < chunkNum; i++ {
		result = append(result, fileName+"_chunk_"+strconv.Itoa(i))
	}
	return result
}

func GetChunkIdOfFileChunkName(fileChunkName string) int32 {
	if fileChunkName == "" {
		return -1
	}
	indexAny := strings.LastIndexAny(fileChunkName, "_chunk_")
	if indexAny <= 0 {
		return -1
	}
	if indexAny+1 >= len(fileChunkName) {
		return -1
	}
	s := fileChunkName[indexAny+1:]
	chunkID, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("GetChunkIdOfFileChunkName(%s) err: %v", fileChunkName, err)
		return -1
	}
	return int32(chunkID)
}

// GetFileNameFromChunkName filePathChunkName: /root/app/test.txt_chunk_0  /roo_t/app/t_est.txt_chunk_0
// return: /root/app/test.txt
func GetFileNameFromChunkName(fileChunkName string) string {
	if fileChunkName == "" {
		return ""
	}
	indexAny := strings.LastIndexAny(fileChunkName, "_chunk_")
	if indexAny <= 0 {
		return ""
	}
	if indexAny-6 <= 0 {
		return ""
	}
	return fileChunkName[:indexAny-6]
}
func SplitClientFileNamePath(fileNamePath string) (fileName, path string) {
	index := strings.LastIndex(fileNamePath, string(filepath.Separator))
	if index < 0 {
		return "", ""
	}
	path = fileNamePath[:index]
	if path == "" {
		path = string(filepath.Separator)
		fileName = fileNamePath[index+1:]
		return
	}
	fileName = fileNamePath[index+1:]
	return
}

func SplitFileNamePath(fileNamePath string) (fileName, path string) {
	index := strings.LastIndex(fileNamePath, "/")
	if index < 0 {
		return "", ""
	}
	path = fileNamePath[:index]
	if path == "" {
		path = "/"
		fileName = fileNamePath[index+1:]
		return
	}
	fileName = fileNamePath[index+1:]
	return
}

func ClearDir(dir string) {
	_, err := os.Stat(dir)
	if err == nil {
		if err = os.RemoveAll(dir); err != nil {
			common.Panic(err)
		}
	}
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		_ = fmt.Sprintf("create dir %s failed", dir)
	}
}
