package common

import (
	"errors"
	"strconv"
	"strings"
)

var ErrFileCorruption = errors.New("client: file Corruption")

var ErrCanNotChangeRootDir = errors.New("NameNode: can not change root dir")
var ErrCanNotDelNotEmptyDir = errors.New("NameNode: can not del not empty dir")
var ErrPathFormatError = errors.New("NameNode: path format error")
var ErrFileFormatError = errors.New("NameNode: file format error")
var ErrFileNotFound = errors.New("NameNode: file not found")
var ErrNotDir = errors.New("NameNode: the files is not directory")
var ErrFileAlreadyExists = errors.New("NameNode: the file already exists")
var ErrDirAlreadyExists = errors.New("NameNode: the dir already exists")
var ErrCommitChunkType = errors.New("NameNode: CommitChunk not found")
var ErrChunkReplicaNotFound = errors.New("NameNode: chunk replica not found")

var ErrNotEnoughStorageSpace = errors.New("DataNode: not enough storage space ")
var ErrEnoughReplicaDataNodeServer = errors.New("DataNode: not enough replica num dataNodeServer")
var ErrEnoughUpDataNodeServer = errors.New("DataNode: not enough up num dataNodeServer")

var ErrNotSupported = errors.New("NameNode: not supported now ")

// GetFileChunkNameOfNum fileName: /root/app/test.txt chunkNum: 3
// return: [ /root/app/test.txt_chunk_0, /root/app/test.txt_chunk_1, /root/app/test.txt_chunk_2]
func GetFileChunkNameOfNum(fileName string, chunkNum int) []string {
	result := make([]string, 0)
	for i := 0; i < chunkNum; i++ {
		result = append(result, fileName+"_chunk_"+strconv.Itoa(i))
	}
	return result
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
